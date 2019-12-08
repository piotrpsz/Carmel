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

package enigma

import (
	"Carmel/rsakeys"
	"Carmel/secret"
	"Carmel/secret/enigma/blowfish"
	"Carmel/secret/enigma/ghost"
	"Carmel/secret/enigma/way3"
	"Carmel/shared"
	"Carmel/shared/tr"
	"Carmel/shared/vtc"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
)

type Enigma struct {
	ServerId       []byte             // 128 bytes identifying the server
	ClientId       []byte             // 128 bytes identifying the client
	privateKey     *rsa.PrivateKey    // my private RSA key
	buddyPublicKey *rsa.PublicKey     // client's RSA public key
	bf             *blowfish.Blowfish // blowfish
	gt             *ghost.Gost        // ghost
	w3             *way3.Way3         // 3-way
	Keys           vtc.Keys
}

func New(buddyName string) *Enigma {
	// Determining server and client identifiers
	// The keys depend on the day and the month number
	// (they are different every day of the year).
	_, month, day, _, _, _ := shared.DateTimeComponents(shared.Now())
	idx0 := 64 % (month + day)
	idx0++
	idx0 *= 2
	idx1 := idx0 + 128
	idx1 += 32

	serverId := vtc.RandomBytes[idx0 : idx0+128]
	clientId := vtc.RandomBytes[idx1 : idx1+128]

	if rsaManager := rsakeys.New(); rsaManager != nil {
		if privateKey := rsaManager.PrivateKeyFromFileForUser(shared.MyUserName); privateKey != nil {
			e := &Enigma{ServerId: serverId, ClientId: clientId, privateKey: privateKey}
			if buddyName != "" {
				if !e.SetBuddyRSAPublicKey(buddyName) {
					return nil
				}
			}
			return e
		}
	}
	return nil
}

// Ta funkcja w serwerze wywoływana jest dopiero po połączeniu.
// Nazwę partnera rozmowy otrzyma przy pierwszej wymianie danych.
// Klient wywołuje tę funkcję przy tworzeniu obiektu 'enigma'.
// Klient z góry musi wiedzieć z kim się łączy.
func (e *Enigma) SetBuddyRSAPublicKey(buddyName string) bool {
	if rsaManager := rsakeys.New(); rsaManager != nil {
		if buddyPublicKey := rsaManager.PublicKeyFromFileForUser(buddyName); buddyPublicKey != nil {
			e.buddyPublicKey = buddyPublicKey
			return true
		}
	}
	return false
}

func (e *Enigma) InitBlowfish(key []byte) bool {
	if len(key) == blowfish.MaxKeyLength {
		if bf := blowfish.New(key); bf != nil {
			e.bf = bf
			e.Keys.Blowfish = key
			return true
		}
	}
	return false
}

func (e *Enigma) InitGhost(key []byte) bool {
	if gt := ghost.New(key); gt != nil {
		e.gt = gt
		e.Keys.Ghost = key
		return true
	}
	return false
}

func (e *Enigma) InitWay3(key []byte) bool {
	if w3 := way3.New(key); w3 != nil {
		e.w3 = w3
		e.Keys.Way3 = key
		return true
	}
	return false
}

// Each data encryption is done with the partner's public RSA key
// The partner decrypts the data with his private RSA key
func (e *Enigma) EncryptRSA(plain []byte) []byte {
	if e.buddyPublicKey != nil {
		if cipher, err := rsa.EncryptPKCS1v15(rand.Reader, e.buddyPublicKey, plain); tr.IsOK(err) {
			return cipher
		}
	}
	return nil
}

// Deciphering the text with my private RSA key
// The data that the partner has encrypted with my public RSA key.
func (e *Enigma) DecryptRsa(cipher []byte) []byte {
	if e.privateKey != nil {
		if plain, err := rsa.DecryptPKCS1v15(rand.Reader, e.privateKey, cipher); tr.IsOK(err) {
			return plain
		}
	}
	return nil
}

// Calculates the signature for the given data
func (e *Enigma) Signature(data []byte) []byte {
	hash := sha512.Sum512(data)
	if sign, err := rsa.SignPKCS1v15(rand.Reader, e.privateKey, crypto.SHA512, hash[:]); tr.IsOK(err) {
		return sign
	}
	return nil
}

// Checking the correctness of the signature for the given data
func (e *Enigma) IsValidSignature(sign, data []byte) bool {
	hash := sha512.Sum512(data)
	if err := rsa.VerifyPKCS1v15(e.buddyPublicKey, crypto.SHA512, hash[:], sign); tr.IsOK(err) {
		return true
	}
	return false
}

// We use a three-stage EDE encryption system (Encryption-Decryption-Encryption)
// 1. Encryption - Blowfish CBC
// 2. Decryption - Ghost ECB
// 3. Encryption - 3-Way CBC
// Notice: we perform all encryptions using a randomly generated IV
func (e *Enigma) Encrypt(plain []byte) []byte {
	if e.bf != nil && e.gt != nil && e.w3 != nil {
		if bfCipher := e.bf.EncryptCBC(plain, nil); bfCipher != nil {
			if gtCipher := e.gt.DecryptECB(bfCipher); gtCipher != nil {
				if cipher := e.w3.EncryptCBC(gtCipher, nil); cipher != nil {
					return cipher
				}
			}
		}
	}
	return nil
}

// See Encrypt
func (e *Enigma) Decrypt(cipher []byte) []byte {
	if e.bf != nil && e.gt != nil && e.w3 != nil {
		if w3Plain := e.w3.DecryptCBC(cipher); w3Plain != nil {
			if gtPlain := e.gt.EncryptECB(w3Plain); gtPlain != nil {
				if plain := e.bf.DecryptCBC(gtPlain); plain != nil {
					return plain
				}
			}
		}
	}
	return nil
}

/********************************************************************
*                                                                   *
*           I N I T   S E C U R E   C O N N E C T I O N             *
*                                                                   *
********************************************************************/

func (e *Enigma) ClearKeys() {
	secret.ClearSlice(&e.Keys.Blowfish)
	secret.ClearSlice(&e.Keys.Ghost)
	secret.ClearSlice(&e.Keys.Way3)
}
