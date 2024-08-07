# go-multiqueue

A go-multiqueue sorts enqueued entities into an underlying SortedQueues sorted by a defined sorting property.
This way e.g. database events of different entities can be stored in separate queues and won't be run in parallel, while events of different entities can run in parallel.

A go-multiqueue is thread safe and will be locked by a mutex upon `Enqueue`, `Dequeue`, `Unblock`, `UnblockWithError` or `GetDebugContent`.
The underlying SortedQueues will be blocked upon `Dequeue` and unblocked upon `Unblock` or `UnblockWithError`.
This way an entity can be dequeued, handled by the post processor, and then get unblocked when the processing is finished.

When the last entity of a SortedQueue is dequeued the SortedQueue is held alive without queued entities.
This way a new entity can still be enqueued while the last one is being processed, without them interfering with parallel processing.
SortedQueues are only deleted if they are empty and then get unblocked.

## Usage
Import the package with `go get -u github.com/otto-rmu-technology/go-multiqueue`

```go
import (
	multiqueue "github.com/otto-rmu-technology/go-multiqueue"
)

// The test entity to be inserted
type TestEntity struct {
    ExampleSortingProperty  string
    ID                      int
}

// Initiate the Multiqueue with desired SortingProperty
// THe sorting property has to be a string
mq, err := multiqueue.NewMultiQueue("ExampleSortingProperty")
if err != nil { 
	// Handle error
}

e1 := TestEntity{
	ExampleSortingProperty: "A",
	ID: 1,
}

e2 := TestEntity{
	ExampleSortingProperty: "B",
	ID: 1,
}

mq.Enqueue(e1)
mq.Enqueue(e2)

entity1, err := mq.Dequeue()
if err != nil {
	// Handle error
}

// process entity1 here

err = mq.Unblock(entity1.DefaultSortingProperty)
if err != nil {
	// Handle error
}

// log the content / state of the MultiQueue
log.Println(fmt.Sprintf("logging current state with content = %v", mq.GetDebugContent()))

entity2, err := mq.Dequeue()
if err != nil {
// Handle error
}

// process entity2 here
// but an error happens

err = mq.UnblockWithError(entity2.DefaultSortingProperty)
if err != nil {
// Handle error
}
// entity2 is now back in the MultiQueue and it is unblocked so the failed entity can be processed again

entity2, err := mq.Dequeue()
if err != nil {
// Handle error
}

// process entity2 here
// no error happens now

err = mq.Unblock(entity2.DefaultSortingProperty)
if err != nil {
// Handle error
}

// all entities cleared from queue 
```