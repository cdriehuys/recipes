package htmlutils_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cdriehuys/recipes/internal/htmlutils"
	"golang.org/x/net/html"
)

func TestFindSelectInput(t *testing.T) {
	testCases := []struct {
		name       string
		selectName string
		html       string
		want       htmlutils.Select
		wantErr    bool
	}{
		{
			name:       "empty input",
			selectName: "missing",
			html:       "",
			wantErr:    true,
		},
		{
			name:       "empty select",
			selectName: "test",
			html:       `<select name="test"></select>`,
			want: htmlutils.Select{
				Options: nil,
			},
		},
		{
			name:       "deeply nested",
			selectName: "test",
			html: `<body>
				<main>
					<section>
						<div>
							<select name="test"></select>
						</div>
					</section>
				</main>
			</body>`,
			want: htmlutils.Select{
				Options: nil,
			},
		},
		{
			name:       "unselected values",
			selectName: "test",
			html: `<select name="test">
				<option value="1">One</option>
				<option value="2">Two</option>
			</select>`,
			want: htmlutils.Select{
				Options: []htmlutils.Option{
					{Value: "1", Selected: false},
					{Value: "2", Selected: false},
				},
			},
		},
		{
			name:       "selected value",
			selectName: "test",
			html: `<select name="test">
				<option value="1">One</option>
				<option value="2" selected>Two</option>
			</select>`,
			want: htmlutils.Select{
				Options: []htmlutils.Option{
					{Value: "1", Selected: false},
					{Value: "2", Selected: true},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Invalid HTML: %v", err)
			}

			got, err := htmlutils.FindSelectInput(node, tt.selectName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Received unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Expected %v; received %v", tt.want, got)
			}
		})
	}
}
