package main

import (
	"errors"
	"sync"
)

func NewStat() *Stat {
	return &Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}
}

func NewStatTracker() *StatTracker {
	return &StatTracker{
		Subs: make(map[interface{}]*Stat),
	}
}

type StatTracker struct {
	Subs map[interface{}]*Stat
	mu   sync.Mutex
}

func (st *StatTracker) Subscribe(subscriber interface{}) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.Subs[subscriber] = NewStat()
}

func (st *StatTracker) Unsubscribe(subscriber interface{}) {
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.Subs, subscriber)
}

func (st *StatTracker) Pull(subscriber interface{}) (*Stat, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if stat, exists := st.Subs[subscriber]; exists {
		st.Subs[subscriber] = NewStat()
		return stat, nil
	}
	return nil, errors.New("subscriber does not exist")
}

func (st *StatTracker) Track(method, consumer string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, stat := range st.Subs {
		stat.ByConsumer[consumer]++
		stat.ByMethod[method]++
	}
}
