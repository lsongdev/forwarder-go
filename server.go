package main

import (
	"fmt"
	"sync/atomic"
)

type Server struct {
	Forwarders map[string]*Forwarder
}

func NewServer() *Server {
	return &Server{
		Forwarders: make(map[string]*Forwarder),
	}
}

func (s *Server) AddForwarder(from, to string) error {
	if _, exists := s.Forwarders[from]; exists {
		return fmt.Errorf("port %s is already in use", from)
	}

	f := &Forwarder{From: from, To: to}
	err := f.Start()
	if err != nil {
		return err
	}

	s.Forwarders[from] = f
	return nil
}

func (s *Server) RemoveForwarder(from string) {
	if f, exists := s.Forwarders[from]; exists {
		f.Stop()
		delete(s.Forwarders, from)
	}
}

func (s *Server) GetStats() []map[string]interface{} {
	stats := make([]map[string]interface{}, 0, len(s.Forwarders))
	for _, f := range s.Forwarders {
		stats = append(stats, map[string]interface{}{
			"from":        f.From,
			"to":          f.To,
			"upload":      atomic.LoadUint64(&f.BytesUploaded),
			"download":    atomic.LoadUint64(&f.BytesDownloaded),
			"connections": atomic.LoadUint64(&f.Connections),
		})
	}
	return stats
}
