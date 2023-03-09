package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

const pemData = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAyPji10zKsJEzRtSpEYCEO5asHkm1sFj4n8PJVPAg/i/896ROFN6F
tfWA+Om0A8jxesgIYH1XBTmEWRF02yaU74K/V5C9Z9yMK0Ta6CbUc0um7p/PCvjQ
W2cuOv9NdSycDlK6msPiQD4InqhKgheETIumT0S8N1TY9421e0sY35viHZznomVI
J3Yj0CBuOdmj0XrW8b4CJjWmRcMtBdeduBevEUL/iZfKgEUTeO7kxKRBAqdYes0L
owSrj1N4owVZBvFv8bNByIc7UlK1I1ZLWHGhJrC9WYwXxszROu3NUHceMLNIe5pQ
QqRafNViSzIpH18Xr0spsxodh5m3nZqc3QIDAQAB
-----END RSA PUBLIC KEY-----
`

func loadPublicKey() (*rsa.PublicKey, error) {
	pemBlock, _ := pem.Decode([]byte(pemData))

	publicKey, err := x509.ParsePKCS1PublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil

}

func verifyBinary(binPath string, serverSignature string) (bool, error) {
	publicKey, err := loadPublicKey()
	if err != nil {
		return false, err
	}

	hash := sha256.New()
	data, err := ioutil.ReadFile(binPath)
	if err != nil {
		return false, err
	}
	_, err = hash.Write(data)
	if err != nil {
		return false, err
	}

	sum := hash.Sum(nil)

	rawSignature, err := base64.StdEncoding.DecodeString(serverSignature)
	if err != nil {
		return false, err
	}

	fmt.Printf("Verifying server provided signature...")
	err = rsa.VerifyPSS(publicKey, crypto.SHA256, sum, rawSignature, nil)
	if err != nil {
		return false, err
	}
	fmt.Printf(" Success\n")
	return true, nil
}
