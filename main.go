package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// TODO: Make this discoverable
	registrationURL          = "http://localhost:8000/v1/o/register"
	certificationURL         = "http://localhost:8000/v1/certification/requests/"
	softwareStatementPath    = "software_statement.json"
	certificationField       = "certification"
	certificationStorageFile = "certification.txt"
)

func main() {
	generateSigningKey()
	/*
		if err := requestCertification(softwareStatement()); err != nil {
			log.Fatalln(err)
		}
		if err := waitForCertification(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	*/
	cert := certification()
	statementToken, err := getSoftwareStatement().token()
	if err != nil {
		log.Fatalln(err)
	}

	if err := requestCredentials(registrationRequest{
		statementToken,
		[]string{cert},
	}); err != nil {
		log.Fatalln(err)
	}

	log.Println("yay! it all worked!")
}

type softwareStatement []byte

func (ss softwareStatement) token() (string, error) {
	claims := &jwt.MapClaims{}
	if err := json.Unmarshal(ss, claims); err != nil {
		log.Println(err)
		return "", err
	}
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	keyPEM, err := ioutil.ReadFile("private.pem")
	if err != nil {
		log.Println(err)
		return "", err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyPEM)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(key)

	log.Println(tokenString, err)
	return tokenString, err
}

func handleCertification(s *http.Server) http.Handler {
	log.Println("registering handler")
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Println("got certification")
		certification := r.FormValue(certificationField)
		log.Println(certification)
		err := ioutil.WriteFile(certificationStorageFile, []byte(certification), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		http.Error(rw, "OK", http.StatusOK)
		go s.Shutdown(context.Background())
	})
}

func waitForCertification() error {
	s := &http.Server{
		Addr: ":8080",
	}
	s.Handler = handleCertification(s)
	return s.ListenAndServe()
}

func getSoftwareStatement() softwareStatement {
	content, err := ioutil.ReadFile(softwareStatementPath)
	if err != nil {
		log.Fatal(err)
	}
	return softwareStatement(content)
}

func certification() string {
	content, err := ioutil.ReadFile(certificationStorageFile)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func requestCertification(ss []byte) error {
	resp, err := http.Post(certificationURL, "application/json", bytes.NewBuffer(ss))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("Failed to request certification: %s", resp.Status)
	}
	return nil
}

type registrationRequest struct {
	SoftwareStatement string   `json:"software_statement"`
	Certifications    []string `json:"certifications"`
}

func requestCredentials(values registrationRequest) error {
	jsonValue, err := json.Marshal(values)
	if err != nil {
		log.Println(err)
		return err
	}
	resp, err := http.Post(registrationURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Println(err)
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("Failed to request certification: %s", resp.Status)
	}
	return nil
}

func generateSigningKey() {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		log.Fatalln(err)
	}

	savePEMKey("private.pem", key)
}

func savePEMKey(fileName string, key *rsa.PrivateKey) {
	outFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(outFile, privateKey)
	if err != nil {
		log.Fatalln(err)
	}
}
