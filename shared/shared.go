package shared

import "fmt"

const (
	AppName    = "Carmel"
	AppVersion = "0.1.0"
	AppSubname = "Secure communicator"
)

var (
	MyIPAddr string
)

func AppNameAndVersion() string {
	return fmt.Sprintf("%s  %s", AppName, AppVersion)
}
