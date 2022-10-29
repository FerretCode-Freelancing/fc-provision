package detectors

type RubyDetector struct {}

func (rd *RubyDetector) Detect(path string) (bool, string, error) {
	contains, err := Contains(path, []string{"Gemfile", ".rb"})

	if err != nil {
		return false, "", err
	}

	if contains == true {
		return true, "ruby", nil
	}

	return false, "", nil
}
