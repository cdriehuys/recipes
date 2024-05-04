package server

import "net/http"

func indexHandler(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := state.Logger.WithGroup("request").With("method", r.Method, "path", r.URL.Path)
		logger.Info("Handling request.")

		if err := state.TemplateEngine.Write(w, "index", nil); err != nil {
			logger.Error("Failed to execute template.", "error", err)
		}
	}
}

func MakeHandler(state State) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", indexHandler(&state))

	return mux
}
