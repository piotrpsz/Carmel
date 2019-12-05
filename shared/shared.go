package shared

import (
	"Carmel/shared/tr"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	AppName    = "Carmel"
	AppVersion = "0.1.0"
	AppSubname = "Secure communicator"
	appDir     = ".carmel"
	keysDir    = "rsa_keys"

	IPClipboardMark   = "IP:"
	PortClipboardMark = "Port:"
	NameClipboardMark = "Name:"
	PINClipboardMark  = "PIN:"

	ConnectionTimeout = 30 // in seconds (1 min)
)

var (
	MyIPAddr   string
	MyUserName string
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

func AreFloat32Equal(a, b float32) bool {
	epsilon := 0.000_01
	diff := float64(a) - float64(b)
	return math.Abs(diff) < epsilon
}

func IsValidIPAddress(text string) bool {
	if items := strings.Split(text, "."); len(items) == 4 {
		for _, value := range items {
			if len(value) > 4 {
				return false
			}
			if !OnlyDigits(value) {
				return false
			}
			if v, err := strconv.Atoi(value); !tr.IsOK(err) || v > 255 {
				return false
			}
		}
		return true
	}
	return false
}

func IsValidName(text string) bool {
	if text == "" {
		return false
	}
	runes := []rune(text)
	for _, c := range runes {
		if !unicode.IsLower(c) && !unicode.IsDigit(c) {
			return false
		}
	}
	if unicode.IsDigit(runes[0]) {
		return false
	}
	return true
}

func OnlyDigits(text string) bool {
	if text == "" {
		return false
	}
	for _, c := range text {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func OnlyHexDigits(text string) bool {
	hd := []*unicode.RangeTable{unicode.Hex_Digit}

	for _, c := range text {
		if !unicode.IsOneOf(hd, c) {
			return false
		}
	}
	return true
}

/********************************************************************
*                                                                   *
*                       D A T E   &   T I M E                       *
*                                                                   *
********************************************************************/

func Now() time.Time {
	t := time.Now().UTC()
	year, month, day := t.Date()
	hour, min, sec := t.Hour(), t.Minute(), t.Second()
	// without miliseconds
	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}

func TimeAsString(t time.Time) string {
	year, month, day, hour, min, sec := DateTimeComponents(t)
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, min, sec)
}

func DateTimeComponents(t time.Time) (int, int, int, int, int, int) {
	year, month, day := t.Date()
	hour, min, sec := t.Hour(), t.Minute(), t.Second()
	return year, int(month), day, hour, min, sec
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
