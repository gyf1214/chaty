package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gyf1214/chaty/model"
	"github.com/gyf1214/chaty/util"
)

type AdminShowRequest struct {
	Token string `json:"t"`
}

type AdminRequest struct {
	Token string          `json:"t"`
	Msg   model.Encrypted `json:"m"`
}

type AddUserRequest struct {
	Token  string `json:"t"`
	Pubkey string `json:"k"`
}

type DelUserRequest struct {
	Token string `json:"t"`
}

func checkAdmin(token string) model.Session {
	if session := model.FindSession(token); session != nil {
		if model.CheckAdmin(session.User()) {
			return session
		}
		return nil
	}
	return nil
}

func show(w http.ResponseWriter, r *http.Request) {
	var par AdminShowRequest
	err := json.NewDecoder(r.Body).Decode(&par)
	if err != nil {
		http.Error(w, "", 400)
		return
	}

	if session := checkAdmin(par.Token); session != nil {
		key := session.Key()
		raw, err := model.DumpChannels()
		if err != nil {
			http.Error(w, "", 500)
			return
		}
		encrypted, iv, err := util.Encrypt(key, []byte(raw))
		if err != nil {
			http.Error(w, "", 500)
			return
		}
		ret := model.Encrypted{Data: encrypted, IV: iv}
		json.NewEncoder(w).Encode(ret)
	} else {
		http.Error(w, "", 403)
	}
}

func parseRequest(r io.Reader, v interface{}) error {
	var par AdminRequest
	err := json.NewDecoder(r).Decode(&par)
	if err != nil {
		return err
	}

	if session := checkAdmin(par.Token); session != nil {
		key := session.Key()
		raw, err := util.Decrypt(key, par.Msg.Data, par.Msg.IV)
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, v)
	}
	return errors.New("forbidden")
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var par AddUserRequest
	err := parseRequest(r.Body, &par)
	if err != nil {
		http.Error(w, "", 400)
		return
	}

	_, err = model.AddUser(par.Token, par.Pubkey)
	if err != nil {
		http.Error(w, "", 500)
		return
	}
	err = model.SaveChannels()
	if err != nil {
		http.Error(w, "", 500)
	}
}

func delUser(w http.ResponseWriter, r *http.Request) {
	var par DelUserRequest
	err := parseRequest(r.Body, &par)
	if err != nil {
		http.Error(w, "", 400)
		return
	}

	err = model.DelUser(par.Token)
	if err != nil {
		http.Error(w, "", 500)
		return
	}
	err = model.SaveChannels()
	if err != nil {
		http.Error(w, "", 500)
	}
}

func init() {
	http.HandleFunc("/admin/show", show)
	http.HandleFunc("/admin/addUser", addUser)
	http.HandleFunc("/admin/delUser", delUser)
}
