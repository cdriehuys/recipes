package staticfiles

import (
	"net/http"
	"os"
	"path"
)

type StaticFilesFromDisk struct {
	BasePath string
}

func (s *StaticFilesFromDisk) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	filePath := path.Join(s.BasePath, path.Clean(req.URL.Path))
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.ServeFile(w, req, filePath)
}

func (s *StaticFilesFromDisk) FileURL(file string) string {
	return path.Join("/", s.BasePath, path.Clean(file))
}
