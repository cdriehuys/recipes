package staticfiles

import (
	"net/http"
	"os"
	"path"
	"time"
)

type StaticFilesFromDisk struct {
	BasePath string
}

// No cache code taken from https://stackoverflow.com/a/33881296/3762084
var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

var etagHeaders = []string{
	"ETag",
	"If-Modified-Since",
	"If-Match",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
}

func (s *StaticFilesFromDisk) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Clear out ETag and other headers indicating the possibility of using a cached values before
	// `http.ServeFile` gets the request.
	for _, v := range etagHeaders {
		req.Header.Del(v)
	}

	// Set cache headers indicating that the response should never be stored.
	for k, v := range noCacheHeaders {
		w.Header().Set(k, v)
	}

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
