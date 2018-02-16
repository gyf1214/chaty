package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gyf1214/chaty/controller"
	"github.com/gyf1214/chaty/model"
	"github.com/gyf1214/chaty/util"
)

func post(url string, data interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return nil, err
	}
	return http.Post(url, "application/json", &buf)
}

const base = "http://127.0.0.1:12450"
const privKey = "6ERz8OH1V5o47ecPnBSbjqk3L2OkwZtT4OSDPfCj1Jg="
const user = "6Ik+FbfTq+SgPisFIVH7fN3TDzzvtraHL8Oqc357ZEM="
const channel = "4HrR+OvPyjKwOjsqeUY+TEGRUaNUT3ZkFK5mLTwyyPs="

func testEnter() (string, []byte) {
	fmt.Println("start enter")
	param := controller.EnterRequest{Token: user}
	resp, err := post(base+"/enter", param)
	if err != nil {
		panic(err)
	}
	fmt.Println("enter:", resp.StatusCode)
	var response controller.EnterResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}
	fmt.Println("c1:", response.C1)
	fmt.Println("c2:", response.C2)
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
	fmt.Println("key:", base64.StdEncoding.EncodeToString(key[:]))
	fmt.Println("token:", response.Token)
	return response.Token, key[:]
}

func testSend(token string, key []byte) {
	fmt.Println("start send")
	msg := model.Data{User: user, Channel: channel, Message: "hello world"}
	data, err := msg.Encrypt(key)
	if err != nil {
		panic(err)
	}
	fmt.Println("encrypted:", data)
	param := controller.SendRequest{Token: token, Message: data}
	resp, err := post(base+"/send", param)
	if err != nil {
		panic(err)
	}
	fmt.Println("send:", resp.StatusCode)
}

func testPoll(token string, key []byte) {
	fmt.Println("start poll")
	param := controller.PollRequest{Token: token}
	resp, err := post(base+"/poll", param)
	if err != nil {
		panic(err)
	}
	fmt.Println("poll:", resp.StatusCode)
	var response []model.Encrypted
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}
	fmt.Println(response)
	for _, data := range response {
		msg, err := data.Decrypt(key)
		if err == nil {
			fmt.Println(msg)
		}
	}
}

func main() {
	token, key := testEnter()
	go testPoll(token, key)
	testSend(token, key)

	time.Sleep(10 * time.Second)
}
