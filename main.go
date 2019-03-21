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
	"text/template"

	"golang.org/x/oauth2"

	jwt "github.com/dgrijalva/jwt-go"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	// TODO: Make this discoverable
	registrationURL          = "http://localhost:8000/v1/o/register"
	certificationURL         = "http://localhost:8000/v1/certification/requests/"
	softwareStatementPath    = "software_statement.json"
	certificationField       = "certification"
	certificationStorageFile = "certification.txt"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	generateSigningKey()

	ss := getSoftwareStatement()

	if _, err := os.Stat("certification.txt"); err != nil {

		if err := requestCertification(ss); err != nil {
			log.Fatalln(err)
		}

		if err := waitForCertification(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}

	cert := certification()
	statementToken, err := ss.token()
	if err != nil {
		log.Fatalln(err)
	}

	conf, err := requestCredentials(registrationRequest{
		statementToken,
		[]string{cert},
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Here is the configuration from BlueButton!!!\n\n\n\n\n\n")
	log.Println(conf)

	conf.Endpoint = oauth2.Endpoint{
		AuthURL:  "http://localhost:8000/v1/o/authorize/",
		TokenURL: "http://localhost:8000/v1/o/token/",
	}

	s := &server{
		conf: conf,
	}
	log.Println("Now we're ready to help some bene's, starting up the server....!!!\n\n\n\n\n\n")
	http.ListenAndServe(":8080", s)

	log.Println("yay! it all worked!")
}

type server struct {
	conf *oauth2.Config
	tok  *oauth2.Token
}

func (s *server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code != "" {
		tok, err := s.conf.Exchange(context.Background(), code)
		if err != nil {
			log.Println(err)
			http.Error(rw, err.Error(), 500)
			return
		}
		s.tok = tok
	}
	if s.tok != nil {
		client := s.conf.Client(context.Background(), s.tok)
		resp, _ := client.Get("http://localhost:8000/v1/fhir/Patient/")
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			http.NotFound(rw, r)
		}
		rw.Write(b)
		return
	}
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := s.conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", url)
	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Println(err)
		http.NotFound(rw, r)
		return
	}
	if err := t.Execute(rw,
		struct {
			URL string
		}{
			URL: url,
		}); err != nil {
		log.Println(err)
		http.NotFound(rw, r)
		return
	}
	return
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

	jwk := jose.JSONWebKey{
		Key: key,
	}

	token.Header["jwk"] = jwk.Public()

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(key)

	log.Println("Here's the software statement signed by me!\n\n\n\n\n", tokenString, "\n\n\n")
	return tokenString, err
}

func handleCertification(s *http.Server) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Println("Hey we're certified!")
		certification := r.FormValue(certificationField)
		log.Println(certification, "\n\n\n\n")
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
	statement := &map[string]interface{}{}
	if err := json.Unmarshal(ss, statement); err != nil {
		return err
	}

	b, err := json.Marshal(&map[string]interface{}{
		"software_statement": statement,
	})

	if err != nil {
		return err
	}

	resp, err := http.Post(certificationURL, "application/json", bytes.NewBuffer(b))
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

type oauth2Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Endpoint     oauth2.Endpoint
	RedirectURL  string
	Scopes       []string
}

func requestCredentials(values registrationRequest) (*oauth2.Config, error) {
	jsonValue, err := json.Marshal(values)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	resp, err := http.Post(registrationURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Failed to request certification: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(string(b))

	c := &oauth2Config{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	return (*oauth2.Config)(c), nil
}

func generateSigningKey() {
	if _, err := os.Stat("private.pem"); err == nil {
		return
	}

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
