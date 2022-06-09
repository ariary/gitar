package webhook

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ariary/go-utils/pkg/color"
)

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		ProcessRequest(req)
	}

	proxy.ModifyResponse = ProcessResponse()
	proxy.ErrorHandler = errorHandler()

	return proxy, nil
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func ProcessRequest(req *http.Request) {
	log := color.Dim(strings.Split(req.RemoteAddr, ":")[0])
	log += color.Dim(" ~ ")
	log += color.Dim("[" + time.Now().Format("09/Jun/2006 15:04:05") + "]")
	log += " ― ― "
	switch req.Method {
	case "GET":
		log += color.Blue(req.Method) + " "
	case "POST":
		log += color.Green(req.Method) + " "
	default:
		log += color.Magenta(req.Method) + " "
	}
	log += req.URL.Path
	// req.Header.Set("X-Proxy", "Simple-Reverse-Proxy")
	fmt.Println(log)
}
func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Got error while modifying response: %v \n", err)
		return
	}
}

func ProcessResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		resp.Header.Set("X-Proxy", "Magical")
		return nil
	}
}

func ProcessResponseWriter(resp http.ResponseWriter) {
	resp.Header().Set("X-Proxy", "Magical")
}
