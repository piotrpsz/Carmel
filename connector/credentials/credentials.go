package credentials

type Credentials struct {
	IPAddress string `json:"ip"`
	Port      string `json:"port"`
	Name      string `json:"name"`
	PIN       string `json:"pin"`
}
