package domain

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestNewRecipe_Validate(t *testing.T) {
	type fields struct {
		Title        string
		Instructions string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name: "valid",
			fields: fields{
				Title:        "Some Title",
				Instructions: "Some content",
			},
			want: map[string]string{},
		},
		{
			name: "missing required fields",
			fields: fields{
				Title:        "",
				Instructions: "",
			},
			want: map[string]string{
				"title":        "This field is required.",
				"instructions": "This field is required.",
			},
		},
		{
			name: "empty required fields",
			fields: fields{
				Title:        " ",
				Instructions: "\t",
			},
			want: map[string]string{
				"title":        "This field is required.",
				"instructions": "This field is required.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRecipe{
				// No point in specifying IDs
				Id:           uuid.New(),
				Title:        tt.fields.Title,
				Instructions: tt.fields.Instructions,
			}
			if got := r.Validate(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRecipe.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
