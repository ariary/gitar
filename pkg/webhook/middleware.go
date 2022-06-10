package webhook

import (
	"net/http"
)

func Middleware(next http.Handler, history *History) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ProcessRequest(r, history)
		next.ServeHTTP(w, r)
	})
}

func FinalHandler(w http.ResponseWriter, r *http.Request) {
	ProcessResponseWriter(w)
}
