package model

import (
	"crypto/sha256"
	"errors"
	"flag"
	"sync"
	"time"

	"github.com/gyf1214/chaty/util"
)

type Session interface {
	Shutdown()
	Key() []byte
	GenerateKey() (string, string, error)
	Token() string
	User() User
}

const keySize = 256 / 8

var (
	sessions map[string]Session
	lock     sync.RWMutex
	expire   = flag.Int("expire", 3600, "session expire time")
)

func init() {
	sessions = make(map[string]Session)
}

type session struct {
	token string
	user  User
	timer *time.Timer
	key   []byte
}

func NewSession(token string, user User) (Session, error) {
	lock.Lock()
	defer lock.Unlock()

	ret := &session{
		token: token,
		user:  user,
	}
	if sessions[token] != nil {
		return nil, errors.New("duplicate token")
	}
	sessions[token] = ret

	ret.timer = time.AfterFunc(time.Duration(*expire)*time.Second, func() {
		ret.user.Leave()
	})
	return ret, nil
}

func FindSession(token string) Session {
	lock.RLock()
	defer lock.RUnlock()
	return sessions[token]
}

func (s *session) Shutdown() {
	lock.Lock()
	defer lock.Unlock()
	s.timer.Stop()
	delete(sessions, s.token)
}

func (s *session) Key() []byte {
	return s.key
}

func (s *session) GenerateKey() (string, string, error) {
	m, err := util.PointFromRandom()
	if err != nil {
		return "", "", err
	}
	r, err := util.RandomFromCurve()
	if err != nil {
		return "", "", err
	}
	k := s.user.PubKey()
	c1 := m.Add(k.Mul(r))
	c2, err := util.PointFromPriv(r)
	if err != nil {
		return "", "", err
	}

	key := sha256.Sum256([]byte(m.Encode()))
	s.key = key[:]
	return c1.Encode(), c2.Encode(), nil
}

func (s *session) Token() string {
	return s.token
}

func (s *session) User() User {
	return s.user
}
