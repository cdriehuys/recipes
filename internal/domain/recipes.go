package domain

import (
	"strings"

	"github.com/google/uuid"
)

type NewRecipe struct {
	Id           uuid.UUID
	Title        string
	Instructions string
}

func (r NewRecipe) Validate() map[string]string {
	problems := make(map[string]string)

	if strings.TrimSpace(r.Title) == "" {
		problems["title"] = "This field is required."
	}

	if strings.TrimSpace(r.Instructions) == "" {
		problems["instructions"] = "This field is required."
	}

	return problems
}
