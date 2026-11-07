package progress

import (
	"customcomplete/internal/i18n"
	"sync"
)

type Progress struct {
	Percentage  int
	Message     i18n.Key
	ArgsMessage []any
}

type Broker struct {
	mu     sync.RWMutex
	latest map[int64]Progress
	subs   map[int64]map[chan Progress]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		latest: make(map[int64]Progress),
		subs:   make(map[int64]map[chan Progress]struct{}),
	}
}

func (broker *Broker) GetLatest(id int64) (Progress, bool) {
	broker.mu.RLock()
	defer broker.mu.RUnlock()

	latest, ok := broker.latest[id]
	return latest, ok
}

func (broker *Broker) Subscribe(id int64) (updates <-chan Progress, initial Progress, hasInitial bool, unsubscribe func()) {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	if broker.subs[id] == nil {
		broker.subs[id] = make(map[chan Progress]struct{})
	}
	channel := make(chan Progress, 1)
	broker.subs[id][channel] = struct{}{}

	latest, ok := broker.latest[id]

	var once sync.Once
	unsub := func() {
		once.Do(func() {
			broker.unsubscribe(id, channel)
		})
	}

	return channel, latest, ok, unsub
}

func (broker *Broker) unsubscribe(id int64, ch chan Progress) {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	if subs, exists := broker.subs[id]; exists {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(broker.subs, id)
		}
	}
}

func (broker *Broker) Update(id int64, percentage int, message i18n.Key, args ...any) {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	prog := Progress{Percentage: percentage, Message: message, ArgsMessage: args}

	broker.latest[id] = prog

	for ch := range broker.subs[id] {
		select {
		case ch <- prog:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- prog:
			default:
			}
		}
	}
	if percentage >= 100 || message == i18n.TextStatusFailed {
		delete(broker.latest, id)
	}
}
