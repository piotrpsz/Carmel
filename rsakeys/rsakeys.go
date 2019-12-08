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

package rsakeys

import (
	"Carmel/shared"
	"Carmel/shared/tr"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	privateKeyType           = "RSA PRIVATE KEY"
	publicKeyType            = "PUBLIC KEY"
	privateKeyFileNameFormat = "%s_priv.pem"
	publicKeyFileNameFormat  = "%s_public.pem"
	keySize                  = 2048
)

type Manager struct {
	dir string
}

func New() *Manager {
	if dir := shared.RSAKeysDir(); dir != "" {
		return &Manager{dir: dir}
	}
	return nil
}

func (m *Manager) MyUserName() string {
	if shared.MyUserName == "" {
		if items, err := ioutil.ReadDir(m.dir); tr.IsOK(err) {
			for _, item := range items {
				if !item.IsDir() {
					if name := getUserNameFromFileName(item.Name()); name != "" {
						shared.MyUserName = name
						fmt.Println(name)
						return shared.MyUserName
					}
				}
			}
		}
	}
	return shared.MyUserName
}

func getUserNameFromFileName(text string) string {
	if idx := strings.Index(text, "_"); idx != -1 {
		name := text[:idx]
		if idx < len(text)-1 {
			if text[idx+1:] == "priv.pem" {
				return name
			}
		}
	}
	return ""
}

func (m *Manager) RemoveKeysFor(userName string) bool {
	return m.RemovePrivateKeyFor(userName) && m.RemovePublicKeyFor(userName)
}

func (m *Manager) RemovePrivateKeyFor(userName string) bool {
	path := filepath.Join(m.dir, fmt.Sprintf(privateKeyFileNameFormat, userName))
	return shared.RemoveFile(path)
}

func (m *Manager) RemovePublicKeyFor(userName string) bool {
	path := filepath.Join(m.dir, fmt.Sprintf(publicKeyFileNameFormat, userName))
	return shared.RemoveFile(path)
}

func (m *Manager) ExistPrivateKeyFor(userName string) bool {
	path := filepath.Join(m.dir, fmt.Sprintf(privateKeyFileNameFormat, userName))
	return shared.ExistsFile(path)
}

func (m *Manager) ExistPublicKeyFor(userName string) bool {
	path := filepath.Join(m.dir, fmt.Sprintf(publicKeyFileNameFormat, userName))
	return shared.ExistsFile(path)
}

func (m *Manager) CreateKeysForUser(userName string) bool {
	if privateKey, err := rsa.GenerateKey(rand.Reader, keySize); tr.IsOK(err) {
		privatePem := privatePemFromKey(privateKey)
		publicPem := publicPemFromKey(privateKey.PublicKey)
		if privatePem != nil && publicPem != nil {
			privateKeyFilePath := filepath.Join(m.dir, fmt.Sprintf(privateKeyFileNameFormat, userName))
			publicKeyFilePath := filepath.Join(m.dir, fmt.Sprintf(publicKeyFileNameFormat, userName))
			if savePemToFile(privateKeyFilePath, privatePem) && savePemToFile(publicKeyFilePath, publicPem) {
				return true
			}
			shared.RemoveFile(privateKeyFilePath)
			shared.RemoveFile(publicKeyFilePath)
		}
	}
	return false
}

func (m *Manager) PrivateKeyFromFileForUser(userName string) *rsa.PrivateKey {
	filePath := filepath.Join(m.dir, fmt.Sprintf(privateKeyFileNameFormat, userName))
	if data, err := ioutil.ReadFile(filePath); tr.IsOK(err) {
		//tr.Info("%s\n%s", userName, string(data))
		if block, _ := pem.Decode(data); block != nil && block.Type == privateKeyType {
			if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); tr.IsOK(err) {
				return privateKey
			}
		}
	}
	return nil
}

func (m *Manager) PublicKeyFromFileForUser(userName string) *rsa.PublicKey {
	filePath := filepath.Join(m.dir, fmt.Sprintf(publicKeyFileNameFormat, userName))
	if data, err := ioutil.ReadFile(filePath); tr.IsOK(err) {
		//tr.Info("%s\n%s", userName, string(data))
		if block, _ := pem.Decode(data); block != nil && block.Type == publicKeyType {
			if publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes); tr.IsOK(err) {
				return publicKey
			}
		}
	}
	return nil
}

func privatePemFromKey(privateKey *rsa.PrivateKey) *pem.Block {
	if encoded := x509.MarshalPKCS1PrivateKey(privateKey); encoded != nil {
		return &pem.Block{Type: privateKeyType, Bytes: encoded}
	}
	return nil
}

func publicPemFromKey(publicKey rsa.PublicKey) *pem.Block {
	if encoded := x509.MarshalPKCS1PublicKey(&publicKey); encoded != nil {
		return &pem.Block{Type: publicKeyType, Bytes: encoded}
	}
	return nil
}

func savePemToFile(filePath string, pemBlock *pem.Block) bool {
	if shared.ExistsFile(filePath) {
		if !shared.RemoveFile(filePath) {
			return false
		}
	}

	if file, err := os.Create(filePath); tr.IsOK(err) {
		defer file.Close()
		if err := pem.Encode(file, pemBlock); tr.IsOK(err) {
			return true
		}
	}
	return false
}
