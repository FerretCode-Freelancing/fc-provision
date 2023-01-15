package builder

import (
	"archive/zip"
	"bytes"
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
	Owner  string // repo owner
	Url    string // url to fc-session-cache instance to get tokens from
	Cookie string // session cookie value to fetch token from
}

type AuthResponse struct {
	Session Sesssion `json:"session"`
}

type Sesssion struct {
	Token string `json:"access_token"`
}

func (e *Extractor) Auth() (success bool, gc github.Client, ct context.Context, token string, err error) {
	ctx := context.Background()

	client := &http.Client{}

	if err != nil {
		return false, github.Client{}, ctx, "", err
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/get?sid=%s", e.Url, e.Cookie[4:36]),
		nil,
	)

	if err != nil {
		return false, github.Client{}, ctx, "", err
	}

	res, err := client.Do(req)

	if err != nil {
		return false, github.Client{}, ctx, "", err
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		return false, github.Client{}, ctx, "", err
	}

	authResponse := AuthResponse{}

	if jsonErr := json.Unmarshal(resBody, &authResponse); jsonErr != nil {
		return false, github.Client{}, ctx, "", jsonErr
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authResponse.Session.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(tc)

	return true, *ghClient, ctx, authResponse.Session.Token, nil
}

func (e *Extractor) DownloadRepo() (int64, string, error) {
	ok, client, ctx, token, err := e.Auth()

	if err != nil {
		return 0, "", err
	}

	if !ok {
		return 0, "", errors.New("there was an error authenticating the current user")
	}

	repoName := strings.Split(e.Repo, "/")

	repo, _, err := client.Repositories.Get(ctx, e.Owner, repoName[len(repoName) - 1])

	if err != nil {
		return 0, "", err
	}

	httpClient := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/zipball/%s", *repo.URL, repo.GetMasterBranch()),
		nil,
	) 

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	if err != nil {
		return 0, "", err
	}

	zipball, err := httpClient.Do(req)

	if err != nil {
		return 0, "", err
	}

	defer zipball.Body.Close()

	zipballBody, err := io.ReadAll(zipball.Body)

	if err != nil {
		return 0, "", err
	}

	file, err := os.Create(
		fmt.Sprintf(
			"/tmp/fc-builder/%s-%s.zip",
			strconv.FormatInt(*repo.Owner.ID, 10),
			*repo.Name,
		),
	)

	if err != nil {
		return 0, "", err
	}

	defer file.Close()

	_, downloadErr := io.Copy(file, bytes.NewReader(zipballBody))

	if downloadErr != nil {
		return 0, "", downloadErr
	}

	fmt.Printf(
		"Downloaded %s-%s.zip\n",
		strconv.FormatInt(*repo.Owner.ID, 10),
		*repo.Name,
	)

	return *repo.Owner.ID, *repo.Name, nil
}

func (e *Extractor) ExtractRepo(ownerId int64, repoName string) (string, error) {
	outputDir := fmt.Sprintf(
		"/tmp/fc-builder/out/%s-%s",
		strconv.FormatInt(ownerId, 10),
		repoName,
	)

	zipball, err := zip.OpenReader(
		fmt.Sprintf(
			"/tmp/fc-builder/%s-%s.zip",
			strconv.FormatInt(ownerId, 10),
			repoName,
		),
	)

	if err != nil {
		return "", err
	}

	defer zipball.Close()

	for _, file := range zipball.File {
		file.Name = filepath.Base(file.Name)

		path := filepath.Join(outputDir, file.Name)

		if !strings.HasPrefix(path, filepath.Clean(outputDir)+string(os.PathSeparator)) {
			return "", errors.New("invalid file path")
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)

			continue
		}

		if err := os.MkdirAll(filepath.Dir(filepath.Dir(path)), os.ModePerm); err != nil {
			return "", err
		}

		// create file in output directory
		destFile, err := os.Create(path)

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
