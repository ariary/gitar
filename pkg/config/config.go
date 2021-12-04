package config

// Config holds the gitar configuration
type Config struct {
	ServerIP  string
	Port      string
	Directory string
	IsCopied  bool
	Tls       bool
}
