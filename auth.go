package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt"
)

// credentials are user credentials and are used in the HTML templates and also
// by handlers that do authorized requests
type credentials struct {
	Name       string   `json:"username"`
	Password   string   `json:"password"`
	IsLoggedIn bool     `json:"isLoggedIn"`
	Posts      []string `json:"posts"`
	Score      uint     `json:"score"`
	jwt.StandardClaims
	// Add google login one day...
	//
	// IP         string `json:"IP"`
	// HasGoogle  bool   `json:"hasGoogle"`
	// GoogleCredentials googleCredentials `json:"gcred"`
	// GoogleToken string `json:"googleToken"`
}

///////////////////////////////////////////////////////////////////////////////
// Auth Routes ////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////

// signin signs a user in. It's a response to an XMLHttpRequest (AJAX request)
// containing the user credentials. It responds with a map[string]string that
// can be converted to JSON by the client. The client expects a boolean
// indicating success or error, and a possible error string.
func signin(w http.ResponseWriter, r *http.Request) {
	// Marshal the Credentials into a credentials struct
	c, err := marshalCredentials(r)
	if err != nil {
		fmt.Println(err)
		ajaxResponse(w, map[string]string{
			"success": "false",
			"error":   "Invalid Credentials",
		})
		return
	}

	// Get the passwords hash from the database by looking up the users
	// name
	hash, err := rdb.Get(rdbctx, c.Name).Result()
	if err != nil {
		fmt.Println(err)
		ajaxResponse(w, map[string]string{
			"success": "false",
			"error":   "User doesn't exist",
		})
		return
	}

	// Check if password matches by hashing it and comparing the hashes
	doesMatch := checkPasswordHash(c.Password, hash)
	if doesMatch {
		newClaims(w, r, c)
		ajaxResponse(w, map[string]string{
			"success": "true",
			"error":   "false",
		})
		return
	}
	ajaxResponse(w, map[string]string{"success": "false", "error": "Bad Password"})
}

// signup signs a user up. It's a response to an XMLHttpRequest (AJAX request)
// containing new user credentials. It responds with a map[string]string that
// can be converted to JSON. The client expects a boolean indicating success or
// error, and a possible error string.
func signup(w http.ResponseWriter, r *http.Request) {
	// Marshal the Credentials into a credentials struct
	c, err := marshalCredentials(r)
	if err != nil {
		fmt.Println(err)
		ajaxResponse(w, map[string]string{
			"success": "false",
			"error":   "Invalid Credentials",
		})
		return
	}

	// Make sure the username doesn't contain forbidden symbols
	match, err := regexp.MatchString("^[A-Za-z0-9]+(?:[ _-][A-Za-z0-9]+)*$", c.Name)
	if err != nil {
		fmt.Println(err)
		ajaxResponse(w, map[string]string{
			"success": "false",
			"error":   "Invalid Username",
		})
		return
	}

	// Make sure the username is longer than 3 characters and shorter than
	// 25, and the password is longer than 7.
	if match && (len(c.Name) < 25) && (len(c.Name) > 3) && (len(c.Password) > 7) {
		// Check if user already exists
		_, err = rdb.Get(rdbctx, c.Name).Result()
		if err != nil {
			// If username is unique and valid, we attempt to hash
			// the password
			hash, err := hashPassword(c.Password)
			if err != nil {
				fmt.Println(err)
				ajaxResponse(w, map[string]string{
					"success": "false",
					"error":   "Invalid Password",
				})
				return
			}

			// Add the user the the USERS set in redis. This
			// associates a score with the user that can be
			// incremented or decremented
			_, err = rdb.ZAdd(rdbctx, "USERS", makeZmem(c.Name)).Result()
			if err != nil {
				fmt.Println(err)
				ajaxResponse(w, map[string]string{
					"success": "false",
					"error":   "Error ",
				})
				return
			}

			// If the password is hashable, and we were able to add
			// the user to the redis ZSET, we store the hash in the
			// database with the username as the key and the hash
			// as the value thats returned by the key.
			_, err = rdb.Set(rdbctx, c.Name, hash, 0).Result()
			if err != nil {
				fmt.Println(err)
				ajaxResponse(w, map[string]string{
					"success": "false",
					"error":   "Error ",
				})
				return
			}

			// Set user token/credentials
			newClaims(w, r, c)

			// success response
			ajaxResponse(w, map[string]string{
				"success": "true",
				"error":   "false",
			})
			return
		}
		ajaxResponse(w, map[string]string{
			"success": "false",
			"error":   "User Exists",
		})
		return
	}
	ajaxResponse(w, map[string]string{
		"success": "false",
		"error":   "Invalid Username",
	})
}

// logout logs the user out by overwriting the token. It must first validate
// the existing token to get the username to overwrite the old token in the
// database
func logout(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		fmt.Println(err)
	}

	c, err := parseToken(token.Value)
	if err != nil {
		fmt.Println(err)
	}
	rdb.Set(rdbctx, c.Name+":token", "loggedout", 0)

	expire := time.Now()
	cookie := http.Cookie{
		Name:    "token",
		Value:   "loggedout",
		Path:    "/",
		Expires: expire,
		MaxAge:  0,
	}
	http.SetCookie(w, &cookie)

	ajaxResponse(w, map[string]string{"error": "false", "success": "true"})
}

// checkAuth parses and renews the authentication token, and adds it to the
// context. checkAuth is used as a middleware function for routes that allow or
// require authentication.
func checkAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create a generic user object that not signed in to be used
		// as a placeholder until credentials are verified.
		user := credentials{IsLoggedIn: false}
		// ctx is a user who isn't logged in
		ctx := context.WithValue(r.Context(), ctxkey, user)

		// get the "token" cookie
		token, err := r.Cookie("token")
		if err != nil {
			fmt.Println(err)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// parse the "token" cookie, making sure it's valid, and
		// obtaining user credentials if it is
		c, err := parseToken(token.Value)
		if err != nil {
			fmt.Println(err)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// check if "token" cookie matches the token stored in the
		// database
		tkn, err := rdb.Get(ctx, c.Name+":token").Result()
		if err != nil {
			fmt.Println(err)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// if the tokens match we renew the token and mark the user as
		// logged in
		if tkn == token.Value {
			c.IsLoggedIn = true
			ctxx := renewToken(w, r, c)
			next.ServeHTTP(w, r.WithContext(ctxx))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
