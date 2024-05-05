package templates

type StaticFileFinder interface {
	FileURL(string) string
}

func CustomFunctionMap(staticFiles StaticFileFinder) map[string]any {
	return map[string]any{
		"staticURL": staticFiles.FileURL,
	}
}
