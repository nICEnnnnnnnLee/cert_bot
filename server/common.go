package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	mrand "math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

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

func initProxyUrl(proxyURL string) {
	if proxyURL != "" {
		proxyUrl, err := url.Parse(proxyURL)
		if err != nil {
			log.Println("proxy url is not valid: ", proxyURL)
		} else {
			http.DefaultClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
		}
	}
}

func HashMd5(raw string) string {
	hash := md5.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}
func HashSHA1(raw string) string {
	hash := sha1.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func RandomString(length int) string {
	mrand.Seed(time.Now().UnixNano())
	const letters = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~!@#$%^&*()_+-=[]\{}|;':",./<>?`
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[mrand.Intn(len(letters))]
	}
	return string(b)
}

func splitHostPort(hostPort string) (host, port string) {
	host = hostPort
	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	return
}

func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}
