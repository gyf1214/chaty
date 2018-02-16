package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gyf1214/chaty/controller"
	"github.com/gyf1214/chaty/util"
)

var (
	base = flag.String("base", "http://127.0.0.1:12450", "base url")
	conf = flag.String("conf", "conf/local-user.json", "user config")
)

type UserConf struct {
	Priv  string `json:"p"`
	Token string `json:"t"`
}

func post(url string, data interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return nil, err
	}
	return http.Post(*base+url, "application/json", &buf)
}

func loadConf() UserConf {
	file, err := os.Open(*conf)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var ret UserConf
	err = json.NewDecoder(file).Decode(&ret)
	if err != nil {
		panic(err)
	}
	return ret
}

func enter(token string, privKey string) (string, []byte) {
	resp, err := post("/enter", controller.EnterRequest{Token: token})
	if err != nil {
		panic(err)
	}
	fmt.Println("enter:", resp.StatusCode)
	var response controller.EnterResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}
	// fmt.Println("c1:", response.C1)
	// fmt.Println("c2:", response.C2)
	c1, err := util.PointFromDecode(response.C1)
	if err != nil {
		panic(err)
	}
	c2, err := util.PointFromDecode(response.C2)
	if err != nil {
		panic(err)
	}
	priv, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		panic(err)
	}
	mm := c1.Add(c2.Mul(priv).Neg())
	key := sha256.Sum256([]byte(mm.Encode()))
	// fmt.Println("key:", base64.StdEncoding.EncodeToString(key[:]))
	fmt.Println("token:", response.Token)
	return response.Token, key[:]
}

func show(token string, key []byte) {
	resp, err := post("/admin/show", controller.AdminShowRequest{Token: token})
	if err != nil {
		panic(err)
	}
	fmt.Println("show:", resp.StatusCode)
	var response controller.AdminShowResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}
	msg, err := util.Decrypt(key, response.Data, response.IV)
	if err != nil {
		panic(err)
	}
	fmt.Println("current conf:", string(msg))
}

func loop(token string, key []byte) {
	for {
		fmt.Print("> ")
		var cmd string
		fmt.Scan(&cmd)
		switch cmd {
		case "end":
			return
		case "show":
			show(token, key)
		}
	}
}

func main() {
	conf := loadConf()
	token, key := enter(conf.Token, conf.Priv)
	loop(token, key)
}
