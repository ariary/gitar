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
func NewProxy(targetHost string, history *History) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		ProcessRequest(req, history)
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

func ProcessRequest(req *http.Request, hist *History) {
	//remote addr
	remote := strings.Split(req.RemoteAddr, ":")[0]
	if hist.LastIp != remote {
		hist.LastIp = remote
		remote = color.Yellow(remote)
	} else {
		remote = color.Dim(remote)
	}
	log := remote
	log += color.Dim(" ~ ")
	//time
	now := time.Now()
	rTime := time.Now().Format("09/Jun/2006 15:04:05")
	expiration := hist.LastTime.Add(2 * time.Minute)
	// get the diff
	diff := expiration.Sub(now)
	if diff < 0 {
		rTime = color.Yellow(rTime)
	} else {
		rTime = color.Dim(rTime)
	}
	log += color.Dim("[") + rTime + color.Dim("]")
	log += " ― ― "
	hist.LastTime = now
	//method
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
