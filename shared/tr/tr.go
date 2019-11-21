package tr

import "fmt"

func IsOK(err error) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
