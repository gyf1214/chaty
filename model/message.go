package model

import (
	"encoding/json"

	"github.com/gyf1214/chaty/util"
)

type Data struct {
	User    string `json:"u"`
	Channel string `json:"c"`
	Message string `json:"m"`
}

type Encrypted struct {
	Data string `json:"d"`
	IV   string `json:"i"`
}

func (d *Data) Encrypt(key []byte) (Encrypted, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return Encrypted{}, err
	}

	encrypted, iv, err := util.Encrypt(key, raw)
	return Encrypted{
		Data: encrypted,
		IV:   iv,
	}, err
}

func (e *Encrypted) Decrypt(key []byte) (Data, error) {
	raw, err := util.Decrypt(key, e.Data, e.IV)
	if err != nil {
		return Data{}, err
	}

	var data Data
	err = json.Unmarshal(raw, &data)
	return data, err
}
