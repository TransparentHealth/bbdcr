package app

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

func Start(c *oauth2.Config) {
	s := &server{}
	s.setConf(c)

	log.Println("Now we're ready to help some bene's, starting up the server....!!!\n\n\n\n\n\n")

	http.ListenAndServe(":8080", s)
}

type server struct {
	conf *oauth2.Config
	tok  *oauth2.Token
}

func (s *server) setConf(c *oauth2.Config) {
	c.Endpoint = oauth2.Endpoint{
		AuthURL:  "http://localhost:8000/v1/o/authorize/",
		TokenURL: "http://localhost:8000/v1/o/token/",
	}
	s.conf = c
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
