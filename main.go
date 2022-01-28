package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/factorysh/microdensity/application"
	"github.com/factorysh/microdensity/queue"
	_sessions "github.com/factorysh/microdensity/sessions"
	"github.com/factorysh/microdensity/version"
	"go.etcd.io/bbolt"
)

func main() {
	fmt.Println("Version", version.Version())

	sessions := _sessions.New()
	// prune old sessions every 15 minutes
	ticker := time.NewTicker(15 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				sessions.Prune()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	s, err := bbolt.Open(
		"/tmp/microdensity.store",
		0600, &bbolt.Options{})
	if err != nil {
		log.Fatal(err)
	}
	q, err := queue.New(s)
	if err != nil {
		log.Fatal(err)
	}
	a, err := application.New(q, "s3cr27", "/tmp/microdensity")
	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe("127.0.0.1:3000", a.Router)
}
