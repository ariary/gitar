package config

import (
	"log"
	"strings"
	"time"
)

// Config holds the gitar configuration
type Config struct {
	ServerIP         string
	Port             string
	DownloadDir      string
	UploadDir        string
	IsCopied         bool
	Tls              bool
	CertDir          string
	Url              string
	Completion       bool
	Secret           string
	BidirectionalDir string
	Windows          bool
	NoRun            bool
}

type ConfigScp struct {
	Host     string
	Port     string
	User     string
	Password string
	KeyFile  string
}

// Webhook
type ConfigWebHook struct {
	History         History
	Params          []string
	FullBody        bool
	FullHeaders     bool
	OverrideHeaders map[string][]string
	DelHeaders      []string
	ReqHeaders      []string
}

type History struct {
	LastIp   string
	LastTime time.Time
	LastPath string
}

//parseHeadersFromCLI: giving the headers providing by flags determine which flag must be deleted and which must be added/updated
func ParseHeadersFromCLI(headers []string) (nHeaders map[string][]string, dHeaders []string) {
	nHeaders = make(map[string][]string)

	// fill new headers struct
	for i := 0; i < len(headers); i++ {
		flagSanitize := strings.ReplaceAll(headers[i], " ", "") // withdraw useless space
		headerValue := strings.Split(flagSanitize, ":")
		switch len(headerValue) {
		case 2: //[header]:[value]
			header := headerValue[0]
			value := headerValue[1]
			if value != "" {
				nHeaders[header] = append(nHeaders[header], value)
			} else {
				dHeaders = append(dHeaders, header)
			}
		case 1:
			dHeaders = append(dHeaders, headerValue[0])
		default:
			log.Fatal("Wrong argument for -H, --header flag: [header]:[value] or [header]: or [header]")
		}
	}

	return nHeaders, dHeaders
}
