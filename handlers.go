package main

import (
	"fmt"
	"log"
	"net/http"
)

// home is the home page, which is just a list of tracks and the global
// audio player and navigation.
func home(w http.ResponseWriter, r *http.Request) {
	page := makePage(r)
	exeTmpl(w, r, page, "home.tmpl")
}

func likeTrack(w http.ResponseWriter, r *http.Request) {
	td, err := marshalTrackData(r)
	if err != nil {
		fmt.Println(err)
		ajaxResponse(w, map[string]string{
			"success": "false",
			"error":   "Error parsing data",
		})
		return
	}
	// Check if the user is logged in. You can't like a post without being
	// logged in
	c := r.Context().Value(ctxkey)
	if a, ok := c.(*credentials); ok && a.IsLoggedIn {
		zmem := makeZmem(td.ID)

		pipe := rdb.Pipeline()
		result, err := rdb.ZAdd(rdbctx, a.Name+":LIKES", zmem).Result()
		if err != nil {
			fmt.Println(err)
		}

		if result == 0 {
			_, err := rdb.ZRem(rdbctx, a.Name+":LIKES", td.ID).Result()
			if err != nil {
				log.Print(err)
			}

			ajaxResponse(w, map[string]string{
				"success": "true",
				"isLiked": "false",
				"error":   "false",
			})
			return
		}

		pipe.ZIncrBy(rdbctx, "FRESH", 1, td.ID)
		_, err = pipe.Exec(rdbctx)
		if err != nil {
			fmt.Println(err)
			ajaxResponse(w, map[string]string{
				"success": "false",
				"isLiked": "",
				"error":   "Error updating database",
			})
			return

		}

		ajaxResponse(w, map[string]string{
			"success": "true",
			"isLiked": "true",
			"error":   "false",
		})
	}
}
