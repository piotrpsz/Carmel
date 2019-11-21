package shared

import "fmt"

const (
	AppName    = "Carmel"
	AppVersion = "0.1.0"
	AppSubname = "Secure communicator"
)

func AppNameAndVersion() string {
	return fmt.Sprintf("%s v.  %s", AppName, AppVersion)
}
