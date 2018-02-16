package model

import (
	"sync"

	"github.com/gyf1214/chaty/util"
)

type User interface {
	Send(Data)
	Poll() []Encrypted
	Token() string
	PubKey() util.Point
	Enter() (Session, error)
	Leave()
}

type user struct {
	token    string
	pubkey   util.Point
	session  Session
	queue    List
	notifier Notifier
	sync.Mutex
}

func NewUser(token string, key string) (User, error) {
	p, err := util.PointFromDecode(key)
	if err != nil {
		return nil, err
	}

	return &user{
		token:    token,
		pubkey:   p,
		queue:    NewList(),
		notifier: NewNotifier(),
	}, nil
}

func (u *user) Send(data Data) {
	u.queue.Push(data)
	u.notifier.Notify(true)
}

func (u *user) Poll() []Encrypted {
	if u.session == nil {
		return nil
	}

	u.notifier.Acquire()
	u.Lock()
	defer u.Unlock()

	if u.queue.Empty() {
		if !u.notifier.Wait() {
			return nil
		}
	}

	key := u.session.Key()
	ret := []Encrypted{}
	for !u.queue.Empty() {
		now := u.queue.Pop()
		encrypted, err := now.Encrypt(key)
		if err == nil {
			ret = append(ret, encrypted)
		}
	}

	return ret
}

func (u *user) Token() string {
	return u.token
}

func (u *user) PubKey() util.Point {
	return u.pubkey
}

func (u *user) Enter() (Session, error) {
	u.notifier.Notify(false)
	u.Lock()
	defer u.Unlock()

	if u.session != nil {
		u.session.Shutdown()
		u.session = nil
	}

	secret, err := util.RandomString(keySize)
	if err != nil {
		return nil, err
	}
	secret = util.SHA256(u.token + "-" + secret)
	session, err := NewSession(secret, u)
	if err != nil {
		return nil, err
	}
	u.session = session
	return session, nil
}

func (u *user) Leave() {
	u.notifier.Notify(false)
	u.Lock()
	defer u.Unlock()
	u.session = nil
}
