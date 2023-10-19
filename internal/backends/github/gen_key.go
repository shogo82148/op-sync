//go:build ignore

package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"

	"golang.org/x/crypto/nacl/box"
)

func main() {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("public key: %s", base64.StdEncoding.EncodeToString(pub[:]))
	log.Printf("private key: %s", base64.StdEncoding.EncodeToString(priv[:]))
}
