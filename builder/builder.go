package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

func (b *Builder) Auth() (gc github.Client, ct context.Context, err error) {
	ctx := context.Background()

	client := &http.Client{}

	body, err := json.Marshal(map[string]string{
		"cookie": b.Cookie,
	})

	if err != nil {
		return github.Client{}, ctx, err
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/get", b.Url),
		bytes.NewReader(body),
	)

	res, err := client.Do(req)
	
	if err != nil {
		return github.Client{}, ctx, err
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		return github.Client{}, ctx, err
	}

	authResponse := &AuthResponse{}

	if jsonErr := json.Unmarshal(resBody, &authResponse); jsonErr != nil {
		return github.Client{}, ctx, jsonErr
	}	

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authResponse.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(tc)

	return *ghClient, ctx, nil
}

func (b *Builder) DownloadRepo() error {
	client, ctx, err := b.Auth()

	if err != nil {
		return err
	}
	
	repo, _, err := client.Repositories.Get(ctx, b.Owner, b.Repo)
	
	if err != nil {
		return err
	}

	return nil
}
