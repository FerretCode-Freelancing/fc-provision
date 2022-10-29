package detectors

type Detectors struct {
	Detectors []Detector
}

type Detector interface {
	Detect(path string) (bool, string, error)
}

func (d *Detectors) GetDetectors() []Detector {
	return []Detector{
		&GoDetector{},
		&RubyDetector{},
		&JsDetector{},
		&RustDetector{}, 
		&PythonDetector{},
	} 	
}
