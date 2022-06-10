package webhook

import (
	"net/http"

	"github.com/ariary/gitar/pkg/config"
)

func Middleware(next http.Handler, cfg *config.ConfigWebHook) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ProcessRequest(r, cfg)
		next.ServeHTTP(w, r)
	})
}

func FinalHandler(w http.ResponseWriter, r *http.Request) {
	ProcessResponseWriter(w)
}
