package detectors

type PythonDetector struct {}

func (pd *PythonDetector) Detect(path string) (bool, string, error) {
	contains, err := Contains(path, []string{"main.py", "requirements.txt", "pyproject.toml"})

	if err != nil {
		return false, "", err
	}

	if contains == true {
		return true, "python", nil
	}

	return false, "", nil
}
