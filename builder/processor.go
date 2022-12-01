package builder

import (
	"errors"
	"fmt"
	"os"

	"github.com/ferretcode-freelancing/fc-provision/detectors"
	"github.com/ferretcode-freelancing/fc-provision/templates"
)

type Processor struct {
	Path string // the path to the extracted repo
}

func (p *Processor) DetectLanguage() (string, error) {
	fmt.Println("detecting...")

	ds := &detectors.Detectors{}

	detectors := ds.GetDetectors()

	for _, detector := range detectors {
		detected, language, err := detector.Detect(p.Path)

		if err != nil {
			fmt.Println(err)
			return "", err
		}

		if detected {
			fmt.Println(language)
			return language, nil
		}

		continue
	}

	return "", errors.New("the language could not be detected")
}

func (p *Processor) GetDockerfile() (string, error) {
	existingDockerfile, err := os.ReadFile(
		fmt.Sprintf("%s/Dockerfile", p.Path),
	)

	if err != nil {
		fmt.Println("No dockerfile found already, generating...")
	}

	if len(existingDockerfile) != 0 {
		return string(existingDockerfile), nil
	}

	language, err := p.DetectLanguage()

	if err != nil {
		return "", err
	}

	fmt.Println(language)

	ts := &templates.Templates{}

	template := ts.GetTemplate(language)

	if template == nil {
		return "", errors.New("the dockerfile could not be built")
	}

	return template.CreateTemplate(), nil
}

func (p *Processor) CopyDockerfile() error {
	dockerfile, err := p.GetDockerfile()

	if err != nil {
		fmt.Println(err)
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/Dockerfile", p.Path))

	if err != nil {
		return err
	}

	file.WriteString(dockerfile)

	return nil
}
