package builder

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Extractor struct {
	Repo   string // repo to fetch from
	Owner string // repo owner
	Url    string // url to fc-session-cache instance to get tokens from
	Cookie string // session cookie value to fetch token from
}

type AuthResponse struct {
	Token string `json:"access_token"`
}

func (e *Extractor) Auth() (success bool, gc github.Client, ct context.Context, err error) {
	ctx := context.Background()

	client := &http.Client{}

	if err != nil {
		return false, github.Client{}, ctx, err
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/get?sid=%s", e.Url, e.Cookie),
		nil,
	)

	res, err := client.Do(req)
	
	if err != nil {
		return false, github.Client{}, ctx, err
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		return false, github.Client{}, ctx, err
	}

	authResponse := &AuthResponse{}

	if jsonErr := json.Unmarshal(resBody, &authResponse); jsonErr != nil {
		return false, github.Client{}, ctx, jsonErr
	}	

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authResponse.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(tc)

	return true, *ghClient, ctx, nil
}

func (e *Extractor) DownloadRepo() (int64, error) {
	ok, client, ctx, err := e.Auth()

	if err != nil {
		return 0, err
	}

	if ok != true {
		return 0, errors.New("There was an error authenticating the current user!")
	}
	
	repo, _, err := client.Repositories.Get(ctx, e.Owner, e.Repo)
	
	if err != nil {
		return 0, err
	}

	zipball, err := http.Get(*repo.ArchiveURL)

	if err != nil {
		return 0, err
	}

	defer zipball.Body.Close()

	file, err := os.Create(
		fmt.Sprintf(
			"/tmp/fc-builder/%s-%s", 
			strconv.FormatInt(*repo.Owner.ID, 10), 
			*repo.Name,
		),
	)

	if err != nil {
		return 0, err
	}

	defer file.Close()

	_, downloadErr := io.Copy(file, zipball.Body)

	if downloadErr != nil {
		return 0, downloadErr
	}

	fmt.Println(
		fmt.Sprintf(
			"Downloaded %s-%s", 
			strconv.FormatInt(*repo.Owner.ID, 10), 
			*repo.Name,
		),
	)

	e.ExtractRepo(*repo.Owner.ID, *repo.Name)

	return 0, nil
}

func (e *Extractor) ExtractRepo(ownerId int64, repoName string) (string, error) {
	outputDir := fmt.Sprintf(
		"/tmp/fc-builder/out/%s-%s",
		strconv.FormatInt(ownerId, 10),
		repoName,
	) 

	zipball, err := zip.OpenReader(outputDir)

	if err != nil {
		return "", err
	}

	defer zipball.Close()

	for _, file := range zipball.File {
		path := filepath.Join(outputDir, file.Name)

		if !strings.HasPrefix(path, filepath.Clean(outputDir) + string(os.PathSeparator)) {
			return "", errors.New("Invalid file path")
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)

			continue
		}

		if err := os.MkdirAll(filepath.Dir(filepath.Dir(path)), os.ModePerm); err != nil {
			return "", err
		}

		// create file in output directory
		destFile, err := os.OpenFile(path, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, file.Mode())

		if err != nil {
			return "", err
		}

		// open file in archive
		zipballFile, err := file.Open()

		if err != nil {
			return "", err
		}

		// copy file in archive to the empty destination file
		if _, err := io.Copy(destFile, zipballFile); err != nil {
			return "", err
		}

		destFile.Close()
		zipballFile.Close()
	}

	return outputDir, nil
}
