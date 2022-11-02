package builder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Delta456/box-cli-maker/v2"
)

func getRegistry() string {
	ip := strings.Trim(os.Getenv("FC_REGISTRY_SERVICE_HOST"), "\n")
	port := strings.Trim(os.Getenv("FC_REGISTRY_SERVICE_PORT"), "\n")

	return fmt.Sprintf("http://%s:%s", ip, port)
}

func Build(path string, imageName string) error {
	registry := getRegistry()

	build := exec.Command(
		"buildah",
		"build",
		"-t",
		fmt.Sprintf("%s/%s", registry, imageName),
		".",
	)

	build.Dir = path

	if err := build.Start(); err != nil {
		return err
	}

	if waitErr := build.Wait(); waitErr != nil {
		return waitErr
	}

	box := box.New(box.Config{Px: 2, Py: 5, Type: "Round", Color: "White"})
	box.Print(imageName, fmt.Sprintf("Building image %s...", imageName))

	fmt.Printf("Image %s was built successfully.\n", imageName)
	fmt.Println("Pushing image...")

	ip := os.Getenv("FC_REGISTRY_SERVICE_HOST")
	port := os.Getenv("FC_REGISTRY_SERVICE_PORT")

	if ip == "" || port == "" {
		return errors.New("the registry is not active")
	}

	push := exec.Command(
		"buildah",
		"push",
		imageName,
		fmt.Sprintf("%s:%s", ip, port),
	)

	push.Dir = path

	if err := push.Start(); err != nil {
		return err
	}

	if waitErr := push.Wait(); waitErr != nil {
		return waitErr
	}

	fmt.Printf("Image %s was built successfully.\n", imageName)

	return nil
}
