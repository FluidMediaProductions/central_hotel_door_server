package hotel_comms

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"io/ioutil"
	"os"
)

func GetKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if _, err := os.Stat("private.pem"); err == nil {
		key, err := loadPEMKey("private.pem")
		if err != nil {
			return nil, nil, err
		}
		return key, &key.PublicKey, nil
	} else {
		reader := rand.Reader
		bitSize := 2048

		key, err := rsa.GenerateKey(reader, bitSize)
		if err != nil {
			return nil, nil, err
		}

		publicKey := key.PublicKey

		err = savePEMKey("private.pem", key)
		if err != nil {
			return nil, nil, err
		}
		err = savePublicPEMKey("public.pem", publicKey)
		if err != nil {
			return nil, nil, err
		}
		return key, &publicKey, nil
	}
}

func loadPEMKey(fileName string) (*rsa.PrivateKey, error) {
	outFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(outFile)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func savePEMKey(fileName string, key *rsa.PrivateKey) error {
	outFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(outFile, privateKey)
	if err != nil {
		return err
	}
	return nil
}

func savePublicPEMKey(fileName string, pubkey rsa.PublicKey) error {
	asn1Bytes, err := asn1.Marshal(pubkey)
	if err != nil {
		return err
	}

	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	pemfile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer pemfile.Close()

	err = pem.Encode(pemfile, pemkey)
	if err != nil {
		return err
	}
	return nil
}
