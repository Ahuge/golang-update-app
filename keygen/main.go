package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func generateKeys() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privatePemData := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})
	err = ioutil.WriteFile("server.pem", privatePemData, 600)
	if err != nil {
		return err
	}

	// The public key is a part of the *rsa.PrivateKey struct
	publicKey := privateKey.PublicKey
	publicPemData := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&publicKey),
		})
	err = ioutil.WriteFile("client.pem", publicPemData, 600)
	if err != nil {
		return err
	}

	return nil
}

func loadPrivateKey() (*rsa.PrivateKey, error) {
	pemData, err := ioutil.ReadFile("server.pem")
	if err != nil {
		return nil, err
	}
	pemBlock, _ := pem.Decode(pemData)

	privateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil

}

func signBinary(binaryPath string) (string, error) {
	privateKey, err := loadPrivateKey()
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	data, err := ioutil.ReadFile(binaryPath)
	if err != nil {
		return "", err
	}
	_, err = hash.Write(data)
	if err != nil {
		return "", err
	}

	sum := hash.Sum(nil)

	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, sum, nil)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(signature)
	return encoded, nil
}

func main() {
	generateFlag := false
	signFlag := ""
	flag.BoolVar(&generateFlag, "generate", false, "generate keys")
	flag.StringVar(&signFlag, "sign", "", "sign a binary")
	flag.Parse()

	if generateFlag {
		if _, err := os.Stat("server.pem"); err == nil {
			fmt.Println("Server key exists")
			return
		}
		err := generateKeys()
		if err != nil {
			fmt.Println("Error generating keys: ", err.Error())
		}
		return
	}

	if signFlag != "" {
		out, err := signBinary(signFlag)
		if err != nil {
			fmt.Printf("Error signing binary \"%s\". %v\n", signFlag, err)
			return
		}
		fmt.Printf("Signature: %s\n", out)
	}

}
