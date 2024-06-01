package templates

type StaticFileFinder interface {
	FileURL(string) string
}

func CustomFunctionMap(staticFiles StaticFileFinder) map[string]any {
	return map[string]any{
		"formField": formField,
		"staticURL": staticFiles.FileURL,
	}
}

type FormField struct {
	Name  string
	Label string
	Value string
	Error string
}

func asString(value any) string {
	parsed, _ := value.(string)

	return parsed
}

func formField(name, label, value, err any) FormField {
	return FormField{
		Name:  asString(name),
		Label: asString(label),
		Value: asString(value),
		Error: asString(err),
	}
}
