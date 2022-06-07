package config

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
