package main

import (
	"log"

	"github.com/whytheplatypus/bbdcr/app"
	"github.com/whytheplatypus/bbdcr/certification"
	"github.com/whytheplatypus/bbdcr/crypto"
	"github.com/whytheplatypus/bbdcr/registration"
)

const (
	// TODO: Make this discoverable
	registrationURL          = "http://localhost:8000/v1/o/register"
	certificationURL         = "http://localhost:8000/v1/certification/requests/"
	softwareStatementPath    = "software_statement.json"
	certificationField       = "certification"
	certificationStorageFile = "certification.txt"
)

var (
	softwareStatement []byte
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	crypto.GenerateSigningKey()
	softwareStatement = registration.SoftwareStatement()
}

func main() {

	certification.SendRequest(softwareStatement)

	certification.Wait()

	cert := certification.Load()

	req := registration.Request{
		registration.Sign(softwareStatement),
		[]string{cert},
	}

	conf := registration.RequestCredentials(req)

	log.Println("Here is the configuration from BlueButton!!!\n\n\n\n\n\n", conf)

	app.Start(conf)

	log.Println("yay! it all worked!")
}
