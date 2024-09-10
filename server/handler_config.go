package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nicennnnnnnlee/cert_bot/dns01"
)

var AcmeConfigs = make(map[string]*AcmeConfig)

type NotEmptyString string

// UnmarshalJSON 实现了 json.Unmarshaler 接口，用于在解码 JSON 时检查字符串是否为空。
func (nes *NotEmptyString) UnmarshalJSON(data []byte) error {
	// 移除字符串两侧的空格
	str := string(data)
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return fmt.Errorf("字符串不能为空")
	}
	*nes = NotEmptyString(str)
	return nil
}

type AcmeConfig struct {
	Id           string              `json:"id"`
	DirectoryUrl string              `json:"directoryUrl"`
	Domains      string              `json:"domains"`
	Account      *Account            `json:"account"`
	Dns01        *dns01.DNS01Setting `json:"dns01"`
	CertPath     string              `json:"certPath"`
	KeyPath      string              `json:"keyPath"`
}

type Account struct {
	PrivateKey string `json:"privateKey"`
	Url        string `json:"url"`
}

type HttpResult struct {
	Err  int         `json:"err"`
	Data interface{} `json:"data"`
}

func getConfigs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(AcmeConfigs)
	w.Write(bytes)
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	conf := AcmeConfigs[id]
	if conf == nil {
		w.Write([]byte("{\"err\": 4000,\"msg\": \"No id matched!!!\"}"))
		return
	}
	bytes, _ := json.Marshal(conf)
	w.Write(bytes)
}

func setConfig(w http.ResponseWriter, r *http.Request) {
	errCode, data := _setConfig(w, r)
	re := &HttpResult{
		Err:  errCode,
		Data: data,
	}
	bytes, _ := json.Marshal(re)
	w.Write(bytes)
}

func _setConfig(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return 4001, fmt.Sprintf("Error reading request body: %+v", err)
	}
	defer r.Body.Close()

	var aconfig AcmeConfig
	if err := json.Unmarshal(body, &aconfig); err != nil {
		return 4002, fmt.Sprintf("Error unmarshaling request body: %+v", err)
	}
	// err = doCertReqDns01(&aconfig, w)
	// if err != nil {
	// 	return 4003, fmt.Sprintf("%+v", err)
	// } else {
	// }
	AcmeConfigs[string(aconfig.Id)] = &aconfig
	return 2000, "ok"
}
