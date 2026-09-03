// Package events implémente le hub temps réel in-process (SSE unique).
// Un seul canal server→client par utilisateur : deltas du chat, statuts de
// jobs, notifications futures. Multi-instance : à migrer sur Redis pub/sub
// lors du passage à asynq (l'interface port.EventBus est conservée).
package events

import (
	"sync"
	"sync/atomic"

	"afrilaunch/backend/internal/application/port"
)

// subscriberBuffer est la taille du buffer par connexion. Au-delà
// (consommateur trop lent), la connexion est fermée : le client se
// reconnecte et resynchronise (les messages du chat sont persistés).
const subscriberBuffer = 128

type subscriber struct {
	ch     chan port.AppEvent
	closed atomic.Bool
}

// Broker distribue les événements aux connexions SSE par utilisateur.
type Broker struct {
	mu   sync.RWMutex
	subs map[string]map[*subscriber]struct{}
}

// NewBroker construit le hub in-process.
func NewBroker() *Broker {
	return &Broker{subs: make(map[string]map[*subscriber]struct{})}
}

// Subscribe enregistre une connexion pour un utilisateur. cancel désabonne
// (idempotent) et ferme le channel retourné.
func (b *Broker) Subscribe(userID string) (<-chan port.AppEvent, func()) {
	sub := &subscriber{ch: make(chan port.AppEvent, subscriberBuffer)}

	b.mu.Lock()
	if b.subs[userID] == nil {
		b.subs[userID] = make(map[*subscriber]struct{})
	}
	b.subs[userID][sub] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			b.detach(userID, sub)
		})
	}
	return sub.ch, cancel
}

// Publish diffuse un événement à toutes les connexions de l'utilisateur
// (non bloquant). Un abonné dont le buffer est plein est déconnecté.
func (b *Broker) Publish(userID string, evt port.AppEvent) {
	b.mu.RLock()
	subs := make([]*subscriber, 0, len(b.subs[userID]))
	for sub := range b.subs[userID] {
		subs = append(subs, sub)
	}
	b.mu.RUnlock()

	for _, sub := range subs {
		if sub.closed.Load() {
			continue
		}
		select {
		case sub.ch <- evt:
		default:
			b.detach(userID, sub)
		}
	}
}

func (b *Broker) detach(userID string, sub *subscriber) {
	b.mu.Lock()
	if set, ok := b.subs[userID]; ok {
		delete(set, sub)
		if len(set) == 0 {
			delete(b.subs, userID)
		}
	}
	b.mu.Unlock()
	if sub.closed.CompareAndSwap(false, true) {
		close(sub.ch)
	}
}
