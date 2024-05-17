package validation

import "strings"

type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

func (v *Validator) IsValid() bool {
	return len(v.NonFieldErrors) == 0 && len(v.FieldErrors) == 0
}

func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

func (v *Validator) AddFieldError(field, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	v.FieldErrors[field] = message
}

func (v *Validator) CheckField(ok bool, field, message string) {
	if !ok {
		v.AddFieldError(field, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}
