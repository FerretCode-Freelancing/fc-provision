package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/basicauth-go"
	"github.com/ferretcode-freelancing/fc-provision/builder"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Api struct{}

func NewApi() Api {
	api := &Api{}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)

	username := strings.Trim(os.Getenv("FC_BUILDER_USERNAME"), "\n")
	password := strings.Trim(os.Getenv("FC_BUILDER_PASSWORD"), "\n")

	if username != "" && password != "" {
		r.Use(basicauth.New("fc-hosting", map[string][]string{
			username: {password},
		}))
	}

	r.Post("/build", func(w http.ResponseWriter, r *http.Request) {
		err := api.Build(w, r)

		if err != nil {
			fmt.Println(err)
		}
	})

	http.ListenAndServe(":3000", r)

	return *api
}

// s is the struct to unmarshal the request body into
func (a *Api) ProcessBody(w http.ResponseWriter, r *http.Request, s interface{}) error {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	if jsonErr := json.Unmarshal(body, s); jsonErr != nil {
		return jsonErr
	}

	return nil
}

type BuildRequest struct {
	Repo   string `json:"repo_name"`
	Owner  string `json:"owner_name"`
	Url    string `json:"cache_url"`
	Cookie string `json:"session_id"`
}

func (a *Api) Build(w http.ResponseWriter, r *http.Request) error {
	deployErr := "There was an error deploying your repository! Please try again later."

	br := &BuildRequest{}
	err := a.ProcessBody(w, r, br)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	if br.Url == "" {
		username := strings.Trim(os.Getenv("FC_SESSION_CACHE_USERNAME"), "\n")
		password := strings.Trim(os.Getenv("FC_SESSION_CACHE_PASSWORD"), "\n")

		ip := os.Getenv("FC_SESSION_CACHE_SERVICE_HOST")
		port := os.Getenv("FC_SESSION_CACHE_SERVICE_PORT")

		if username == "" || password == "" || ip == "" || port == "" {
			http.Error(w, "Internal server error. Please try again later.", http.StatusInternalServerError)
			return errors.New("the cache URL is invalid")
		}

		br.Url = fmt.Sprintf("%s:%s@%s:%s", username, password, ip, port)
	}

	extractor := builder.Extractor{
		Repo:   br.Repo,
		Owner:  br.Owner,
		Url:    br.Url,
		Cookie: br.Cookie,
	}

	ownerId, downloadErr := extractor.DownloadRepo()

	if downloadErr != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return downloadErr
	}

	path, err := extractor.ExtractRepo(ownerId, extractor.Repo)

	if err != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return err
	}

	processor := builder.Processor{
		Path: path,
	}

	copyErr := processor.CopyDockerfile()

	if copyErr != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return err
	}

	buildErr := builder.Build(
		path,
		fmt.Sprintf("%s-%s", br.Owner, br.Repo),
	)

	if buildErr != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("The repository was deployed."))

	return nil
}
