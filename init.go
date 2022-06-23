package main

import (
	"io/fs"
	"math/rand"
	"os"
	"strings"
)

// init runs before main and runs a function that primes our program with data
// from the filesystem
func init() {
	getTracks()
}

// getTracks populates the global tracks variable with audio tracks from the
// filesystem
func getTracks() {
	files, _ := os.ReadDir("./public/assets/audio")
	for _, f := range files {
		t := makeTrack(f)
		tracks = append(tracks, t)
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
		ID:     genPostID(10),
	})
}

// addImage adds an image to a track object by analyzing the first 6 characters
// of the image file name and audio file name and checking for a match. For this
// to work, audio and images must be added to the appropriate directory with
// intent
func addImage(t *track) *track {
	images, _ := os.ReadDir("./public/assets/images")
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
