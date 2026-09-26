package http

import (
	"net/http"
)

// Router wraps http.ServeMux and provides route setup
type Router struct {
	*http.ServeMux
}

// V1Handler returns a handler for v1 API routes
func (router *Router) V1Handler() http.Handler {
	mux := http.NewServeMux()

	uowMux := http.NewServeMux()
	uowMux.HandleFunc("GET /enabled", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	mux.Handle("/uow/", http.StripPrefix("/uow", uowMux))

	return mux
}

// SetupMux creates and configures the main router
func SetupMux() *Router {
	router := Router{http.NewServeMux()}
	router.ServeMux.HandleFunc("/heartbeat", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte("OK"))
	})
	router.ServeMux.Handle("/api/v1/", http.StripPrefix("/api/v1", router.V1Handler()))

	return &router
}
