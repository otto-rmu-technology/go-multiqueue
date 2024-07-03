package go_multiqueue

import (
	"errors"
	"reflect"
	"sync"
	"time"
)

var (
	ErrEmptySortingPropertyValueError            = errors.New("error: given entity has empty sorting property value")
	ErrEntityNotAStructError                     = errors.New("error: given entity is not a struct")
	ErrSortingPropertyNotExistOrNotExportedError = errors.New("error: given entity doesn't contain required sorting property or not exported")
	ErrSortingPropertyNotAStringError            = errors.New("error: given entity is not a string")
	ErrNoUnblockedQueueFoundError                = errors.New("error: no queue was found which is unblocked and contains entities")
	ErrNoEntitiesInQueueError                    = errors.New("error: no entities were found in queue")
	ErrSortedQueueAlreadyBlockedError            = errors.New("error: sorted queue is already blocked")
	ErrSortedQueueAlreadyUnblockedError          = errors.New("error: sorted queue is already unblocked")
	ErrSortedQueueNotFoundError                  = errors.New("error: sorted queue not exist")
	ErrEmptySortingPropertyNameError             = errors.New("error: sorting property name is empty")
)

type MultiQueue struct {
	sortingProperty string
	sortedQueues    map[string]*SortedQueue
	mu              sync.Mutex
}

type SortedQueue struct {
	entities         []SortedQueueEntity
	oldestEntityTime time.Time
	isBlocked        bool
}

type SortedQueueEntity struct {
	entity        any
	insertionTime time.Time
}

func NewSortedQueue() *SortedQueue {
	return &SortedQueue{
		entities:         []SortedQueueEntity{},
		oldestEntityTime: time.Time{},
		isBlocked:        false,
	}
}

func NewMultiQueue(sortingProperty string) (*MultiQueue, error) {
	if sortingProperty == "" {
		return nil, ErrEmptySortingPropertyNameError
	}

	return &MultiQueue{
		sortingProperty: sortingProperty,
		sortedQueues:    map[string]*SortedQueue{},
	}, nil
}

