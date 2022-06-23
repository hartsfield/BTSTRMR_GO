package main

import (
	"fmt"
	"net/http"
)

// home is the home page, which is just a list of tracks and the global
// audio player and navigation.
func home(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "home.tmpl", page{Tracks: tracks})
	if err != nil {
		fmt.Println(err)
	}
}
