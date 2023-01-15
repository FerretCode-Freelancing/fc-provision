package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/basicauth-go"
	events "github.com/ferretcode-freelancing/fc-bus"
	"github.com/ferretcode-freelancing/fc-provision/builder"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/kubemq-io/kubemq-go"
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
	Cookie string `json:"cookie"`
	ProjectId string `json:"project_id"`
}

type Ports struct {
	ContainerPort int `json:"containerPort"`
	Name string `json:"name"`
}

type DeployRequest struct {
	Ports []Ports `json:"ports"`
	Env map[string]string `json:"env"`
	ImageName string `json:"image_name"`
	ProjectId string `json:"project_id"`
}

func (a *Api) Build(w http.ResponseWriter, r *http.Request) error {
	deployErr := "There was an error deploying your repository! Please try again later."

	ctx := context.Background()

	bus := events.Bus{
		Channel: "deploy-pipeline",
		ClientId: uuid.NewString(),
		Context: ctx,
		TransportType: kubemq.TransportTypeGRPC,
	}

	client, connectErr := bus.Connect()

	if connectErr != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return connectErr
	}

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

		br.Url = fmt.Sprintf("http://%s:%s@%s:%s", username, password, ip, port)
	}

	extractor := builder.Extractor{
		Repo:   br.Repo,
		Owner:  br.Owner,
		Url:    br.Url,
		Cookie: br.Cookie,
	}

	ownerId, repo, downloadErr := extractor.DownloadRepo()

	if downloadErr != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return downloadErr
	}

	path, err := extractor.ExtractRepo(ownerId, repo)

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

		fmt.Println(copyErr)

		return err
	}

	imageName := fmt.Sprintf("%s-%s", strings.ToLower(br.Owner), strings.ToLower(repo))

	res, err := UpdateProject(r, imageName, *br)

	if err != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return err
	}
		
	if res.StatusCode != 200 {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return err
	}

	buildErr := builder.Build(
		path,
		imageName,
	)

	if buildErr != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return err
	}

	// TODO: add ports processing
	// TODO: clean this up (function too long)

	deployRequest := DeployRequest{
		ImageName: imageName,
		ProjectId: br.ProjectId,
	}

	stringified, err := json.Marshal(deployRequest)

	if err != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return err
	}

	_, sendErr := client.Send(ctx, kubemq.NewQueueMessage().
		SetId(uuid.NewString()).
		SetChannel(bus.Channel).
		SetBody(stringified))

	if sendErr != nil {
		http.Error(w, deployErr, http.StatusInternalServerError)

		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("The repository was built successfully."))

	return nil
}

func UpdateProject(
	r *http.Request, 
	imageName string, 
	br BuildRequest, 
) (http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"http://%s:%s/api/projects/update?id=%s", 
			os.Getenv("FC_PROJECTS_SERVICE_HOST"), 
			os.Getenv("FC_PROJECTS_SERVICE_PORT"),
			br.ProjectId,
		),
		bytes.NewReader([]byte(fmt.Sprintf(`{ "imageName": "%s" }`, imageName))),
	)

	if err != nil {
		return http.Response{}, err
	}

	req.AddCookie(&http.Cookie{
		Name: "fc-hosting",
		Value: br.Cookie,
	})

	res, err := client.Do(req)

	if err != nil {
		return http.Response{}, err
	}

	return *res, nil
}
