package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func oauth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	errCode, data := _oauth(w, r)
	re := &HttpResult{Err: errCode, Data: data}
	bytes, _ := json.Marshal(re)
	w.Write(bytes)
}

func _oauth(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	var code = r.URL.Query().Get("code")
	if code == "" {
		return 4005, "no code"
	}
	token, err := getToken(code)
	if err != nil {
		return 4006, err.Error()
	}
	id, err := getId(token)
	if err != nil {
		return 4007, err.Error()
	}
	hashId := HashMd5(strings.ToLower(id))
	_, ok := oauthValidHashes[hashId]
	if !ok {
		return 4008, id + " is not a valid user"
	}

	hostWithoutPort, _ := splitHostPort(r.Host)
	w.Header().Set("Location", uStatic)
	cookieFmt := `%s=%s; domain=%s; path=%s; max-age=%s; secure; HttpOnly; SameSite=Lax`
	// github id
	cid := fmt.Sprintf(cookieFmt, oauthCookieNamePrefix+"_id", hashId, hostWithoutPort, oauthCookiePath, oauthCookieTTL)
	w.Header().Add("Set-Cookie", cid)
	// time
	now := time.Now().Unix()
	nowStr := fmt.Sprintf("%d", now)
	ctime := fmt.Sprintf(cookieFmt, oauthCookieNamePrefix+"_t", nowStr, hostWithoutPort, oauthCookiePath, oauthCookieTTL)
	w.Header().Add("Set-Cookie", ctime)
	// verify
	raw := fmt.Sprintf("%s|%s|%s", hashId, oauthSalt, nowStr)
	hash := HashSHA1(raw)
	cvp := fmt.Sprintf(cookieFmt, oauthCookieNamePrefix+"_vp", hash, hostWithoutPort, oauthCookiePath, oauthCookieTTL)
	w.Header().Add("Set-Cookie", cvp)
	w.WriteHeader(http.StatusFound)
	return 2000, id
}

func getToken(code string) (string, error) {
	var data = fmt.Sprintf(`{
		"client_id": "%s",
		"client_secret": "%s",
		"code": "%s"
	}`, oauthClientId, oauthClientSecret, code)
	// log.Println("getToken: ", data)
	req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBufferString(data))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	// log.Println(string(body))
	result := make(map[string]json.RawMessage)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}
	token, ok := result["access_token"]
	if ok {
		return strings.Trim(string(token), `"`), nil
	} else {
		return "", fmt.Errorf("no 'access_token' in response: %s", string(body))
	}
}

func getId(token string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	// fmt.Println(string(body))
	result := make(map[string]json.RawMessage)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}
	id, ok := result["login"]
	if ok {
		return strings.Trim(string(id), `"`), nil
	} else {
		return "", fmt.Errorf("no 'login' in response: %s", string(body))
	}
}
