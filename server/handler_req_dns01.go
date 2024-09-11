package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eggsampler/acme/v3"
)

func doCertReqDns01(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	conf := AcmeConfigs[id]
	if conf == nil {
		w.Write([]byte("{\"err\": 4000,\"msg\": \"No id matched!!!\"}"))
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	err := _doCertReqDns01(conf, w)
	if err != nil {
		fmt.Fprintf(w, "%+v", err)
	}
}

func _doCertReqDns01(aconfig *AcmeConfig, w http.ResponseWriter) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error doCertReqDns01: %v", r)
		}
	}()
	flusher, _ := w.(http.Flusher)

	var Fprintf = func(format string, a ...any) {
		fmt.Fprintf(w, format, a...)
		fmt.Fprintln(w)
		flusher.Flush()
	}
	Fprintf("Dns01 http challenge")

	// make sure a CertPath/ directory exists
	var parentDir string
	parentDir = filepath.Dir(aconfig.CertPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		log.Printf("Making directory path: %s", parentDir)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("error creating certPath parentdir %q: %v", parentDir, err)
		}
	}
	parentDir = filepath.Dir(aconfig.KeyPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		log.Printf("Making directory path: %s", parentDir)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("error creating keyPath parentdir %q: %v", parentDir, err)
		}
	}

	domainList := strings.Split(aconfig.Domains, ",")
	var ids []acme.Identifier
	for _, domain := range domainList {
		ids = append(ids, acme.Identifier{Type: "dns", Value: domain})
	}

	client, err := acme.NewClient(aconfig.DirectoryUrl)
	if err != nil {
		return fmt.Errorf("error connecting to acme directory: %v", err)
	}

	var account acme.Account
	if aconfig.Account != nil {
		Fprintf("Updating existing account: %s", aconfig.Domains)
		account, err = client.UpdateAccount(acme.Account{PrivateKey: pem2key([]byte(aconfig.Account.PrivateKey)), URL: aconfig.Account.Url})
		if err != nil {
			return fmt.Errorf("error updating existing account: %v", err)
		}
	} else {
		Fprintf("Creating new account: %s", aconfig.Domains)
		account, err = createAccount(client, aconfig)
		if err != nil {
			return fmt.Errorf("error creaing new account: %v", err)
		}
	}

	Fprintf("Creating new order for domains: %s", domainList)
	order, err := client.NewOrder(account, ids)
	if err != nil {
		return fmt.Errorf("error creating new order: %v", err)
	}
	Fprintf("Order created: %s", order.URL)
	// loop through each of the provided authorization urls
	dMap := make(map[string]interface{})
	for _, authUrl := range order.Authorizations {
		// fetch the authorization data from the acme service given the provided authorization url
		Fprintf("Fetching authorization: %s", authUrl)
		auth, err := client.FetchAuthorization(account, authUrl)
		if err != nil {
			return fmt.Errorf("error fetching authorization url %q: %v", authUrl, err)
		}
		Fprintf("Fetched authorization: %s", auth.Identifier.Value)
		chal, ok := auth.ChallengeMap[acme.ChallengeTypeDNS01]
		if !ok {
			return fmt.Errorf("unable to find dns challenge for auth %s", auth.Identifier.Value)
		}
		txt := acme.EncodeDNS01KeyAuthorization(chal.KeyAuthorization)

		Fprintf("TXT record to set: %s", txt)
		dns01, err := aconfig.Dns01.NewDNS01()
		if err != nil {
			return fmt.Errorf("no valid dns01 config json provided: %v", err)
		}
		if _, ok := dMap[auth.Identifier.Value]; !ok {
			dns01.DeleteTXT(auth.Identifier.Value)
			dMap[auth.Identifier.Value] = nil
		}
		err = dns01.SetTXT(txt)
		if err != nil {
			return fmt.Errorf("error set txt record: %v", err)
		}
		// wait for record refresh
		for i := 1; i <= 2; i++ {
			Fprintf("Wait %ds, let the txt record update", i*5)
			time.Sleep(time.Second * 5)
		}
		Fprintf("-------------")
		for i := 1; i <= 3; i++ {
			Fprintf("Wait %ds, let the txt record update and check", i*5)
			time.Sleep(time.Second * 5)
			err = checkTxtRecord(auth.Identifier.Value, txt)
			if err != nil {
				Fprintf("%v", err)
			} else {
				break
			}
		}
		Fprintf("-------------")
		for i := 1; i <= 6; i++ {
			Fprintf("Wait %ds, let the txt record update", i*5)
			time.Sleep(time.Second * 5)
		}

		// update the acme server that the challenge file is ready to be queried
		Fprintf("Updating challenge for authorization %s: %s", auth.Identifier.Value, chal.URL)
		chal, err = client.UpdateChallenge(account, chal)
		if err != nil {
			return fmt.Errorf("error updating authorization %s challenge: %v", auth.Identifier.Value, err)
		}
		Fprintf("Challenge updated")
	}
	// all the challenges should now be completed

	// create a csr for the new certificate
	Fprintf("Generating certificate private key")
	certKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("error generating certificate key: %v", err)
	}

	b := key2pem(certKey)

	// write the key to the key file as a pem encoded key
	Fprintf("Writing key file: %s", aconfig.KeyPath)
	if err := os.WriteFile(string(aconfig.KeyPath), b, 0600); err != nil {
		return fmt.Errorf("error writing key file %q: %v", aconfig.KeyPath, err)
	}

	// create the new csr template
	Fprintf("Creating csr")
	tpl := &x509.CertificateRequest{
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		PublicKeyAlgorithm: x509.ECDSA,
		PublicKey:          certKey.Public(),
		Subject:            pkix.Name{CommonName: domainList[0]},
		DNSNames:           domainList,
	}
	csrDer, err := x509.CreateCertificateRequest(rand.Reader, tpl, certKey)
	if err != nil {
		return fmt.Errorf("error creating certificate request: %v", err)
	}
	csr, err := x509.ParseCertificateRequest(csrDer)
	if err != nil {
		return fmt.Errorf("error parsing certificate request: %v", err)
	}

	// finalize the order with the acme server given a csr
	Fprintf("Finalising order: %s", order.URL)
	order, err = client.FinalizeOrder(account, order, csr)
	if err != nil {
		return fmt.Errorf("error finalizing order: %v", err)
	}

	// fetch the certificate chain from the finalized order provided by the acme server
	Fprintf("Fetching certificate: %s", order.Certificate)
	certs, err := client.FetchCertificates(account, order.Certificate)
	if err != nil {
		return fmt.Errorf("error fetching order certificates: %v", err)
	}

	// write the pem encoded certificate chain to file
	Fprintf("Saving certificate to: %s", aconfig.CertPath)
	var pemData []string
	for _, c := range certs {
		pemData = append(pemData, strings.TrimSpace(string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: c.Raw,
		}))))
	}
	if err := os.WriteFile(aconfig.CertPath, []byte(strings.Join(pemData, "\n")), 0600); err != nil {
		return fmt.Errorf("error writing certificate file %q: %v", aconfig.CertPath, err)
	}

	Fprintf("Done.")
	return nil
}
