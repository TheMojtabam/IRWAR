package store

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Instance struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Resolver       string    `json:"resolver"`
	Domain         string    `json:"domain"`
	SocksPort      int       `json:"socks_port"`
	RestartMinutes int       `json:"restart_minutes"`
	AutoRestart    bool      `json:"auto_restart"`
	ExtraArgs      string    `json:"extra_args"`
	CreatedAt      time.Time `json:"created_at"`
}

type Store struct {
	mu   sync.RWMutex
	path string
	data []Instance
}

func New(path string) (*Store, error) {
	s := &Store{path: path}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	return s, json.Unmarshal(b, &s.data)
}

func (s *Store) save() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0644)
}

func (s *Store) List() []Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Instance, len(s.data))
	copy(out, s.data)
	return out
}

func (s *Store) Get(id string) (Instance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.data {
		if v.ID == id {
			return v, true
		}
	}
	return Instance{}, false
}

func (s *Store) Create(in Instance) (Instance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.data {
		if v.SocksPort == in.SocksPort {
			return Instance{}, fmt.Errorf("port %d already used", in.SocksPort)
		}
	}
	in.ID = randID()
	in.CreatedAt = time.Now()
	s.data = append(s.data, in)
	return in, s.save()
}

func (s *Store) Update(id string, in Instance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, v := range s.data {
		if v.ID == id {
			in.ID = id
			in.CreatedAt = v.CreatedAt
			s.data[i] = in
			return s.save()
		}
	}
	return fmt.Errorf("not found")
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := s.data[:0]
	for _, v := range s.data {
		if v.ID != id {
			n = append(n, v)
		}
	}
	s.data = n
	return s.save()
}

func randID() string {
	const c = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = c[rand.Intn(len(c))]
	}
	return string(b)
}
