package admin

import "net/http"

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("POST /admin", RequireLocalhost(changeLogLevel))
	mux.Handle("POST /admin/reset", RequireLocalhost(resetLogLevel))
	return mux
}
