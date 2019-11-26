package enigma

import (
	"Carmel/connector/datagram"
	"Carmel/connector/tcpiface"
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
	"encoding/json"
	"log"
)

type Enigma struct {
	ServerId      []byte             // 128 bytes identifying the server
	ClientId      []byte             // 128 bytes identifying the client
	PrivateKey    *rsa.PrivateKey    // my private RSA key
	MatePublicKey *rsa.PublicKey     // client's RSA public key
	bf            *blowfish.Blowfish // blowfish
	gt            *ghost.Gost        // ghost
	w3            *way3.Way3         // 3-way
	Keys          vtc.Keys
}

func New() *Enigma {
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
			return &Enigma{ServerId: serverId, ClientId: clientId, PrivateKey: privateKey}
		}
	}
	return nil
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
	if e.MatePublicKey != nil {
		if cipher, err := rsa.EncryptPKCS1v15(rand.Reader, e.MatePublicKey, plain); tr.IsOK(err) {
			return cipher
		}
	}
	return nil
}

// Deciphering the text with my private RSA key
// The data that the partner has encrypted with my public RSA key.
func (e *Enigma) DecryptRsa(cipher []byte) []byte {
	if e.PrivateKey != nil {
		if plain, err := rsa.DecryptPKCS1v15(rand.Reader, e.PrivateKey, cipher); tr.IsOK(err) {
			return plain
		}
	}
	return nil
}

// Calculates the signature for the given data
func (e *Enigma) Signature(data []byte) []byte {
	hash := sha512.Sum512(data)
	if sign, err := rsa.SignPKCS1v15(rand.Reader, e.PrivateKey, crypto.SHA512, hash[:]); tr.IsOK(err) {
		return sign
	}
	return nil
}

// Checking the correctness of the signature for the given data
func (e *Enigma) IsValidSignature(sign, data []byte) bool {
	hash := sha512.Sum512(data)
	if err := rsa.VerifyPKCS1v15(e.MatePublicKey, crypto.SHA512, hash[:], sign); tr.IsOK(err) {
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

func (e *Enigma) InitConnection(iface *tcpiface.TCPInterface, role vtc.RoleType) bool {
	switch role {
	case vtc.Server:
		return e.initConnectionAsServer(iface)
	case vtc.Client:
		return e.initConnectionAsClient(iface)
	}
	return false
}

func (e *Enigma) initConnectionAsServer(iface *tcpiface.TCPInterface) bool {
	if e.exchangeIdentifierBlockAsServer(iface) {
		if e.sendKeys(iface) {
			return true
		}
	}
	return false
}

func (e *Enigma) initConnectionAsClient(iface *tcpiface.TCPInterface) bool {
	if e.exchangeIdentifierBlockAsClient(iface) {
		if e.readKeys(iface) {
			return true
		}
	}
	return false
}

func (e *Enigma) exchangeIdentifierBlockAsServer(iface *tcpiface.TCPInterface) bool {
	// Serwer jako pierwszy wysyła swój blok identyfikacyjny
	if cipher := e.EncryptRSA(e.ServerId); cipher != nil {
		if !datagram.Send(iface, cipher) {
			return false
		}
	}
	// Serwer odczytuje blok identyfikacyjny klienta
	// i sprawdza jego poprawność.
	if cipher := datagram.Read(iface); cipher != nil {
		if data := e.DecryptRsa(cipher); data != nil {
			if !secret.AreSlicesEqual(data, e.ClientId) {
				log.Println("invalid client identifier")
				return false
			}

		}
	}
	return true
}

func (e *Enigma) exchangeIdentifierBlockAsClient(iface *tcpiface.TCPInterface) bool {
	// Klient odczytuje blok identyfikacyjny serwera
	// i sprawdza jego poprawność.
	if cipher := datagram.Read(iface); cipher != nil {
		if data := e.DecryptRsa(cipher); data != nil {
			if !secret.AreSlicesEqual(data, e.ServerId) {
				log.Println("invalid server identifier")
				return false
			}
		}
	}
	// Klient wysyła swój blok identyfikacyjny
	if cipher := e.EncryptRSA(e.ClientId); cipher != nil {
		if !datagram.Send(iface, cipher) {
			return false
		}
	}
	return true
}

func (e *Enigma) sendKeys(iface *tcpiface.TCPInterface) bool {
	defer e.clearKeys()

	if data, err := json.Marshal(e.Keys); tr.IsOK(err) {
		if cipher := e.EncryptRSA(data); cipher != nil {
			if datagram.Send(iface, cipher) {
				return true
			}
		}
	}
	return false
}

func (e *Enigma) readKeys(iface *tcpiface.TCPInterface) bool {
	defer e.clearKeys()

	if cipher := datagram.Read(iface); cipher != nil {
		if data := e.DecryptRsa(cipher); data != nil {
			if err := json.Unmarshal(data, &e.Keys); tr.IsOK(err) {
				if e.InitBlowfish(e.Keys.Blowfish) {
					if e.InitGhost(e.Keys.Ghost) {
						if e.InitWay3(e.Keys.Way3) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (e *Enigma) clearKeys() {
	secret.ClearSlice(&e.Keys.Blowfish)
	secret.ClearSlice(&e.Keys.Ghost)
	secret.ClearSlice(&e.Keys.Way3)
}
