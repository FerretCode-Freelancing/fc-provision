package templates

type Templates struct {
	Templates []Template
}

type Template interface {
	CreateTemplate() string
	GetLanguage() string
}

func (t *Templates) GetTemplates() []Template {
	return []Template{
		&GoTemplate{},
		&RubyTemplate{},
		&JsTemplate{},
		&RustTemplate{},
		&PythonTemplate{},
	}
}

func (t *Templates) GetTemplate(language string) Template {
	for _, template := range t.GetTemplates() {
		if template.GetLanguage() == language {
			return template
		}
	}

	return nil	
}
