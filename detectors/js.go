package detectors

type JsDetector struct {}

func (jd *JsDetector) Detect(path string) (bool, string, error) {
	contains, err := Contains(path, []string{"package.json", "index.js"})

	if err != nil {
		return false, "", err
	}

	if contains == true {
		return true, "js", nil
	}

	return false, "", nil
}
