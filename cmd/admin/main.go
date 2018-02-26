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
	"github.com/gyf1214/chaty/model"
	"github.com/gyf1214/chaty/util"
)

var (
	base  = flag.String("base", "http://127.0.0.1:12450", "base url")
	conf  = flag.String("conf", "conf/local-user.json", "user config")
	token string
	key   []byte
)

type UserConf struct {
	Priv  string `json:"p"`
	Token string `json:"t"`
}

func post(url string, data interface{}) *http.Response {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		panic(err)
	}
	resp, err := http.Post(*base+url, "application/json", &buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(url, resp.StatusCode)
	return resp
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
	resp := post("/enter", controller.EnterRequest{Token: token})
	var response controller.EnterResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
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

func show() {
	resp := post("/admin/show", controller.AdminShowRequest{Token: token})
	var response model.Encrypted
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}
	msg, err := util.Decrypt(key, response.Data, response.IV)
	if err != nil {
		panic(err)
	}
	fmt.Println("current conf:", string(msg))
}

func encodeRequest(v interface{}) controller.AdminRequest {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	data, iv, err := util.Encrypt(key, raw)
	if err != nil {
		panic(err)
	}
	encrypted := model.Encrypted{Data: data, IV: iv}
	return controller.AdminRequest{Token: token, Msg: encrypted}
}

func addUser(token string, pub string) {
	msg := controller.AddUserRequest{Token: token, Pubkey: pub}
	post("/admin/addUser", encodeRequest(&msg))
}

func delUser(token string) {
	msg := controller.DelUserRequest{Token: token}
	post("/admin/delUser", encodeRequest(&msg))
}

func loop() {
	for {
		fmt.Print("> ")
		var cmd string
		fmt.Scan(&cmd)
		var arg1, arg2 string
		switch cmd {
		case "end":
			return
		case "show":
			show()
		case "addUser":
			fmt.Scan(&arg1, &arg2)
			addUser(arg1, arg2)
		case "delUser":
			fmt.Scan(&arg1)
			delUser(arg1)
		}
	}
}

func main() {
	flag.Parse()
	conf := loadConf()
	token, key = enter(conf.Token, conf.Priv)
	loop()
}
