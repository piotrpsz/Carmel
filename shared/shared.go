/*
 * BSD 2-Clause License
 *
 *	Copyright (c) 2019, Piotr Pszczółkowski
 *	All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 * list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 * this list of conditions and the following disclaimer in the documentation
 * and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package shared

import (
	"Carmel/shared/tr"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
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
	MyInternetIP string
	MyLocalIP    string
	MyUserName   string
)

func init() {
	MyInternetIP = internetIP()
	MyLocalIP = localIP()
}

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

func internetIP() string {
	if response, err := http.Get("https://api.ipify.org/?format=json"); tr.IsOK(err) {
		defer response.Body.Close()
		if content, err := ioutil.ReadAll(response.Body); tr.IsOK(err) {
			data := make(map[string]interface{})
			if err := json.Unmarshal(content, &data); tr.IsOK(err) {
				if text, ok := data["ip"].(string); ok {
					return text
				}
			}
		}
	}
	return ""
}

func localIP() string {
	if addresses := lookup(); addresses != nil {
		for _, ip := range addresses {
			if ip.To4() != nil && !ip.IsLoopback() {
				return ip.String()
			}
		}
	}
	return ""
}

func lookup() []net.IP {
	if hostname, err := os.Hostname(); tr.IsOK(err) {
		if retv, err := net.LookupIP(hostname); tr.IsOK(err) {
			return retv
		}
	}
	return nil
}