// Dequeue dequeues an entity from the SortedQueue with the oldest entity.
// The SortedQueue will only be altered when unblocking.
func (m *MultiQueue) Dequeue() (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sortedQueue, err := m.getNonBlockedSortedQueueWithOldestEntity()
	if err != nil {
		return nil, err
	}

	entity, err := sortedQueue.getNextEntity()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// Enqueue checks if the sorting property is present and enqueues the entity to the correct SortedQueue.
func (m *MultiQueue) Enqueue(e any) error {
	ts := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	sortingPropertyValue, err := m.getSortingPropertyValue(e)
	if err != nil {
		return err
	}

	sortedQueue, err := m.getSortedQueue(sortingPropertyValue)
	if err != nil {
		return err
	}
	sortedQueue.enqueue(e, ts)

	return nil
}

// Unblock should be called when an entity was dequeued, otherwise the SortedQueue will be blocked forever.
// Unblocks a SortingQueue after the event for the entity was successfully handled.
// If the sorted queue is empty when trying to unblock it will get deleted.
func (m *MultiQueue) Unblock(sortingPropertyValue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.sortedQueues[sortingPropertyValue]
	if !ok {
		return ErrSortedQueueNotFoundError
	}
	if !m.sortedQueues[sortingPropertyValue].isBlocked {
		return ErrSortedQueueAlreadyUnblockedError
	}

	// delete the last entity from queue
	err := m.sortedQueues[sortingPropertyValue].dequeue()
	if err != nil {
		return err
	}

	// delete if no entities are there anymore
	if len(m.sortedQueues[sortingPropertyValue].entities) <= 0 {
		delete(m.sortedQueues, sortingPropertyValue)
		return nil
	}

	err = m.sortedQueues[sortingPropertyValue].unblock()
	if err != nil {
		return err
	}

	return nil
}

// UnblockWithError should be called when an entity was dequeued but an error happened while processing otherwise the SortedQueue will be blocked forever.
// This way the entity will be pushed back to the front of the queue, so that the event can be handled again.
// This is not optimal, but it keeps the integrity of the events.
// Worst case is, that the event get tried to handle over and over again and always fails, still this gives you time to fix the problem while the sorted queue is in an endless loop.
// This potentially blocks the whole MultiQueue if only one entity is dequeued and processed at a time.
func (m *MultiQueue) UnblockWithError(sortingPropertyValue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.sortedQueues[sortingPropertyValue]
	if !ok {
		return ErrSortedQueueNotFoundError
	}
	if !m.sortedQueues[sortingPropertyValue].isBlocked {
		return ErrSortedQueueAlreadyUnblockedError
	}

	// there is nothing dequeued or deleted here, since the entity should remain in the queue to be dequeued again

	err := m.sortedQueues[sortingPropertyValue].unblock()
	if err != nil {
		return err
	}

	return nil
}

// getSortingPropertyValue checks if the required property to sort the entity exists.
// If yes the value is returned else ErrNoSortingPropertyError is returned.
func (m *MultiQueue) getSortingPropertyValue(e any) (string, error) {
	value := reflect.ValueOf(e)
	// check if entity is a struct
	if value.Kind() != reflect.Struct {
		return "", ErrEntityNotAStructError
	}

	// get sorting property by name
	fieldValue := value.FieldByName(m.sortingProperty)

	// check if the field exists and is exported
	if !fieldValue.IsValid() || !fieldValue.CanInterface() {
		return "", ErrSortingPropertyNotExistOrNotExportedError
	}

	// check if sorting property is a string
	if fieldValue.Kind() != reflect.String {
		return "", ErrSortingPropertyNotAStringError
	}

	// check if sorting property value is an empty string
	sortingPropertyValue := fieldValue.String()
	if sortingPropertyValue == "" {
		return "", ErrEmptySortingPropertyValueError
	}

	return fieldValue.String(), nil
}

// getSortedQueue gets the SortedQueue for the given sortingPropertyValue for enqueue purposes.
// If no SortedQueue exists a new is created and returned.
func (m *MultiQueue) getSortedQueue(sortingPropertyValue string) (*SortedQueue, error) {
	if sortingPropertyValue == "" {
		return nil, ErrEmptySortingPropertyValueError
	}
	sortedQueue, ok := m.sortedQueues[sortingPropertyValue]
	// If the sortedQueue not exists
	if !ok {
		m.sortedQueues[sortingPropertyValue] = NewSortedQueue()
		sortedQueue = m.sortedQueues[sortingPropertyValue]
	}

	return sortedQueue, nil
}

// getNonBlockedSortedQueueWithOldestEntity gets the SortedQueue with the oldest entity where the SortedQueue is not blocked.
// This is for dequeue purposes.
func (m *MultiQueue) getNonBlockedSortedQueueWithOldestEntity() (*SortedQueue, error) {
	oldestTimestamp := time.Unix(0, 0)
	oldestSortedQueueKey := ""

	for sortedQueueKey, _ := range m.sortedQueues {
		if m.sortedQueues[sortedQueueKey].oldestEntityTime.After(oldestTimestamp) && !oldestTimestamp.Equal(time.Unix(0, 0)) {
			continue
		}
		if m.sortedQueues[sortedQueueKey].isBlocked {
			continue
		}
		oldestTimestamp = m.sortedQueues[sortedQueueKey].oldestEntityTime
		oldestSortedQueueKey = sortedQueueKey
	}
	if oldestSortedQueueKey == "" {
		return nil, ErrNoUnblockedQueueFoundError
	}
	err := m.sortedQueues[oldestSortedQueueKey].block()
	if err != nil {
		return nil, err
	}
	return m.sortedQueues[oldestSortedQueueKey], nil
}

// enqueue enqueues an entity to a SortedQueue.
func (s *SortedQueue) enqueue(e any, ts time.Time) {
	sortedQueueEntity := SortedQueueEntity{
		entity:        e,
		insertionTime: ts,
	}

	if len(s.entities) == 0 {
		s.oldestEntityTime = ts
	}

	s.entities = append(s.entities, sortedQueueEntity)
}

// getNextEntity gets the next entity of the queue without deleting it from the queue.
func (s *SortedQueue) getNextEntity() (any, error) {
	if len(s.entities) == 0 {
		return nil, ErrNoEntitiesInQueueError
	}
	sortedQueueEntity := s.entities[0]

	return sortedQueueEntity.entity, nil
}

// dequeue deletes the next entity from a SortedQueue.
// The reading of the element, which is deleted from the queue here, happens in getNextEntity.
func (s *SortedQueue) dequeue() error {
	if len(s.entities) == 0 {
		return ErrNoEntitiesInQueueError
	}
	s.entities = s.entities[1:]

	if len(s.entities) == 0 {
		return nil
	}

	s.oldestEntityTime = s.entities[0].insertionTime
	return nil
}

// block blocks a SortedQueue.
func (s *SortedQueue) block() error {
	if s.isBlocked {
		return ErrSortedQueueAlreadyBlockedError
	}
	s.isBlocked = true
	return nil
}

// unblock unblocks a SortedQueue.
func (s *SortedQueue) unblock() error {
	if !s.isBlocked {
		return ErrSortedQueueAlreadyUnblockedError
	}
	s.isBlocked = false
	return nil
}
