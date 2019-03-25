package main

import (
	"log"
	"net/http"
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
	GenerateSigningKey()
	softwareStatement = getSoftwareStatement()
}

func main() {

	requestCertification(softwareStatement)

	waitForCertification()

	cert := loadCertification()

	req := registrationRequest{
		signSoftwareStatement(softwareStatement),
		[]string{cert},
	}

	conf := requestCredentials(req)

	log.Println("Here is the configuration from BlueButton!!!\n\n\n\n\n\n", conf)

	s := &server{}
	s.setConf(conf)

	log.Println("Now we're ready to help some bene's, starting up the server....!!!\n\n\n\n\n\n")

	http.ListenAndServe(":8080", s)

	log.Println("yay! it all worked!")
}
