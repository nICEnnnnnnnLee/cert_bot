package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func isNeedOAuth() bool {
	needAuth := oauthClientId != "" && oauthClientSecret != ""
	if needAuth {
		if oauthValidUsers == "" {
			log.Fatalln(`'OAuthValidUsers' should be set in env. 
			For multi-users, seperate github logins by ','. 
			For example, "user1,user2"`)
		}
		log.Println("OAuthSalt: ", oauthSalt)
		oauthValidHashes = make(map[string]interface{})
		for _, user := range strings.Split(oauthValidUsers, ",") {
			hash := HashMd5(strings.ToLower(user))
			log.Println("OAuth Valid User: ", user, hash)
			oauthValidHashes[hash] = nil
		}
	}
	return needAuth
}

func AuthH(h http.Handler) http.HandlerFunc {
	return AuthHF(h.ServeHTTP)
}

func AuthHF(h http.HandlerFunc) http.HandlerFunc {
	if bNeedOAuth {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "private, max-age=0, no-cache")
			w.Header().Add("Cache-Control", "private, max-age=0, must-revalidate")
			w.Header().Add("Cache-Control", "no-store")
			log.Println("check url:", r.RequestURI)
			// 检查 cookie 的id time hash
			if isValid(r) {
				h(w, r)
				return
			} else {
				http.Redirect(w, r, "https://github.com/login/oauth/authorize?client_id="+oauthClientId, http.StatusFound)
			}
		}
	} else {
		return h
	}
}

func isValid(r *http.Request) bool {
	hashId, isIdValid := checkValidId(r)
	if isIdValid {
		time, isTimeValid := checkValidTime(r)
		if isTimeValid {
			return isValidHash(r, hashId, time)
		}
	}
	return false
}
func checkValidId(r *http.Request) (*string, bool) {
	cid, err := r.Cookie(oauthCookieNamePrefix + "_id")
	if err != nil {
		return nil, false
	}
	_, ok := oauthValidHashes[cid.Value]
	if !ok {
		return nil, false
	}
	return &cid.Value, true
}

func checkValidTime(r *http.Request) (*int64, bool) {
	cTime, err := r.Cookie(oauthCookieNamePrefix + "_t")
	if err != nil {
		return nil, false
	}
	cTimeInt64, _ := strconv.ParseInt(cTime.Value, 10, 64)
	delta := time.Now().Unix() - cTimeInt64
	if delta > oauthCookieTTLInt64 || delta < -oauthCookieTTLInt64 {
		return nil, false
	}
	return &cTimeInt64, true
}

func isValidHash(r *http.Request, hashId *string, timeInt64 *int64) bool {
	hash, err := r.Cookie(oauthCookieNamePrefix + "_vp")
	if err != nil {
		return false
	}
	expectedRaw := fmt.Sprintf("%s|%s|%d", *hashId, oauthSalt, *timeInt64)
	expectedHash := HashSHA1(expectedRaw)
	// log.Println("expectedHash:", expectedHash)
	// log.Println("actualHash:", hash.Value)
	return hash.Value == expectedHash
}
