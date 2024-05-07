package domain

import "strings"

type UserDetails struct {
	Name string
}

func (d UserDetails) Validate() map[string]string {
	problems := make(map[string]string)

	if strings.TrimSpace(d.Name) == "" {
		problems["name"] = "This field is required."
	}

	if len(d.Name) > 50 {
		problems["name"] = "Your name cannot be more than 50 characters."
	}

	return problems
}
