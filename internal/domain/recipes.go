package domain

import (
	"strings"

	"github.com/google/uuid"
)

type NewRecipe struct {
	Id           uuid.UUID
	Owner        string
	Title        string
	Instructions string
}

func (r NewRecipe) Validate() map[string]string {
	problems := make(map[string]string)

	if strings.TrimSpace(r.Title) == "" {
		problems["title"] = "This field is required."
	}

	if len(r.Title) > 200 {
		problems["title"] = "Title must be 200 characters or less."
	}

	if strings.TrimSpace(r.Instructions) == "" {
		problems["instructions"] = "This field is required."
	}

	return problems
}
