package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Builder struct {
	Repo   string // repo to fetch from
	Owner string // repo owner
	Url    string // url to fc-session-cache instance to get tokens from
	Cookie string // session cookie value to fetch token from
}

type AuthResponse struct {
	Token string `json:"access_token"`
}

func (b *Builder) Auth() (success bool, gc github.Client, ct context.Context, err error) {
	ctx := context.Background()

	client := &http.Client{}

	body, err := json.Marshal(map[string]string{
		"cookie": b.Cookie,
	})

	if err != nil {
		return false, github.Client{}, ctx, err
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/get", b.Url),
		bytes.NewReader(body),
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

func (b *Builder) DownloadRepo() error {
	ok, client, ctx, err := b.Auth()

	if err != nil {
		return err
	}

	if ok != true {
		return errors.New("There was an error authenticating the current user!")
	}
	
	repo, _, err := client.Repositories.Get(ctx, b.Owner, b.Repo)
	
	if err != nil {
		return err
	}

	zipball, err := http.Get(*repo.ArchiveURL)

	if err != nil {
		return err
	}

	defer zipball.Body.Close()

	file, err := os.Create(fmt.Sprintf("/tmp/fc-builder/%s-%s", repo.Owner.ID, *repo.Name))

	if err != nil {
		return err
	}

	defer file.Close()

	_, downloadErr := io.Copy(file, zipball.Body)

	if downloadErr != nil {
		return downloadErr
	}

	fmt.Println(fmt.Sprintf("Downloaded %s-%s", repo.Owner.ID, *repo.Name))

	return nil
}
