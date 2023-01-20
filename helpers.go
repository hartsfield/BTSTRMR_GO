package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func getFresh() (fresh []*track) {
	IDs, err := rdb.ZRevRange(rdbctx, "FRESH", 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	for _, ID := range IDs {
		TrackMap, _ := rdb.HGetAll(rdbctx, "TRACK:"+ID).Result()
		LikesInt, err := strconv.Atoi(TrackMap["Likes"])
		if err != nil {
			fmt.Println(err)
		}
		var isLiked bool
		if TrackMap["Liked"] == "true" {
			isLiked = true
		} else {
			isLiked = false
		}

		t := &track{
			TrackMap["Artist"],
			TrackMap["Title"],
			TrackMap["Image"],
			TrackMap["Path"],
			TrackMap["ID"],
			LikesInt,
			isLiked,
		}

		fresh = append(fresh, t)
	}
	return
}

func getHot() (hot []*track) {
	IDs, err := rdb.ZRevRangeByScore(rdbctx, "HOT", &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  -1,
	}).Result()
	if err != nil {
		fmt.Println(err)
	}
	for _, ID := range IDs {
		TrackMap, _ := rdb.HGetAll(rdbctx, "TRACK:"+ID).Result()
		LikesInt, err := strconv.Atoi(TrackMap["Likes"])
		if err != nil {
			fmt.Println(err)
		}
		var isLiked bool
		if TrackMap["Liked"] == "true" {
			isLiked = true
		} else {
			isLiked = false
		}

		t := &track{
			TrackMap["Artist"],
			TrackMap["Title"],
			TrackMap["Image"],
			TrackMap["Path"],
			TrackMap["ID"],
			LikesInt,
			isLiked,
		}

		hot = append(hot, t)
	}
	return
}

// exeTmpl is used to build and execute an html template.
func exeTmpl(w http.ResponseWriter, r *http.Request, page *pageData, tmpl string) {
	// Add the user data to the page if they're logged in.
	c := r.Context().Value(ctxkey)
	if a, ok := c.(*credentials); ok && a.IsLoggedIn {
		page.UserData = a

		err := templates.ExecuteTemplate(w, tmpl, page)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	err := templates.ExecuteTemplate(w, tmpl, page)
	if err != nil {
		fmt.Println(err)
	}
}

func getLikes(r *http.Request) (likedTracks []*track) {
	c := r.Context().Value(ctxkey)
	if a, ok := c.(*credentials); ok && a.IsLoggedIn {
		likes, err := rdb.ZRange(rdbctx, a.Name+":LIKES", 0, -1).Result()
		if err != nil {
			fmt.Println(err)
		}

		for _, trackID := range likes {
			data, _ := rdb.HGetAll(rdbctx, "TRACK:"+trackID).Result()
			i, _ := strconv.Atoi(data["Likes"])
			t := &track{
				data["Artist"],
				data["Title"],
				data["Image"],
				data["Path"],
				data["ID"],
				i,
				true,
			}
			likedTracks = append(likedTracks, t)
		}

	}
	return
}

func setLikes(r *http.Request, ts []*track) []*track {
	c := r.Context().Value(ctxkey)
	if a, ok := c.(*credentials); ok && a.IsLoggedIn {
		for _, track := range ts {
			_, err := rdb.ZScore(rdbctx, a.Name+":LIKES", track.ID).Result()
			if err != nil {
				track.Liked = false
			} else {
				track.Liked = true
			}
		}
		return ts
	}
	for _, track := range ts {
		track.Liked = false
	}
	return ts
}

// marshalCredentials is used convert a request body into a credentials{}
// struct
func marshalCredentials(r *http.Request) (*credentials, error) {
	t := &credentials{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(t)
	if err != nil {
		return t, err
	}
	return t, nil
}

func marshalTrackData(r *http.Request) (*track, error) {
	t := &track{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(t)
	if err != nil {
		return t, err
	}
	return t, nil
}

// ajaxResponse is used to respond to ajax requests with arbitrary data in the
// format of map[string]string
func ajaxResponse(w http.ResponseWriter, res map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(err)
	}
}

// checkPasswordHash compares a password to a hash and returns true if they
// match
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// renewToken renews a users token using existing claims, sets it as a cookie
// on the client, and adds it to the database.
// TODO: FIX EXPIRY
func renewToken(w http.ResponseWriter, r *http.Request, claims *credentials) (ctxx context.Context) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(hmacSampleSecret)
	if err != nil {
		fmt.Println(err)
	}

	expire := time.Now().Add(10 * time.Minute)
	cookie := http.Cookie{Name: "token", Value: ss, Path: "/", Expires: expire, MaxAge: 0}
	http.SetCookie(w, &cookie)

	rdb.Set(rdbctx, claims.Name+":token", ss, 0)
	ctxx = context.WithValue(r.Context(), ctxkey, claims)
	return
}

// newClaims creates a new set of claims using user credentials, and uses
// the claims to create a new token using renewToken()
func newClaims(w http.ResponseWriter, r *http.Request, c *credentials) (ctxx context.Context) {
	claims := credentials{
		c.Name,
		"",
		true,
		[]string{},
		0,
		jwt.StandardClaims{
			// ExpiresAt: 15000,
			// Issuer:    "test",
		},
	}

	return renewToken(w, r, &claims)
}

// hashPassword takes a password string and returns a hash
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// makeZmem returns a redis Z member for use in a ZSET. Score is set to zero
func makeZmem(st string) *redis.Z {
	return &redis.Z{
		Member: st,
		Score:  0,
	}
}

// parseToken takes a token string, checks its validity, and parses it into a
// set of credentials. If the token is invalid it returns an error
func parseToken(tokenString string) (*credentials, error) {
	var claims *credentials
	token, err := jwt.ParseWithClaims(tokenString, &credentials{}, func(token *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})
	if err != nil {
		fmt.Println(err)
		cc := credentials{IsLoggedIn: false}
		return &cc, err
	}

	if claims, ok := token.Claims.(*credentials); ok && token.Valid {
		return claims, nil
	}
	return claims, err
}
