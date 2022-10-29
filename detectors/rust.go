package detectors

type RustDetector struct {}

func (rd *RustDetector) Detect(path string) (bool, string, error) {
	contains, err := Contains(path, []string{"Cargo.toml"})

	if err != nil {
		return false, "", err
	}

	if contains == true {
		return true, "rust", nil
	}

	return false, "", nil
}
