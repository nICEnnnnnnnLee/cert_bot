package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/eggsampler/acme/v3"
	"github.com/nicennnnnnnlee/cert_bot/dns01"
)

var (
	domains             string
	directoryUrl        string
	contactsList        string
	accountFile         string
	dns01File           string
	exitIfDns01NotValid bool
	certFile            string
	keyFile             string
	dialer              net.Dialer
	dnsServer           string
	txtMaxCheck         int
)

type acmeAccountFile struct {
	PrivateKey string `json:"privateKey"`
	Url        string `json:"url"`
}

func main() {
	flag.StringVar(&directoryUrl, "dirurl", acme.LetsEncryptProduction,
		// flag.StringVar(&directoryUrl, "dirurl", acme.LetsEncryptStaging,
		"acme directory url - defaults to lets encrypt v2 staging url if not provided.\n LetsEncryptProduction = https://acme-v02.api.letsencrypt.org/directory\n LetsEncryptStaging = https://acme-staging-v02.api.letsencrypt.org/directory \n ZeroSSLProduction = https://acme.zerossl.com/v2/DV90")
	flag.StringVar(&contactsList, "contact", "",
		"a list of comma separated contact emails to use when creating a new account (optional, dont include 'mailto:' prefix)")
	flag.StringVar(&domains, "domains", "",
		"a comma separated list of domains to issue a certificate for")
	flag.StringVar(&accountFile, "accountfile", "account.json",
		"the file that the account json data will be saved to/loaded from (will create new file if not exists)")
	flag.StringVar(&dns01File, "dns01file", "dns01.json",
		"the file that the dns01 json data will be loaded from (will exit if not exists)")
	flag.StringVar(&dnsServer, "dnsserver", "8.8.8.8:53",
		"dnsServer to check txt record")
	flag.BoolVar(&exitIfDns01NotValid, "exitifdns01fail", true,
		"exit if dns01 config is not valid, or just manualy set dns txt record")
	flag.IntVar(&txtMaxCheck, "txtmaxcheck", 30,
		"the max time trying to verify the txt record. program will continue after max retries no matter if the txt record is valid or not from local spec")
	flag.StringVar(&certFile, "certfile", "cert.pem",
		"the file that the pem encoded certificate chain will be saved to")
	flag.StringVar(&keyFile, "keyfile", "privkey.pem",
		"the file that the pem encoded certificate private key will be saved to")
	flag.Parse()

	// check domains are provided
	if domains == "" {
		log.Fatal("No domains provided")
	}

	dns01, err := dns01.FromFile(dns01File)
	if err != nil {
		if exitIfDns01NotValid {
			log.Fatalf("%v", err)
		} else {
			log.Println(err)
			log.Println("dns01 config is not valid, you need manualy change the DNS txt record youself")
		}
	}

	// create a new acme client given a provided (or default) directory url
	log.Printf("Connecting to acme directory url: %s", directoryUrl)
	client, err := acme.NewClient(directoryUrl)
	if err != nil {
		log.Fatalf("Error connecting to acme directory: %v", err)
	}

	// attempt to load an existing account from file
	log.Printf("Loading account file %s", accountFile)
	account, err := loadAccount(client)
	if err != nil {
		log.Printf("Error loading existing account: %v", err)
		// if there was an error loading an account, just create a new one
		log.Printf("Creating new account")
		account, err = createAccount(client)
		if err != nil {
			log.Fatalf("Error creaing new account: %v", err)
		}
	}
	log.Printf("Account url: %s", account.URL)

	// collect the comma separated domains into acme identifiers
	domainList := strings.Split(domains, ",")
	var ids []acme.Identifier
	for _, domain := range domainList {
		ids = append(ids, acme.Identifier{Type: "dns", Value: domain})
	}

	// create a new order with the acme service given the provided identifiers
	log.Printf("Creating new order for domains: %s", domainList)
	order, err := client.NewOrder(account, ids)
	if err != nil {
		log.Fatalf("Error creating new order: %v", err)
	}
	log.Printf("Order created: %s", order.URL)

	// loop through each of the provided authorization urls
	dMap := make(map[string]interface{})
	for _, authUrl := range order.Authorizations {
		// fetch the authorization data from the acme service given the provided authorization url
		log.Printf("Fetching authorization: %s", authUrl)
		auth, err := client.FetchAuthorization(account, authUrl)
		if err != nil {
			log.Fatalf("Error fetching authorization url %q: %v", authUrl, err)
		}
		log.Printf("Fetched authorization: %s", auth.Identifier.Value)

		chal, ok := auth.ChallengeMap[acme.ChallengeTypeDNS01]
		if !ok {
			log.Fatalf("Unable to find dns challenge for auth %s", auth.Identifier.Value)
		}
		txt := acme.EncodeDNS01KeyAuthorization(chal.KeyAuthorization)

		log.Println("TXT record to set:", txt)
		if dns01 != nil {
			if _, ok := dMap[auth.Identifier.Value]; !ok {
				dns01.DeleteTXT(auth.Identifier.Value)
				dMap[auth.Identifier.Value] = nil
			}
			err = dns01.SetTXT(txt)
			if err != nil {
				log.Fatalf("Error set txt record: %v", err)
			}
			// wait for record refresh
			for i := 1; i <= txtMaxCheck; i++ {
				log.Printf("Wait %ds, let the txt record update\n", i*5)
				time.Sleep(time.Second * 5)
				err = checkTxtRecord(auth.Identifier.Value, txt)
				if err != nil {
					log.Println(err)
				} else {
					break
				}
			}
			if err != nil {
				log.Println("txt record do not match after a long time")
			}
		} else {
			var input string
			log.Println("Please Press Enter after txt record is set：")
			fmt.Scanln(&input)
			log.Println("Please ensure again:")
			fmt.Scanln(&input)
		}
		// update the acme server that the challenge file is ready to be queried
		log.Printf("Updating challenge for authorization %s: %s", auth.Identifier.Value, chal.URL)
		chal, err = client.UpdateChallenge(account, chal)
		if err != nil {
			log.Fatalf("Error updating authorization %s challenge: %v", auth.Identifier.Value, err)
		}
		log.Printf("Challenge updated")
	}

	// all the challenges should now be completed

	// create a csr for the new certificate
	log.Printf("Generating certificate private key")
	certKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Error generating certificate key: %v", err)
	}

	b := key2pem(certKey)

	// write the key to the key file as a pem encoded key
	log.Printf("Writing key file: %s", keyFile)
	if err := os.WriteFile(keyFile, b, 0600); err != nil {
		log.Fatalf("Error writing key file %q: %v", keyFile, err)
	}

	// create the new csr template
	log.Printf("Creating csr")
	tpl := &x509.CertificateRequest{
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		PublicKeyAlgorithm: x509.ECDSA,
		PublicKey:          certKey.Public(),
		Subject:            pkix.Name{CommonName: domainList[0]},
		DNSNames:           domainList,
	}
	csrDer, err := x509.CreateCertificateRequest(rand.Reader, tpl, certKey)
	if err != nil {
		log.Fatalf("Error creating certificate request: %v", err)
	}
	csr, err := x509.ParseCertificateRequest(csrDer)
	if err != nil {
		log.Fatalf("Error parsing certificate request: %v", err)
	}

	// finalize the order with the acme server given a csr
	log.Printf("Finalising order: %s", order.URL)
	order, err = client.FinalizeOrder(account, order, csr)
	if err != nil {
		log.Fatalf("Error finalizing order: %v", err)
	}

	// fetch the certificate chain from the finalized order provided by the acme server
	log.Printf("Fetching certificate: %s", order.Certificate)
	certs, err := client.FetchCertificates(account, order.Certificate)
	if err != nil {
		log.Fatalf("Error fetching order certificates: %v", err)
	}

	// write the pem encoded certificate chain to file
	log.Printf("Saving certificate to: %s", certFile)
	var pemData []string
	for _, c := range certs {
		pemData = append(pemData, strings.TrimSpace(string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: c.Raw,
		}))))
	}
	if err := os.WriteFile(certFile, []byte(strings.Join(pemData, "\n")), 0600); err != nil {
		log.Fatalf("Error writing certificate file %q: %v", certFile, err)
	}

	log.Printf("Done.")
}

