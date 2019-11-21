package shared

import (
	"Carmel/shared/tr"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	AppName    = "Carmel"
	AppVersion = "0.1.0"
	AppSubname = "Secure communicator"
	appDir     = ".carmel"
	keysDir    = "rsa_keys"
)

var (
	MyIPAddr string
)

func AppNameAndVersion() string {
	return fmt.Sprintf("%s %s", AppName, AppVersion)
}

func AppDir() string {
	if homeDir, err := os.UserHomeDir(); tr.IsOK(err) {
		appDir := filepath.Join(homeDir, appDir)
		if CreateDirIfNeeded(appDir) {
			return appDir
		}
	}
	return ""
}

func RSAKeysDir() string {
	if appDir := AppDir(); appDir != "" {
		rsaKeysDir := filepath.Join(appDir, keysDir)
		if CreateDirIfNeeded(rsaKeysDir) {
			return rsaKeysDir
		}
	}
	return ""
}

/********************************************************************
*                                                                   *
*             F I L E S   &   D I R E C T O R I E S                 *
*                                                                   *
********************************************************************/

func ExistsFile(filePath string) bool {
	var err error
	var fi os.FileInfo

	if fi, err = os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			log.Println(err)
		}
		return false
	}

	if fi.IsDir() {
		return false
	}
	return true
}

func ExistsDir(dirPath string) bool {
	var err error
	var fi os.FileInfo

	if fi, err = os.Stat(dirPath); err != nil {
		if !os.IsNotExist(err) {
			log.Println(err)
		}
		return false
	}

	if fi.IsDir() {
		return true
	}
	return false
}

func CreateDirIfNeeded(dirPath string) bool {
	if ExistsDir(dirPath) {
		return true
	}
	if err := os.MkdirAll(dirPath, os.ModePerm); tr.IsOK(err) {
		return true
	}
	return false
}

func RemoveFile(filePath string) bool {
	if err := os.Remove(filePath); tr.IsOK(err) {
		return true
	}
	return false
}

func ReadFromFile(filePath string) []byte {
	if handle, err := os.OpenFile(filePath, os.O_RDONLY, 0666); tr.IsOK(err) {
		if buffer, err := ioutil.ReadAll(handle); tr.IsOK(err) {
			return buffer
		}
	}
	return nil
}

func WriteToFile(filePath string, data []byte) bool {
	if handle, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666); tr.IsOK(err) {
		if nbytes, err := handle.Write(data); tr.IsOK(err) {
			return nbytes == len(data)
		}
	}
	return false
}
