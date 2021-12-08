package config

// Config holds the gitar configuration
type Config struct {
	ServerIP    string
	Port        string
	DownloadDir string
	UploadDir   string
	IsCopied    bool
	Tls         bool
	AliasUrl    string
	Completion  bool
}
