package model

import "sync"

type List interface {
	Push(Data)
	Pop() Data
	Empty() bool
}

type node struct {
	data Data
	next *node
}

type list struct {
	head, tail *node
	sync.RWMutex
}

func NewList() List {
	guard := &node{}
	return &list{
		head: guard,
		tail: guard,
	}
}

func (l *list) Empty() bool {
	l.RLock()
	defer l.RUnlock()
	return l.head.next == nil
}

func (l *list) Push(data Data) {
	l.Lock()
	defer l.Unlock()
	t := &node{data: data}
	l.tail.next = t
	l.tail = t
}

func (l *list) Pop() Data {
	l.Lock()
	defer l.Unlock()
	t := l.head.next
	if t == l.tail {
		l.tail = l.head
	}
	l.head.next = t.next
	return t.data
}