func checkTxtRecord(identifier, expectedValue string) error {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, "udp", dnsServer)
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

func loadAccount(client acme.Client) (acme.Account, error) {
	raw, err := os.ReadFile(accountFile)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error reading account file %q: %v", accountFile, err)
	}
	var aaf acmeAccountFile
	if err := json.Unmarshal(raw, &aaf); err != nil {
		return acme.Account{}, fmt.Errorf("error parsing account file %q: %v", accountFile, err)
	}
	account, err := client.UpdateAccount(acme.Account{PrivateKey: pem2key([]byte(aaf.PrivateKey)), URL: aaf.Url}, getContacts()...)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error updating existing account: %v", err)
	}
	return account, nil
}

func createAccount(client acme.Client) (acme.Account, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error creating private key: %v", err)
	}
	account, err := client.NewAccount(privKey, false, true, getContacts()...)
	if err != nil {
		return acme.Account{}, fmt.Errorf("error creating new account: %v", err)
	}
	raw, err := json.Marshal(acmeAccountFile{PrivateKey: string(key2pem(privKey)), Url: account.URL})
	if err != nil {
		return acme.Account{}, fmt.Errorf("error parsing new account: %v", err)
	}
	if err := os.WriteFile(accountFile, raw, 0600); err != nil {
		return acme.Account{}, fmt.Errorf("error creating account file: %v", err)
	}
	return account, nil
}

func getContacts() []string {
	var contacts []string
	if contactsList != "" {
		contacts = strings.Split(contactsList, ",")
		for i := 0; i < len(contacts); i++ {
			contacts[i] = "mailto:" + contacts[i]
		}
	}
	return contacts
}

func key2pem(certKey *ecdsa.PrivateKey) []byte {
	certKeyEnc, err := x509.MarshalECPrivateKey(certKey)
	if err != nil {
		log.Fatalf("Error encoding key: %v", err)
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
		log.Fatalf("Error decoding key: %v", err)
	}
	return key
}
