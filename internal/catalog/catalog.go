package catalog

import (
	"customcomplete/internal/progress"
	"customcomplete/internal/store"
	"errors"
	"log"
)

const (
	MinWordsContents int = 2
)

var (
	ErrQueueFull error = errors.New("Processing-queue full")
)

type Service struct {
	store  textStore
	broker *progress.Broker
	jobs   chan<- int64
}

func NewService(readWrite store.ReadWrite, read store.Read, broker *progress.Broker, jobs chan<- int64) *Service {
	return &Service{
		store: textStore{
			readWrite: readWrite,
			read:      read,
		},
		broker: broker,
		jobs:   jobs,
	}
}

//---unexported

func (service *Service) enqueue(srcID sourceID) error {
	select {
	case service.jobs <- int64(srcID):
		log.Printf("Job %d queued", srcID)
		return nil
	default:
		return ErrQueueFull
	}
}
