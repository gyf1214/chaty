package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gyf1214/chaty/model"
)

type EnterRequest struct {
	Token string `json:"t"`
}

type EnterResponse struct {
	Token string `json:"t"`
	C1    string `json:"c1"`
	C2    string `json:"c2"`
}

func enter(w http.ResponseWriter, r *http.Request) {
	var par EnterRequest
	err := json.NewDecoder(r.Body).Decode(&par)
	if err != nil {
		http.Error(w, "", 400)
		return
	}

	user := model.FindUser(par.Token)
	if user == nil {
		http.Error(w, "", 400)
		return
	}

	session, err := user.Enter()
	if err != nil {
		http.Error(w, "", 400)
		return
	}

	c1, c2, err := session.GenerateKey()
	if err != nil {
		http.Error(w, "", 400)
		return
	}

	resp := EnterResponse{Token: session.Token(), C1: c1, C2: c2}
	json.NewEncoder(w).Encode(&resp)
}

func init() {
	http.HandleFunc("/enter", enter)
}
