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

func FinalProcessResponseHandler(cfg *config.ConfigWebHook) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ProcessResponseWriter(cfg, w)
	})

}

func ProcessResponseHandler(next http.Handler, cfg *config.ConfigWebHook) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ProcessResponseWriter(cfg, w)
		next.ServeHTTP(w, r)
	})

}
