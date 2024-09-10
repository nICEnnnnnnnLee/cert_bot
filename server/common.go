package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"

	"github.com/eggsampler/acme/v3"
)

func checkTxtRecord(identifier, expectedValue string) error {
	var dialer net.Dialer
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, "udp", "1.1.1.1:53")
		},
	}
	txts, err := resolver.LookupTXT(context.Background(), "_acme-challenge."+identifier)
	// txts, err := net.LookupTXT("_acme-challenge." + identifier)
	if err != nil {
		return err
	}
	if len(txts) == 0 {
		return fmt.Errorf("no txt record found")
	}
	for idx := range txts {
		if txts[idx] == expectedValue {
			return nil
		}
	}
	return fmt.Errorf("expected %s, found %s", expectedValue, txts)
}

func createAccount(client acme.Client, aconfig *AcmeConfig) (acme.Account, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error creating private key: %v", err)
	}
	account, err := client.NewAccount(privKey, false, true)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error creating new account: %v", err)
	}
	aconfig.Account = &Account{PrivateKey: string(key2pem(privKey)), Url: account.URL}
	return account, nil
}

func key2pem(certKey *ecdsa.PrivateKey) []byte {
	certKeyEnc, err := x509.MarshalECPrivateKey(certKey)
	if err != nil {
		log.Panicf("Error encoding key: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: certKeyEnc,
	})
}

func pem2key(data []byte) *ecdsa.PrivateKey {
	b, _ := pem.Decode(data)
	key, err := x509.ParseECPrivateKey(b.Bytes)
	if err != nil {
		log.Panicf("Error decoding key: %v", err)
	}
	return key
}
