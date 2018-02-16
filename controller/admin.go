package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gyf1214/chaty/model"
	"github.com/gyf1214/chaty/util"
)

type AdminShowRequest struct {
	Token string `json:"t"`
}

type AdminShowResponse struct {
	Data string `json:"d"`
	IV   string `json:"i"`
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
		ret := AdminShowResponse{Data: encrypted, IV: iv}
		json.NewEncoder(w).Encode(ret)
	} else {
		http.Error(w, "", 403)
	}
}

func init() {
	http.HandleFunc("/admin/show", show)
}
