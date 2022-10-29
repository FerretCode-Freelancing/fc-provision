package detectors

type GoDetector struct {}

func (gd *GoDetector) Detect(path string) (bool, string, error) {
	contains, err := Contains(path, []string{"go.mod", "main.go"})

	if err != nil {
		return false, "", err
	}

	if contains == true {
		return true, "go", nil
	}

	return false, "", nil
}
