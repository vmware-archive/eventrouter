package main

import (
	"io"
	"log"
	"net/http"

	"github.com/crewjam/rfc5424"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("request method=%s from=%s", r.Method, r.RemoteAddr)
	if r.Body == nil {
		return
	}
	defer r.Body.Close()
	m := new(rfc5424.Message)
	for {
		_, err := m.ReadFrom(r.Body)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Parsing rfc5424 message failed: %+v", err)
		}
		log.Printf("%s", m.Message)
	}
}

func main() {
	log.Println("starting httpsink server")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
