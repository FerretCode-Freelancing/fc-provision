package builder

import (
	"fmt"
	"os/exec"
)

func Build(path string, imageName string) error {
	build := exec.Command(
		"buildah",
		"build",
		"-t",
		imageName,
		".",
	)

	build.Dir = path

	if err := build.Start(); err != nil {
		return err
	}

	if waitErr := build.Wait(); waitErr != nil {
		return waitErr
	}

	fmt.Println(fmt.Sprintf("Image %s was built successfully.", imageName))

	return nil
}
