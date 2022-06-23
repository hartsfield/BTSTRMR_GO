package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"
)

// track is an audio track
type track struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
	Image  string `json:"image"`
	Path   string `json:"path"`
	ID     string `json:"id"`
}

// page models the data structure of our html page, in this case just a list of
// tracks
type page struct {
	Tracks []*track `json:"tracks"`
}

var (
	templates = template.Must(template.New("main").ParseGlob("internal/*/*.tmpl"))
	tracks    = []*track{}
)

func main() {
	// for generating IDs
	rand.Seed(time.Now().UTC().UnixNano())

	// multiplexer with / and /public set up. /public is our public assets
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// Server configuration
	srv := &http.Server{
		// in production only use SSL
		Addr:              ":8089",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	ctx, cancelCtx := context.WithCancel(context.Background())

	// This can be used as a template for running concurrent servers
	// https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go
	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server two closed\n")
		} else if err != nil {
			fmt.Printf("error listening for server two: %s\n", err)
		}
		cancelCtx()
	}()

	fmt.Println("Server started @ " + srv.Addr)

	// without this the program would not stay running
	<-ctx.Done()
}
