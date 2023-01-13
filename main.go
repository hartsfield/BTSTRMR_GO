package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
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

// ckey/ctxkey is used as the key for the HTML context and is how we retrieve
// token information and pass it around to handlers
type ckey int

const (
	ctxkey ckey = iota
)

var (
	// NOTE: The following two variables are initiated through your
	// operating system environment variables and are required for
	// TagMachine to work properly

	// hmacss=hmac_sample_secret
	// testPass=testingPassword

	// hmacSampleSecret is used for creating the token
	hmacSampleSecret = []byte(os.Getenv("hmacss"))

	// connect to redis
	redisIP = os.Getenv("redisIP")
	rdb     = redis.NewClient(&redis.Options{
		Addr:     redisIP + ":6379",
		Password: "",
		DB:       0,
	})

	// HTML templates. We use them like components and compile them
	// together at runtime.
	templates = template.Must(template.New("main").ParseGlob("internal/*/*.tmpl"))
	// this context is used for the client/server connection. It's useful
	// for passing the token/credentials around.
	rdbctx = context.Background()

	tracks = []*track{}
)

func main() {
	// for generating IDs
	rand.Seed(time.Now().UTC().UnixNano())

	// multiplexer with / and /public set up. /public is our public assets
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/api/signup", signup)
	mux.HandleFunc("/api/signin", signin)
	mux.HandleFunc("/api/logout", logout)
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// Server configuration
	srv := &http.Server{
		// in production only use SSL
		Addr:              ":5555",
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
