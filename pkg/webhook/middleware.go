package webhook

import (
	"net/http"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ProcessRequest(r)
		next.ServeHTTP(w, r)
	})
}

func FinalHandler(w http.ResponseWriter, r *http.Request) {
	ProcessResponseWriter(w)
}
