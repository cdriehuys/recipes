package staticfiles

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

type HashedStaticFiles struct {
	baseFS fs.FS

	hashedToPath map[string]string
	pathToHashed map[string]string
}

func NewHashedStaticFiles(logger *slog.Logger, files fs.FS, baseURL string) (HashedStaticFiles, error) {
	hashedToPath := make(map[string]string)
	pathToHashed := make(map[string]string)

	collectStaticFiles := func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}

		f, err := files.Open(filePath)
		if err != nil {
			return fmt.Errorf("could not open %s: %w", filePath, err)
		}
		defer f.Close()

		hash := sha256.New()
		if _, err := io.Copy(hash, f); err != nil {
			return fmt.Errorf("failed to compute hash for %s: %w", filePath, err)
		}

		relPath, err := filepath.Rel("static", filePath)
		if err != nil {
			return fmt.Errorf("could not compute relative path for %s: %w", filePath, err)
		}

		dir := filepath.Dir(relPath)
		ext := filepath.Ext(relPath)
		fileName, err := filepath.Rel(dir, relPath)
		if err != nil {
			return fmt.Errorf("could not compute file name for %s: %w", relPath, err)
		}

		baseName := strings.TrimSuffix(fileName, ext)
		hashedPath := path.Join(dir, fmt.Sprintf("%s.%x%s", baseName, hash.Sum(nil), ext))

		hashedToPath[hashedPath] = filePath
		pathToHashed[relPath] = baseURL + hashedPath

		return nil
	}

	if err := fs.WalkDir(files, "static", collectStaticFiles); err != nil {
		return HashedStaticFiles{}, fmt.Errorf("failed to collect static files: %w", err)
	}

	return HashedStaticFiles{
		baseFS:       files,
		hashedToPath: hashedToPath,
		pathToHashed: pathToHashed,
	}, nil
}

func (s *HashedStaticFiles) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path, ok := s.hashedToPath[req.URL.Path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.ServeFileFS(w, req, s.baseFS, path)
}

func (s *HashedStaticFiles) FileURL(file string) string {
	url, ok := s.pathToHashed[file]
	if !ok {
		return file
	}

	return url
}
