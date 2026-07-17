package nats

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type WSMessage struct {
	Subject string `json:"subject"`
	Data    string `json:"data"`
}

type Service struct {
	nc        *nats.Conn
	subs      map[string]*nats.Subscription
	onMessage func(subject string, payload []byte)
	mu        sync.Mutex
}

func NewService(url string, onMessage func(subject string, payload []byte)) (*Service, error) {
	var nc *nats.Conn
	var err error

	maxRetries := 10
	for i := 1; i <= maxRetries; i++ {
		nc, err = nats.Connect(url)
		if err == nil {
			break
		}
		log.Printf("failed to connect to NATS (attempt %d/%d): %v. retrying in 2 seconds...", i, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS after %d attempts: %w", maxRetries, err)
	}

	return &Service{
		nc:        nc,
		subs:      make(map[string]*nats.Subscription),
		onMessage: onMessage,
	}, nil
}

func (s *Service) Subscribe(subject string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.subs[subject]; exists {
		return nil
	}

	sub, err := s.nc.Subscribe(subject, func(msg *nats.Msg) {
		wsMsg := WSMessage{
			Subject: msg.Subject,
			Data:    string(msg.Data),
		}
		payload, err := json.Marshal(wsMsg)
		if err != nil {
			log.Println("failed to marshal message:", err)
			return
		}
		s.onMessage(msg.Subject, payload)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", subject, err)
	}

	s.subs[subject] = sub
	log.Println("NATS subscription created for topic:", subject)
	return nil
}

func (s *Service) Unsubscribe(subject string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sub, exists := s.subs[subject]; exists {
		err := sub.Unsubscribe()
		if err != nil {
			log.Printf("failed to unsubscribe from NATS topic %s: %v", subject, err)
		}
		delete(s.subs, subject)
		log.Println("NATS subscription canceled for topic:", subject)
	}
}

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for subject, sub := range s.subs {
		err := sub.Unsubscribe()
		if err != nil {
			log.Printf("failed to unsubscribe from NATS topic %s during shutdown: %v", subject, err)
		}
		delete(s.subs, subject)
	}

	if s.nc != nil {
		s.nc.Close()
	}
}
