package certification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	// TODO: Make this discoverable
	certificationURL         = "http://localhost:8000/v1/certification/requests/"
	certificationField       = "certification"
	certificationStorageFile = "certification.txt"
)

func alreadyCertified() bool {
	_, err := os.Stat("certification.txt")
	return err == nil
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

func Wait() {
	if alreadyCertified() {
		return
	}
	s := &http.Server{
		Addr: ":8080",
	}
	s.Handler = handleCertification(s)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln(err)
	}
}

func Load() string {
	content, err := ioutil.ReadFile(certificationStorageFile)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func SendRequest(ss []byte) {
	statement := &map[string]interface{}{}
	if err := json.Unmarshal(ss, statement); err != nil {
		log.Fatalln(err)
	}

	b, err := json.Marshal(&map[string]interface{}{
		"software_statement": statement,
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Requesting certification...")
	resp, err := http.Post(certificationURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode >= 300 {
		log.Fatalln(fmt.Errorf("Failed to request certification: %s", resp.Status))
	}
}
