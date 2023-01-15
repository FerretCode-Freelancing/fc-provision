package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	events "github.com/ferretcode-freelancing/fc-bus"
	"github.com/ferretcode-freelancing/fc-provision/builder"
	"github.com/google/uuid"
	"github.com/kubemq-io/kubemq-go"
)

type BuildRequest struct {
	Repo   string `json:"repo_url"`
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

func StartBuilder() chan struct{} {
	ctx := context.Background()

	bus := events.Bus{
		Channel: "build-pipeline",
		ClientId: uuid.NewString(),
		Context: ctx,
		TransportType: kubemq.TransportTypeGRPC,
	}

	client, err := bus.Connect()

	if err != nil {
		log.Fatal(fmt.Sprintf("There was an error starting the builder: %s", err))
	}

	fmt.Println("bus is connected")

	done, err := bus.Subscribe(client, func(msgs *kubemq.ReceiveQueueMessagesResponse, subscribeErr error) {
		fmt.Println("message received")

		err := Build(msgs)

		if err != nil {
			log.Printf("There was an error building the image: %s", err)
		}
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("There was an error subscribing to the bus: %s", err))
	}

	return done 
}

func Build(msgs *kubemq.ReceiveQueueMessagesResponse) error {
	ctx := context.Background()

	bus := events.Bus{
		Channel: "deploy-pipeline",
		ClientId: uuid.NewString(),
		Context: ctx,
		TransportType: kubemq.TransportTypeGRPC,
	}

	client, connectErr := bus.Connect()

	if connectErr != nil {
		return connectErr
	}

	message := msgs.Messages[0]

	br := &BuildRequest{}

	if err := json.Unmarshal(message.Body, &br); err != nil {
		return err
	}

	if br.Url == "" {
		username := strings.Trim(os.Getenv("FC_SESSION_CACHE_USERNAME"), "\n")
		password := strings.Trim(os.Getenv("FC_SESSION_CACHE_PASSWORD"), "\n")

		ip := os.Getenv("FC_SESSION_CACHE_SERVICE_HOST")
		port := os.Getenv("FC_SESSION_CACHE_SERVICE_PORT")

		if username == "" || password == "" || ip == "" || port == "" {
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
		return downloadErr
	}

	path, err := extractor.ExtractRepo(ownerId, repo)

	if err != nil {
		return err
	}

	processor := builder.Processor{
		Path: path,
	}

	copyErr := processor.CopyDockerfile()

	if copyErr != nil {
		return err
	}

	imageName := fmt.Sprintf("%s-%s", strings.ToLower(br.Owner), strings.ToLower(repo))

	res, err := UpdateProject(imageName, *br)

	if err != nil {
		return err
	}
		
	if res.StatusCode != 200 {
		return err
	}

	buildErr := builder.Build(
		path,
		imageName,
	)

	if buildErr != nil {
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
		return err
	}

	_, sendErr := client.Send(ctx, kubemq.NewQueueMessage().
		SetId(uuid.NewString()).
		SetChannel(bus.Channel).
		SetBody(stringified))

	if sendErr != nil {
		return err
	}

	return nil
}

func UpdateProject(
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
