package gitar

import (
	"fmt"

	"github.com/ariary/gitar/pkg/config"
)

//PortForwarding: forward all tcp port to specified port in config
func PortForwarding(config *config.Config) {
	fmt.Println("Redirect all the traffic to ", config.RedirectedPort)
}
