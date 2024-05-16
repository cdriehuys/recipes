package main

import "net/http"

type templateData struct{}

func newTemplateData(_ *http.Request) templateData {
	return templateData{}
}
