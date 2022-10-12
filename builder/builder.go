package builder

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Builder struct {
	Repo   string // repo to fetch from
	Url    string // url to fc-session-cache instance to get tokens from
	Cookie string // session cookie value to fetch token from
}

func (b *Builder) Auth() error {
	client := &http.Client{}

	body, err := json.Marshal(map[string]string{
		"cookie": b.Cookie,
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/get", b.Url),
		nil,
	)
}

func (b *Builder) Build() {

}
