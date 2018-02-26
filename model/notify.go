package model

import (
	"flag"
	"time"

	"github.com/gyf1214/chaty/util"
)

type Notifier interface {
	Acquire()
	Wait() bool
	Notify(bool)
}

type notifier struct {
	notify chan bool
	util.SpinMutex
}

var timeout = flag.Int("timeout", 5000, "poll timeout")

func NewNotifier() Notifier {
	return &notifier{
		notify: make(chan bool),
	}
}

func (n *notifier) Acquire() {
	if !n.TryLock() {
		n.notify <- false
		n.Lock()
	}
}

func (n *notifier) Wait() bool {
	defer n.Unlock()
	select {
	case data := <-n.notify:
		return data
	case <-time.After(time.Duration(*timeout) * time.Millisecond):
	}
	return true
}

func (n *notifier) Notify(tf bool) {
	if !n.TryLock() {
		n.notify <- tf
	} else {
		n.Unlock()
	}
}
