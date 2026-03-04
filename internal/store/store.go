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
	mu       sync.RWMutex
	dataFile string
	data     []Instance
}

func New(dataFile string) (*Store, error) {
	s := &Store{dataFile: dataFile}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load store: %w", err)
	}
	return s, nil
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.dataFile)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(b, &s.data)
}

func (s *Store) save() error {
	s.mu.RLock()
	b, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(s.dataFile, b, 0644)
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
	for _, inst := range s.data {
		if inst.ID == id {
			return inst, true
		}
	}
	return Instance{}, false
}

func (s *Store) Create(inst Instance) (Instance, error) {
	s.mu.Lock()
	// Check duplicate port
	for _, existing := range s.data {
		if existing.SocksPort == inst.SocksPort {
			s.mu.Unlock()
			return Instance{}, fmt.Errorf("port %d already in use", inst.SocksPort)
		}
	}
	inst.ID = randID()
	inst.CreatedAt = time.Now()
	s.data = append(s.data, inst)
	s.mu.Unlock()
	return inst, s.save()
}

func (s *Store) Update(id string, updated Instance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, inst := range s.data {
		if inst.ID == id {
			updated.ID = id
			updated.CreatedAt = inst.CreatedAt
			s.data[i] = updated
			break
		}
	}
	return s.save()
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := s.data[:0]
	for _, inst := range s.data {
		if inst.ID != id {
			filtered = append(filtered, inst)
		}
	}
	s.data = filtered
	return s.save()
}

func randID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
