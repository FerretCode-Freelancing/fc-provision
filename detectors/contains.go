package detectors

import (
	"io/ioutil"
	"strings"
)

func Contains(path string, matches []string) (bool, error) {
	files, err := ioutil.ReadDir(path)

	if err != nil {
		return false, err
	}

	for _, file := range files {
		for _, match := range matches {
			if strings.Contains(file.Name(), match) {
				return true, nil
			}
		}	
	}

	return false, nil
}
