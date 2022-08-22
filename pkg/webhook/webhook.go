package webhook

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ariary/gitar/pkg/config"
	"github.com/ariary/go-utils/pkg/color"
	stringSlice "github.com/ariary/go-utils/pkg/stringSlice"
)

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string, cfg *config.ConfigWebHook) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		ProcessRequest(req, cfg)
	}

	proxy.ModifyResponse = ProcessResponse(cfg)
	proxy.ErrorHandler = errorHandler()

	return proxy, nil
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func ProcessRequest(req *http.Request, cfg *config.ConfigWebHook) {
	//remote addr
	remote := strings.Split(req.RemoteAddr, ":")[0]
	if cfg.History.LastIp != remote {
		cfg.History.LastIp = remote
		remote = color.Yellow(remote)
	} else {
		remote = color.Dim(remote)
	}
	log := remote
	log += color.Dim(" ~ ")
	//time
	now := time.Now()
	rTime := time.Now().Format("09/Jun/2006 15:04:05")
	expiration := cfg.History.LastTime.Add(2 * time.Minute)
	// get the diff
	diff := expiration.Sub(now)
	if diff < 0 {
		rTime = color.Yellow(rTime)
	} else {
		rTime = color.Dim(rTime)
	}
	log += color.Dim("[") + rTime + color.Dim("]")
	log += " ― ― "
	cfg.History.LastTime = now
	//method
	switch req.Method {
	case "GET":
		log += color.Blue(req.Method) + " "
	case "POST":
		log += color.Green(req.Method) + " "
	default:
		log += color.Magenta(req.Method) + " "
	}
	//path
	path := req.URL.Path
	if path != cfg.History.LastPath {
		cfg.History.LastPath = path
		path = color.Cyan(path)
	}
	log += path
	// filter header
	rHeaders := req.Header
	if cfg.FullHeaders { //print all headers
		log += "\n"
		for header, value := range rHeaders {
			if stringSlice.Contains(cfg.ReqHeaders, header) {
				header = color.Teal(header)
			}
			log += header + ": " + strings.Join(value, ", ") + "\n"
		}
		log += "\t"
	} else { //only print the specified ones
		for i := 0; i < len(cfg.ReqHeaders); i++ {
			log += "\n"
			value := rHeaders.Get(cfg.ReqHeaders[i])
			if value != "" {
				log += "\t" + color.Teal(cfg.ReqHeaders[i]) + ": " + value
			} else {
				log += "\t" + cfg.ReqHeaders[i] + " header was not found in request"
			}
		}
	}

	// filter params & body
	if cfg.FullBody {
		if bodyB, err := ioutil.ReadAll(req.Body); err != nil {
			fmt.Println(color.Red("error while reading request body"))
		} else {
			log += "\n" + string(bodyB)
		}
	} else {
		var param string
		req.ParseForm()
		for i := 0; i < len(cfg.Params); i++ {
			log += "\n"
			if req.Method == "GET" {
				param = req.URL.Query().Get(cfg.Params[i])
			} else {
				param = req.PostForm.Get(cfg.Params[i])
			}
			if param != "" {
				log += "\t" + color.Teal(cfg.Params[i]) + ": " + param
			} else {
				log += "\t" + color.Dim(cfg.Params[i]) + ": "
			}
		}
	}
	fmt.Println(log)
}
func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Got error while modifying response: %v \n", err)
		return
	}
}

//ProcessResponse & ProcessResponseWriter are the same fucntion, but do not find a way do use only one

func ProcessResponse(cfg *config.ConfigWebHook) func(*http.Response) error {
	return func(resp *http.Response) error {
		respHeader := resp.Header
		// override/add headers
		for header, value := range cfg.OverrideHeaders {
			delete(respHeader, header)
			respHeader[header] = append(respHeader[header], value...)
		}
		// delete headers
		for i := 0; i < len(cfg.DelHeaders); i++ {
			delete(respHeader, cfg.DelHeaders[i])
		}
		return nil
	}
}

func ProcessResponseWriter(cfg *config.ConfigWebHook, resp http.ResponseWriter) {
	respHeader := resp.Header()
	// override/add headers
	for header, value := range cfg.OverrideHeaders {
		delete(respHeader, header)
		respHeader[header] = append(respHeader[header], value...)
	}
	// delete headers
	for i := 0; i < len(cfg.DelHeaders); i++ {
		delete(respHeader, cfg.DelHeaders[i])
	}
}
