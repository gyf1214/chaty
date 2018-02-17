package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"sync"
)

type channel struct {
	channels map[string]map[User]bool
	users    map[string]User
	admin    string
	sync.RWMutex
}

type channelData struct {
	Channels map[string][]string `json:"c"`
	Users    map[string]string   `json:"u"`
	Admin    string              `json:"a"`
}

var (
	c           channel
	channelPath = flag.String("channel", "conf/server.json", "channel db path")
)

func init() {
	c = channel{
		channels: make(map[string]map[User]bool),
		users:    make(map[string]User),
	}
}

func LoadChannels() error {
	file, err := os.Open(*channelPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return c.load(file)
}

func DumpChannels() (string, error) {
	buf := new(bytes.Buffer)
	err := c.dump(buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func SaveChannels() error {
	file, err := os.Create(*channelPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return c.dump(file)
}

func FindUser(token string) User {
	c.RLock()
	defer c.RUnlock()
	return c.users[token]
}

func FindChannel(sender User, token string) []User {
	c.RLock()
	defer c.RUnlock()

	users := c.channels[token]
	if users == nil {
		if user := c.users[token]; user != nil {
			if user.Token() == sender.Token() {
				return []User{sender}
			}
			return []User{sender, user}
		}
		return nil
	}

	ret := []User{}
	found := false
	for k := range users {
		ret = append(ret, k)
		if sender.Token() == k.Token() {
			found = true
		}
	}
	if !found {
		return nil
	}

	return ret
}

func AddUser(token string, key string) (User, error) {
	c.Lock()
	defer c.Unlock()

	if c.users[token] != nil {
		return nil, errors.New("user exist")
	}
	user, err := NewUser(token, key)
	if err != nil {
		return nil, err
	}

	c.users[token] = user
	return user, nil
}

func DelUser(token string) error {
	c.Lock()
	defer c.Unlock()

	user := c.users[token]
	if user == nil {
		return errors.New("user not exist")
	}
	delete(c.users, token)

	for k, users := range c.channels {
		if users[user] {
			delete(users, user)
			if len(users) == 0 {
				delete(c.channels, k)
			}
		}
	}
	return nil
}

func AddChannel(token string, user User) error {
	c.Lock()
	defer c.Unlock()

	if c.channels[token][user] {
		return errors.New("channel-user exist")
	}
	if c.channels[token] == nil {
		c.channels[token] = make(map[User]bool)
	}
	c.channels[token][user] = true

	return nil
}

func DelChannel(token string, user User) error {
	c.Lock()
	defer c.Unlock()

	if !c.channels[token][user] {
		return errors.New("channel-user not exist")
	}
	ch := c.channels[token]
	delete(ch, user)
	if len(ch) == 0 {
		delete(c.channels, token)
	}

	return nil
}

func ClearChannel(token string) error {
	c.Lock()
	defer c.Unlock()

	if c.channels[token] == nil {
		return errors.New("channel not exist")
	}

	delete(c.channels, token)
	return nil
}

func CheckAdmin(u User) bool {
	return u.Token() == c.admin
}

func (c *channel) dump(w io.Writer) error {
	c.RLock()
	defer c.RUnlock()

	data := &channelData{
		Channels: make(map[string][]string),
		Users:    make(map[string]string),
		Admin:    c.admin,
	}

	for k, users := range c.channels {
		data.Channels[k] = []string{}
		for t := range users {
			data.Channels[k] = append(data.Channels[k], t.Token())
		}
	}

	for k, u := range c.users {
		data.Users[k] = u.PubKey().Encode()
	}

	return json.NewEncoder(w).Encode(data)
}

func (c *channel) load(r io.Reader) error {
	c.Lock()
	defer c.Unlock()

	var data channelData
	err := json.NewDecoder(r).Decode(&data)
	if err != nil {
		return err
	}

	c.admin = data.Admin

	c.users = make(map[string]User)
	for k, u := range data.Users {
		c.users[k], err = NewUser(k, u)
		if err != nil {
			return err
		}
	}

	c.channels = make(map[string]map[User]bool)
	for k, v := range data.Channels {
		c.channels[k] = make(map[User]bool)
		for _, u := range v {
			user := c.users[u]
			if user == nil {
				return errors.New("bad channel data")
			}
			c.channels[k][user] = true
		}
	}

	return nil
}
