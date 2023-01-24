package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
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
	Likes  int    `json:"likes"`
	Liked  bool   `json:"liked"`
}

// pageData is used in the HTML templates as the main page model. It is
// composed of credentials, postData, and threadData.
type pageData struct {
	UserData *credentials `json:"userData"`
	Tracks   []*track     `json:"tracks"`
	Number   string       `json:"pageNumber,number"`
	PageName string       `json:"pageName"`
	Category string       `json:"category"`
}

// ckey/ctxkey is used as the key for the HTML context and is how we retrieve
// token information and pass it around to handlers
type ckey int

const (
	ctxkey ckey = iota
)

var (
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

	// tracks = []*track{}
)

func main() {
	// for generating IDs
	rand.Seed(time.Now().UTC().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// multiplexer with / and /public set up. /public is our public assets
	mux := http.NewServeMux()
	mux.Handle("/", checkAuth(http.HandlerFunc(freshView)))
	mux.Handle("/fresh", checkAuth(http.HandlerFunc(freshView)))
	mux.Handle("/hot", checkAuth(http.HandlerFunc(hotView)))
	mux.Handle("/api/like", checkAuth(http.HandlerFunc(likeTrack)))
	mux.Handle("/api/getTracks", checkAuth(http.HandlerFunc(getTracks)))
	mux.Handle("/♥/", checkAuth(http.HandlerFunc(likesView)))
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
		if err != nil {
			fmt.Println(err)
		}
		cancelCtx()
	}()

	fmt.Println("Server started @ " + srv.Addr)

	// without this the program would not stay running
	<-ctx.Done()
}
