package main

import (
	"context"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"strings"
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

var (
	// connect to redis
	redisIP = os.Getenv("redisIP")
	rdb     = redis.NewClient(&redis.Options{
		Addr:     redisIP + ":6379",
		Password: "",
		DB:       0,
	})

	// this context is used for the client/server connection. It's useful
	// for passing the token/credentials around.
	rdbctx = context.Background()
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	getTracks()
}

// makeZmem returns a redis Z member for use in a ZSET. Score is set to zero
func makeZmem(st string) *redis.Z {
	return &redis.Z{
		Member: st,
		Score:  0,
	}
}

// getTracks populates the global tracks variable with audio tracks from the
// filesystem
func getTracks() {
	for i := 0; i <= 5; i++ {
		files, _ := os.ReadDir("../public/assets/audio")
		for _, f := range files {
			t := makeTrack(f)
			// tracks = append(tracks, t)
			addToRedis(t)
		}
	}
}

// makeTrack takes a directory entry and uses the information to return a track
// object
func makeTrack(file fs.DirEntry) *track {
	fn := strings.Split(file.Name(), "-")[1]
	fn = fn[:strings.LastIndex(fn, ".")]
	return addImage(&track{
		Artist: strings.Split(file.Name(), "-")[0],
		Title:  fn,
		Path:   file.Name(),
		ID:     genPostID(5),
	})
}

// addImage adds an image to a track object by analyzing the first 6 characters
// of the image file name and audio file name and checking for a match. For this
// to work, audio and images must be added to the appropriate directory with
// intent
func addImage(t *track) *track {
	images, _ := os.ReadDir("../public/assets/images")
	for _, image := range images {
		if strings.ToLower(image.Name())[0:6] == strings.ToLower(t.Artist[0:6]) {
			t.Image = image.Name()
		}
	}
	return t
}

// genPostID generates a post ID
func genPostID(length int) (ID string) {
	symbols := "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i <= length; i++ {
		s := rand.Intn(len(symbols))
		ID += symbols[s : s+1]
	}
	return
}

func addToRedis(t *track) {
	pipe := rdb.Pipeline()
	_, err := pipe.HMSet(rdbctx, "TRACK:"+t.ID, map[string]interface{}{
		"Artist": t.Artist,
		"Title":  t.Title,
		"Image":  t.Image,
		"Path":   t.Path,
		"ID":     t.ID,
		"Likes":  fmt.Sprint(t.Likes),
		"Liked":  false,
	}).Result()
	if err != nil {
		fmt.Println(err)
	}
	pipe.ZAdd(rdbctx, "FRESH", makeZmem(t.ID))
	pipe.Exec(rdbctx)
	fmt.Println(t.Title, t.ID)
}
