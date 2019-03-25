package registration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	// TODO: Make this discoverable
	registrationURL       = "http://localhost:8000/v1/o/register"
	softwareStatementPath = "software_statement.json"
)

func Sign(ss []byte) string {
	claims := &jwt.MapClaims{}
	if err := json.Unmarshal(ss, claims); err != nil {
		log.Fatalln(err)
	}
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	keyPEM, err := ioutil.ReadFile("private.pem")
	if err != nil {
		log.Fatalln(err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyPEM)
	if err != nil {
		log.Fatalln(err)
	}

	jwk := jose.JSONWebKey{
		Key: key,
	}

	token.Header["jwk"] = jwk.Public()

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(key)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Here's the software statement signed by me!\n\n\n\n\n", tokenString, "\n\n\n")
	return tokenString
}

func SoftwareStatement() []byte {
	content, err := ioutil.ReadFile(softwareStatementPath)
	if err != nil {
		log.Fatal(err)
	}
	return content
}

type Request struct {
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

func RequestCredentials(values Request) *oauth2.Config {
	jsonValue, err := json.Marshal(values)
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := http.Post(registrationURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode >= 300 {
		log.Fatalln(fmt.Errorf("Failed to request certification: %s", resp.Status))
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	c := &oauth2Config{}
	if err := json.Unmarshal(b, c); err != nil {
		log.Fatalln(err)
	}

	return (*oauth2.Config)(c)
}
