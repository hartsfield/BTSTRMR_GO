package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func freshView(w http.ResponseWriter, r *http.Request) {
	var page pageData
	ts := getFresh()
	page.Tracks = setLikes(r, ts)
	page.UserData = &credentials{}
	page.PageName = "LATEST TRACKS"
	exeTmpl(w, r, &page, "main.tmpl")
}

func hotView(w http.ResponseWriter, r *http.Request) {
	var page pageData
	ts := getHot()
	page.Tracks = setLikes(r, ts)
	page.UserData = &credentials{}
	page.PageName = "HOTTEST TRACKS"
	exeTmpl(w, r, &page, "main.tmpl")
}

func likesView(w http.ResponseWriter, r *http.Request) {
	name := strings.Split(r.URL.Path, "/")[2]
	var page pageData
	page.Tracks = setLikes(r, getLikes(r, name))
	page.UserData = &credentials{}
	page.PageName = name + "'s Liked Tracks"
	exeTmpl(w, r, &page, "main.tmpl")
}

func getTracks(w http.ResponseWriter, r *http.Request) {
	page, err := marshalPageData(r)
	if err != nil {
		log.Println(err)
	}

	var ts []*track
	log.Println(page.Category)
	if page.Category == "FRESH" {
		log.Println("test fresj")
		ts = getFresh()
		page.PageName = "LATEST TRACKS"
	} else if page.Category == "HOT" {
		log.Println("test hot")
		ts = getHot()
		page.PageName = "HOTTEST TRACKS"
	} else {
		ts = setLikes(r, getLikes(r, page.Category))
		page.PageName = page.Category + "'s Liked Tracks"
	}
	page.Tracks = setLikes(r, ts)
	page.UserData = &credentials{}

	var b bytes.Buffer
	err = templates.ExecuteTemplate(&b, "updateList.tmpl", page)
	if err != nil {
		fmt.Println(err)
	}
	ajaxResponse(w, map[string]string{
		"success":  "true",
		"error":    "false",
		"template": b.String(),
	})
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

		// If the track is already in the users LIKES, we remove it,
		// and decrement the score from FRESH
		if result == 0 {
			_, err := rdb.ZRem(rdbctx, a.Name+":LIKES", td.ID).Result()
			if err != nil {
				log.Print(err)
			}

			_, err = rdb.ZIncrBy(rdbctx, "HOT", -1, td.ID).Result()
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

		pipe.ZIncrBy(rdbctx, "HOT", 1, td.ID)
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
