package builder

import (
	"bufio"
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

	return fmt.Sprintf("%s:%s", ip, port)
}

func Build(path string, imageName string) error {
	fmt.Println("building...")
	registry := getRegistry()
	fmt.Println(registry)

	box := box.New(box.Config{Px: 2, Py: 5, Type: "Round", Color: "White"})
	box.Print(imageName, fmt.Sprintf("Building image %s...", imageName))

	build := exec.Command(
		"img",
		"build",
		"-t",
		fmt.Sprintf("%s/%s", registry, imageName),
		path,
	)

	// build.Dir = path

	stderr, _ := build.StderrPipe()

	if err := build.Start(); err != nil {
		fmt.Println(err)
		return err
	}

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	if waitErr := build.Wait(); waitErr != nil {
		fmt.Println(waitErr)
		return waitErr
	}

	fmt.Printf("Image %s was built successfully.\n", imageName)
	fmt.Println("Pushing image...")

	ip := os.Getenv("FC_REGISTRY_SERVICE_HOST")
	port := os.Getenv("FC_REGISTRY_SERVICE_PORT")

	if ip == "" || port == "" {
		return errors.New("the registry is not active")
	}

	push := exec.Command(
		"img",
		"push",
		fmt.Sprintf("%s/%s", registry, imageName),
	)

	// push.Dir = path

	stderr, _ = push.StderrPipe()

	if err := push.Start(); err != nil {
		return err
	}

	scanner = bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}


	if waitErr := push.Wait(); waitErr != nil {
		return waitErr
	}

	fmt.Printf("Image %s was built successfully.\n", imageName)

	return nil
}
