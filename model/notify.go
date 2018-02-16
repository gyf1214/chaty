package model

import (
	"flag"
	"sync"
	"time"
)

type Notifier interface {
	Acquire()
	Release()
	Wait() bool
	Notify(bool)
}

type notifier struct {
	wait   bool
	notify chan bool
	sync.Mutex
}

var timeout = flag.Int("timeout", 5000, "poll timeout")

func NewNotifier() Notifier {
	return &notifier{
		notify: make(chan bool),
	}
}

func (n *notifier) Acquire() {
	n.Lock()
	defer n.Unlock()
	if n.wait {
		n.notify <- false
	}
	n.wait = true
}

func (n *notifier) Release() {
	n.Lock()
	defer n.Unlock()
	n.wait = false
}

func (n *notifier) Wait() bool {
	defer n.Release()
	select {
	case data := <-n.notify:
		return data
	case <-time.After(time.Duration(*timeout) * time.Millisecond):
	}
	return true
}

func (n *notifier) Notify(tf bool) {
	n.Lock()
	defer n.Unlock()
	if n.wait {
		n.notify <- tf
	}
}
