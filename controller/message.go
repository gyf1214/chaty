package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gyf1214/chaty/model"
)

type PollRequest struct {
	Token string `json:"t"`
}

func poll(w http.ResponseWriter, r *http.Request) {
	var par PollRequest
	err := json.NewDecoder(r.Body).Decode(&par)
	if err != nil {
		http.Error(w, "", 400)
		return
	}

	token := par.Token

	if session := model.FindSession(token); session != nil {
		ret := session.User().Poll()
		if ret == nil {
			http.Error(w, "", 403)
		} else {
			json.NewEncoder(w).Encode(ret)
		}
	} else {
		http.Error(w, "", 400)
	}
}

type SendRequest struct {
	Token   string          `json:"t"`
	Message model.Encrypted `json:"m"`
}

func send(w http.ResponseWriter, r *http.Request) {
	var par SendRequest
	err := json.NewDecoder(r.Body).Decode(&par)
	if err != nil {
		http.Error(w, "", 400)
		return
	}

	session := model.FindSession(par.Token)
	if session == nil {
		http.Error(w, "", 400)
		return
	}

	msg, err := par.Message.Decrypt(session.Key())
	if err != nil {
		http.Error(w, "", 400)
		return
	}
	sender := session.User()
	if msg.User != sender.Token() {
		http.Error(w, "", 400)
	}

	users := model.FindChannel(sender, msg.Channel)
	if users == nil {
		http.Error(w, "", 400)
		return
	}

	for _, user := range users {
		user.Send(msg)
	}
}

func init() {
	http.HandleFunc("/poll", poll)
	http.HandleFunc("/send", send)
}
