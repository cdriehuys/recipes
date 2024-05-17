package templates

import "github.com/cdriehuys/recipes/internal/validation"

type StaticFileFinder interface {
	FileURL(string) string
}

func CustomFunctionMap(staticFiles StaticFileFinder) map[string]any {
	return map[string]any{
		"formField":   formField,
		"formInvalid": formInvalid,
		"staticURL":   staticFiles.FileURL,
	}
}

type FormField struct {
	Name  string
	ID    string
	Label string
	Value string
	Error string
}

func asString(value any) string {
	parsed, _ := value.(string)

	return parsed
}

func formField(name, id, label, value, err any) FormField {
	return FormField{
		Name:  asString(name),
		ID:    asString(id),
		Label: asString(label),
		Value: asString(value),
		Error: asString(err),
	}
}

func formInvalid(form any) bool {
	validatable, ok := form.(validation.Validator)
	if !ok {
		return false
	}

	return !validatable.IsValid()
}
