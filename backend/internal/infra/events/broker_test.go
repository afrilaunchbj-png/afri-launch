package events

import (
	"encoding/json"
	"testing"

	"afrilaunch/backend/internal/application/port"
)

func TestBrokerSubscribeReceive(t *testing.T) {
	b := NewBroker()
	ch, cancel := b.Subscribe("user-1")
	defer cancel()

	raw, _ := json.Marshal(map[string]string{"hello": "world"})
	b.Publish("user-1", port.AppEvent{Type: port.EventChatDelta, Data: raw})

	select {
	case evt, ok := <-ch:
		if !ok {
			t.Fatal("channel fermé inattendu")
		}
		if evt.Type != port.EventChatDelta || string(evt.Data) != string(raw) {
			t.Fatalf("événement inattendu : %+v", evt)
		}
	default:
		t.Fatal("aucun événement reçu")
	}
}

func TestBrokerIsolation(t *testing.T) {
	b := NewBroker()
	ch1, cancel1 := b.Subscribe("user-1")
	defer cancel1()
	ch2, cancel2 := b.Subscribe("user-2")
	defer cancel2()

	b.Publish("user-1", port.AppEvent{Type: port.EventChatDelta, Data: []byte(`{}`)})

	select {
	case <-ch1:
	default:
		t.Fatal("user-1 aurait dû recevoir l'événement")
	}
	select {
	case evt := <-ch2:
		t.Fatalf("user-2 n'aurait pas dû recevoir : %+v", evt)
	default:
	}
}

func TestBrokerMultipleSubscribers(t *testing.T) {
	b := NewBroker()
	ch1, cancel1 := b.Subscribe("user-1")
	defer cancel1()
	ch2, cancel2 := b.Subscribe("user-1")
	defer cancel2()

	b.Publish("user-1", port.AppEvent{Type: port.EventChatCompleted, Data: []byte(`{}`)})

	for _, ch := range []<-chan port.AppEvent{ch1, ch2} {
		select {
		case <-ch:
		default:
			t.Fatal("chaque abonné devrait recevoir l'événement")
		}
	}
}

func TestBrokerCancelClosesChannel(t *testing.T) {
	b := NewBroker()
	ch, cancel := b.Subscribe("user-1")
	cancel()

	if _, ok := <-ch; ok {
		t.Fatal("le channel devrait être fermé après cancel")
	}
	cancel() // idempotent
}

func TestBrokerSlowSubscriberDisconnected(t *testing.T) {
	b := NewBroker()
	ch, cancel := b.Subscribe("user-1")
	defer func() { _ = recover() }()

	for i := 0; i < subscriberBuffer+10; i++ {
		b.Publish("user-1", port.AppEvent{Type: port.EventChatDelta, Data: []byte(`{}`)})
	}

	// Les événements bufferisés restent lisibles, puis le channel est fermé.
	for i := 0; i < subscriberBuffer; i++ {
		if _, ok := <-ch; !ok {
			t.Fatalf("l'événement bufferisé %d aurait dû être reçu", i)
		}
	}
	if _, ok := <-ch; ok {
		t.Fatal("l'abonné lent devrait être déconnecté (channel fermé)")
	}
	cancel()
}
