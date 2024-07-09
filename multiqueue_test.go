package go_multiqueue

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"sync"
	"testing"
	"time"
)

type TestEntity struct {
	SortingProperty string
	ID              int
}

func TestMultiqueue_Integration(t *testing.T) {
	testVars := struct {
		sortingProperty string
	}{
		sortingProperty: "SortingProperty",
	}

	mq, err := NewMultiQueue(testVars.sortingProperty)
	if err != nil {
		t.Errorf("Integration test failed with unexpected error = %v", err)
		return
	}

	eA1 := TestEntity{
		SortingProperty: "A",
		ID:              1,
	}
	eA2 := TestEntity{
		SortingProperty: "A",
		ID:              2,
	}
	eA3 := TestEntity{
		SortingProperty: "A",
		ID:              3,
	}
	eB1 := TestEntity{
		SortingProperty: "B",
		ID:              1,
	}
	mq.Enqueue(eA1)
	time.Sleep(1 * time.Nanosecond)
	mq.Enqueue(eA2)
	time.Sleep(1 * time.Nanosecond)
	mq.Enqueue(eA3)
	time.Sleep(1 * time.Nanosecond)
	mq.Enqueue(eB1)
	// Added Entities to two different sorted queues
	if len(mq.SortedQueues) != 2 {
		t.Errorf("Integration test failed, len sorted queues got = %v, want = 2", len(mq.SortedQueues))
		return
	}
	// Check if i get oldest Entity first
	respEA1, err := mq.Dequeue()
	if err != nil {
		t.Errorf("Integration test failed with unexpected error = %v", err)
		return
	}
	if !cmp.Equal(respEA1, eA1) {
		t.Errorf("Integration test failed with unexpected return value got = %v, want = %v", respEA1, eA1)
		return
	}
	// Check if i get oldest Entity from second sorted queue
	respEB1, err := mq.Dequeue()
	if err != nil {
		t.Errorf("Integration test failed with unexpected error = %v", err)
		return
	}
	if !cmp.Equal(respEB1, eB1) {
		t.Errorf("Integration test failed with unexpected return value got = %v, want = %v", respEA1, eA1)
		return
	}
	// all queue should be blocked
	val, err := mq.Dequeue()
	if !errors.Is(err, ErrNoUnblockedQueueFoundError) {
		t.Errorf("Integration test failed with wrong error got= %v, want = %v", err, ErrNoUnblockedQueueFoundError)
		return
	}
	if val != nil {
		t.Errorf("Integration test failed with wrong return value got = %v, want = %v", val, nil)
		return
	}
	if len(mq.SortedQueues["B"].Entities) != 1 {
		t.Errorf("Integration test failed with wrong sortedQueue B len got = %v, want = %v", len(mq.SortedQueues["B"].Entities), 1)
		return
	}
	if len(mq.SortedQueues["A"].Entities) != 3 {
		t.Errorf("Integration test failed with wrong sortedQueue B len got = %v, want = %v", len(mq.SortedQueues["B"].Entities), 3)
		return
	}
	// Unblock B queue - should be deleted then since there are no Entities left
	err = mq.Unblock("B")
	if err != nil {
		t.Errorf("Integration test failed with wrong error got = %v, want = %v", err, nil)
		return
	}
	if _, okay := mq.SortedQueues["B"]; okay == true {
		t.Errorf("Integration test failed with still existing sorted queue B got = %v", mq.SortedQueues["B"])
		return
	}
	// All queues should still be blocked
	val, err = mq.Dequeue()
	if !errors.Is(err, ErrNoUnblockedQueueFoundError) {
		t.Errorf("Integration test failed with wrong error got= %v, want = %v", err, ErrNoUnblockedQueueFoundError)
		return
	}
	if val != nil {
		t.Errorf("Integration test failed with wrong return value got = %v, want = %v", val, nil)
		return
	}
	err = mq.Unblock("A")
	if err != nil {
		t.Errorf("Integration test failed with wrong error got = %v, want = %v", err, nil)
		return
	}
	// dequeue should work again
	respEA2, err := mq.Dequeue()
	if err != nil {
		t.Errorf("Integration test failed with unexpected error = %v", err)
		return
	}
	if !cmp.Equal(respEA2, eA2) {
		t.Errorf("Integration test failed with unexpected return value got = %v, want = %v", respEA1, eA1)
		return
	}
	err = mq.Unblock("A")
	if err != nil {
		t.Errorf("Integration test failed with wrong error got = %v, want = %v", err, nil)
		return
	}
	// dequeue should work again
	respEA3, err := mq.Dequeue()
	if err != nil {
		t.Errorf("Integration test failed with unexpected error = %v", err)
		return
	}
	if !cmp.Equal(respEA3, eA3) {
		t.Errorf("Integration test failed with unexpected return value got = %v, want = %v", respEA1, eA1)
		return
	}
	// enqueue another Entity while no Entities left but still should exist while blocked
	mq.Enqueue(eA1)
	if len(mq.SortedQueues["A"].Entities) != 2 {
		t.Errorf("Integration test failed with wrong sortedQueue A len got = %v, want = %v", len(mq.SortedQueues["A"].Entities), 2)
		return
	}
	err = mq.Unblock("A")
	if err != nil {
		t.Errorf("Integration test failed with wrong error got = %v, want = %v", err, nil)
		return
	}
	// get the last Entity now but return with error
	respEA12, err := mq.Dequeue()
	if err != nil {
		t.Errorf("Integration test failed with unexpected error = %v", err)
		return
	}
	if !cmp.Equal(respEA12, eA1) {
		t.Errorf("Integration test failed with unexpected return value got = %v, want = %v", respEA1, eA1)
		return
	}
	err = mq.UnblockWithError("A")
	if err != nil {
		t.Errorf("Integration test failed with wrong error got = %v, want = %v", err, nil)
		return
	}
	// get the last Entity now frfr
	respEA12, err = mq.Dequeue()
	if err != nil {
		t.Errorf("Integration test failed with unexpected error = %v", err)
		return
	}
	if !cmp.Equal(respEA12, eA1) {
		t.Errorf("Integration test failed with unexpected return value got = %v, want = %v", respEA1, eA1)
		return
	}
	err = mq.Unblock("A")
	if err != nil {
		t.Errorf("Integration test failed with wrong error got = %v, want = %v", err, nil)
		return
	}
	if len(mq.SortedQueues) != 0 {
		t.Errorf("Integration test failed with wrong sorted queue len got = %v, want = %v", len(mq.SortedQueues), 0)
		return
	}
}

func TestMultiQueue_Dequeue(t *testing.T) {
	type fields struct {
		sortingProperty string
		sortedQueues    map[string]*SortedQueue
	}
	tests := []struct {
		name    string
		fields  fields
		want    any
		wantErr bool
	}{
		{
			name: "expect error while getting non blocked sorted queue with oldest Entity without sorted queues",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues:    map[string]*SortedQueue{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "expect error while getting non blocked sorted queue with oldest Entity with one blocked sorted queue",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy case dequeue last Entity",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
				},
			},
			want: TestEntity{
				SortingProperty: "A",
				ID:              1,
			},
			wantErr: false,
		},
		{
			name: "happy case dequeue Entity with more Entities existing going forward",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              2,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 10, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
				},
			},
			want: TestEntity{
				SortingProperty: "A",
				ID:              1,
			},
			wantErr: false,
		},
		{
			name: "happy case dequeue Entity and sorted queue exists afterwards",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              2,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 10, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
					"B": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "B",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 5, 0, 0, time.UTC),
							},
							{
								Entity: TestEntity{
									SortingProperty: "B",
									ID:              2,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 10, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 5, 0, 0, time.UTC),
						IsBlocked:        false,
					},
				},
			},
			want: TestEntity{
				SortingProperty: "A",
				ID:              1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MultiQueue{
				SortingProperty: tt.fields.sortingProperty,
				SortedQueues:    tt.fields.sortedQueues,
				mu:              sync.Mutex{},
			}
			copiedSortedQueues := map[string]*SortedQueue{}
			for key, copiedValue := range tt.fields.sortedQueues {
				copiedSortedQueue := *copiedValue
				copiedSortedQueues[key] = &copiedSortedQueue
			}
			got, err := m.Dequeue()
			if (err != nil) != tt.wantErr {
				t.Errorf("Dequeue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("Dequeue() got = %v, want %v", got, tt.want)
			}
			if !tt.wantErr {
				sortingPropertyValue, err1 := m.getSortingPropertyValue(tt.want)
				if err1 != nil {
					t.Errorf("Dequeue() error getting sorting property value with error = %v", err)
				}
				if len(m.SortedQueues[sortingPropertyValue].Entities) != len(copiedSortedQueues[sortingPropertyValue].Entities) {
					t.Errorf("Dequeue() len of sorted queue was reduced with got = %v want = %v", copiedSortedQueues[sortingPropertyValue].Entities, m.SortedQueues[sortingPropertyValue].Entities)
				}
			}
		})
	}
}

func TestMultiQueue_Enqueue(t *testing.T) {
	type fields struct {
		sortingProperty string
		sortedQueues    map[string]*SortedQueue
	}
	type args struct {
		e any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error at getting the sorting property value",
			fields: fields{
				sortingProperty: "WrongSortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				TestEntity{
					SortingProperty: "A",
					ID:              1,
				},
			},
			wantErr: true,
		},
		{
			name: "error at getting the sorted queue with empty sorting property value",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				TestEntity{
					SortingProperty: "",
					ID:              1,
				},
			},
			wantErr: true,
		},
		{
			name: "happy case with creating a new sorted queue",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				TestEntity{
					SortingProperty: "B",
					ID:              1,
				},
			},
			wantErr: false,
		},
		{
			name: "happy case with adding to a existing sorted queue",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				TestEntity{
					SortingProperty: "A",
					ID:              2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MultiQueue{
				SortingProperty: tt.fields.sortingProperty,
				SortedQueues:    tt.fields.sortedQueues,
				mu:              sync.Mutex{},
			}
			if err := m.Enqueue(tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("Enqueue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				sortingPropertyValue, err := m.getSortingPropertyValue(tt.args.e)
				if err != nil {
					t.Errorf("Enqueue() error getting sorting property value with error = %v", err)
				}
				if !cmp.Equal(m.SortedQueues[sortingPropertyValue].Entities[len(m.SortedQueues[sortingPropertyValue].Entities)-1].Entity, tt.args.e) {
					t.Errorf("Enqueue() wrong Entity in newest position of sorted queue = %v got = %v, want %v", sortingPropertyValue, m.SortedQueues[sortingPropertyValue].Entities[len(m.SortedQueues[sortingPropertyValue].Entities)-1].Entity, tt.args.e)
				}
			}
		})
	}
}

func TestMultiQueue_Unblock(t *testing.T) {
	type fields struct {
		sortingProperty string
		sortedQueues    map[string]*SortedQueue
	}
	type args struct {
		sortingPropertyValue string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantEntitiesLen int
		wantErr         bool
	}{
		{
			name: "error queue not exist",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				sortingPropertyValue: "B",
			},
			wantErr: true,
		},
		{
			name: "error no Entities in queue",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities:         []SortedQueueEntity{},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				sortingPropertyValue: "A",
			},
			wantErr: true,
		},
		{
			name: "error queue already unblocked",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
				},
			},
			args: args{
				sortingPropertyValue: "A",
			},
			wantErr: true,
		},
		{
			name: "happy case with queue deleted",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				sortingPropertyValue: "A",
			},
			wantErr: false,
		},
		{
			name: "happy case with queue existing going forward",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
							},
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              2,
								},
								InsertionTime: time.Date(2024, 06, 26, 11, 00, 12, 0, time.UTC),
							},
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              3,
								},
								InsertionTime: time.Date(2024, 06, 26, 11, 00, 15, 0, time.UTC),
							},
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              4,
								},
								InsertionTime: time.Date(2024, 06, 26, 11, 00, 18, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				sortingPropertyValue: "A",
			},
			wantEntitiesLen: 3,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MultiQueue{
				SortingProperty: tt.fields.sortingProperty,
				SortedQueues:    tt.fields.sortedQueues,
				mu:              sync.Mutex{},
			}
			if err := m.Unblock(tt.args.sortingPropertyValue); (err != nil) != tt.wantErr {
				t.Errorf("Unblock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_, okay := m.SortedQueues[tt.args.sortingPropertyValue]
			if okay {
				if !tt.wantErr && m.SortedQueues[tt.args.sortingPropertyValue].IsBlocked {
					t.Errorf("Unblock() for sorting property value = %v got = %v, want false", tt.args.sortingPropertyValue, m.SortedQueues[tt.args.sortingPropertyValue].IsBlocked)
					return
				}
				if len(m.SortedQueues[tt.args.sortingPropertyValue].Entities) != tt.wantEntitiesLen && !tt.wantErr {
					t.Errorf("Unblock() Entities were not altered got = %v with len =%v, want len %v", m.SortedQueues[tt.args.sortingPropertyValue].Entities, len(m.SortedQueues[tt.args.sortingPropertyValue].Entities), tt.wantEntitiesLen)
					return
				}
			}
		})
	}
}

func TestMultiQueue_UnblockWithError(t *testing.T) {
	type fields struct {
		sortingProperty string
		sortedQueues    map[string]*SortedQueue
	}
	type args struct {
		sortingPropertyValue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error queue not exist",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				sortingPropertyValue: "B",
			},
			wantErr: true,
		},
		{
			name: "error queue already unblocked",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
				},
			},
			args: args{
				sortingPropertyValue: "A",
			},
			wantErr: true,
		},
		{
			name: "happy case with Entity still in queue",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
							},
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              2,
								},
								InsertionTime: time.Date(2024, 06, 26, 11, 00, 12, 0, time.UTC),
							},
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              3,
								},
								InsertionTime: time.Date(2024, 06, 26, 11, 00, 15, 0, time.UTC),
							},
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              4,
								},
								InsertionTime: time.Date(2024, 06, 26, 11, 00, 18, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
			},
			args: args{
				sortingPropertyValue: "A",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MultiQueue{
				SortingProperty: tt.fields.sortingProperty,
				SortedQueues:    tt.fields.sortedQueues,
				mu:              sync.Mutex{},
			}

			copiedEntities := []SortedQueueEntity{}
			_, okay := m.SortedQueues[tt.args.sortingPropertyValue]
			if okay {
				copiedEntities = m.SortedQueues[tt.args.sortingPropertyValue].Entities
			}

			if err := m.UnblockWithError(tt.args.sortingPropertyValue); (err != nil) != tt.wantErr {
				t.Errorf("UnblockWithError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if okay {
				if !tt.wantErr && m.SortedQueues[tt.args.sortingPropertyValue].IsBlocked {
					t.Errorf("UnblockWithError() for sorting property value = %v got = %v, want false", tt.args.sortingPropertyValue, m.SortedQueues[tt.args.sortingPropertyValue].IsBlocked)
					return
				}
				if len(m.SortedQueues[tt.args.sortingPropertyValue].Entities) != len(copiedEntities) && !tt.wantErr {
					t.Errorf("UnblockWithError() Entities were altered got = %v, want %v", m.SortedQueues[tt.args.sortingPropertyValue].Entities, copiedEntities)
				}
			}
		})
	}
}

func TestMultiQueue_getNonBlockedSortedQueueWithOldestEntity(t *testing.T) {
	type fields struct {
		sortingProperty string
		sortedQueues    map[string]*SortedQueue
		mu              sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		want    *SortedQueue
		wantErr bool
	}{
		{
			name: "expect no unblocked queue found error with no queues at all",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues:    map[string]*SortedQueue{},
				mu:              sync.Mutex{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "expect no unblocked queue found error with one blocked queue",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities:         nil,
						OldestEntityTime: time.Time{},
						IsBlocked:        true,
					},
				},
				mu: sync.Mutex{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get unblocked sorted queue with one sorted queue present",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
				},
				mu: sync.Mutex{},
			},
			want: &SortedQueue{

				Entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
					},
				},
				OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
				IsBlocked:        true,
			},
			wantErr: false,
		},
		{
			name: "get unblocked sorted queue with two unblocked sorted queue present",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
					"B": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "B",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 1, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 1, 0, 0, time.UTC),
						IsBlocked:        false,
					},
				},
				mu: sync.Mutex{},
			},
			want: &SortedQueue{

				Entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
					},
				},
				OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
				IsBlocked:        true,
			},
			wantErr: false,
		},
		{
			name: "get unblocked sorted queue with two unblocked sorted queue present and one older blocked sorted queue",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
					"B": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "B",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 1, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 1, 0, 0, time.UTC),
						IsBlocked:        false,
					},
					"C": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "C",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 10, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 10, 0, 0, 0, time.UTC),
						IsBlocked:        true,
					},
				},
				mu: sync.Mutex{},
			},
			want: &SortedQueue{

				Entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
					},
				},
				OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
				IsBlocked:        true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MultiQueue{
				SortingProperty: tt.fields.sortingProperty,
				SortedQueues:    tt.fields.sortedQueues,
				mu:              tt.fields.mu,
			}
			got, err := m.getNonBlockedSortedQueueWithOldestEntity()
			if (err != nil) != tt.wantErr {
				t.Errorf("getNonBlockedSortedQueueWithOldestEntity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if no sorted queue expected return
			if tt.want == nil {
				return
			}
			// compare sorted queue manually
			// compare OldestEntityTime
			if !got.OldestEntityTime.Equal(tt.want.OldestEntityTime) {
				t.Errorf("getNonBlockedSortedQueueWithOldestEntity() OldestEntityTime got = %v, want %v", got.OldestEntityTime, tt.want.OldestEntityTime)
			}
			// compare IsBlocked
			if got.IsBlocked != tt.want.IsBlocked {
				t.Errorf("getNonBlockedSortedQueueWithOldestEntity() IsBlocked got = %v, want %v", got.IsBlocked, tt.want.IsBlocked)
			}
			// compare Entities
			for i, _ := range got.Entities {
				// compare Entity
				if !cmp.Equal(got.Entities[i].Entity, tt.want.Entities[i].Entity) {
					t.Errorf("getNonBlockedSortedQueueWithOldestEntity() Entities got = %v, want %v with diff = %v", got.Entities[i].Entity, tt.want.Entities[i].Entity, cmp.Diff(got.Entities[i].Entity, tt.want.Entities[i].Entity))
				}
				// compare insertion time
				if !got.Entities[i].InsertionTime.Equal(tt.want.Entities[i].InsertionTime) {
					t.Errorf("getNonBlockedSortedQueueWithOldestEntity() insertion time got = %v, want %v", got.Entities[i].InsertionTime.String(), tt.want.Entities[i].InsertionTime.String())
				}
			}
		})
	}
}

func TestMultiQueue_getSortedQueue(t *testing.T) {
	type fields struct {
		sortingProperty string
		sortedQueues    map[string]*SortedQueue
		mu              sync.Mutex
	}
	type args struct {
		sortingPropertyValue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SortedQueue
		wantErr bool
	}{
		{
			name: "expect empty sorting property value error",
			fields: fields{
				sortingProperty: "A",
				sortedQueues:    map[string]*SortedQueue{},
				mu:              sync.Mutex{},
			},
			args: args{
				sortingPropertyValue: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy case create new sorted queue",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues:    map[string]*SortedQueue{},
				mu:              sync.Mutex{},
			},
			args: args{
				sortingPropertyValue: "A",
			},
			want:    NewSortedQueue(),
			wantErr: false,
		},
		{
			name: "happy case get existing sorted queue with only one SortedQueue existing",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
				},
				mu: sync.Mutex{},
			},
			args: args{
				sortingPropertyValue: "A",
			},
			want: &SortedQueue{
				Entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
					},
				},
				OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
				IsBlocked:        false,
			},
			wantErr: false,
		},
		{
			name: "happy case get existing sorted queue with three SortedQueue existing",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues: map[string]*SortedQueue{
					"A": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "A",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
					"B": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "B",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
					"C": {
						Entities: []SortedQueueEntity{
							{
								Entity: TestEntity{
									SortingProperty: "C",
									ID:              1,
								},
								InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
							},
						},
						OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
						IsBlocked:        false,
					},
				},
				mu: sync.Mutex{},
			},
			args: args{
				sortingPropertyValue: "A",
			},
			want: &SortedQueue{
				Entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
					},
				},
				OldestEntityTime: time.Date(2024, 06, 27, 11, 0, 0, 0, time.UTC),
				IsBlocked:        false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MultiQueue{
				SortingProperty: tt.fields.sortingProperty,
				SortedQueues:    tt.fields.sortedQueues,
				mu:              tt.fields.mu,
			}

			got, err := m.getSortedQueue(tt.args.sortingPropertyValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSortedQueue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if no sorted queue expected return
			if tt.want == nil {
				return
			}
			// compare sorted queue manually
			// compare OldestEntityTime
			if !got.OldestEntityTime.Equal(tt.want.OldestEntityTime) {
				t.Errorf("getSortedQueue() OldestEntityTime got = %v, want %v", got.OldestEntityTime, tt.want.OldestEntityTime)
			}
			// compare IsBlocked
			if got.IsBlocked != tt.want.IsBlocked {
				t.Errorf("getSortedQueue() IsBlocked got = %v, want %v", got.IsBlocked, tt.want.IsBlocked)
			}
			// compare Entities
			for i, _ := range got.Entities {
				// compare Entity
				if !cmp.Equal(got.Entities[i].Entity, tt.want.Entities[i].Entity) {
					t.Errorf("getSortedQueue() Entities got = %v, want %v with diff = %v", got.Entities[i].Entity, tt.want.Entities[i].Entity, cmp.Diff(got.Entities[i].Entity, tt.want.Entities[i].Entity))
				}
				// compare insertion time
				if !got.Entities[i].InsertionTime.Equal(tt.want.Entities[i].InsertionTime) {
					t.Errorf("getSortedQueue() insertion time got = %v, want %v", got.Entities[i].InsertionTime.String(), tt.want.Entities[i].InsertionTime.String())
				}
			}
		})
	}
}

func TestMultiQueue_getSortingPropertyValue(t *testing.T) {
	type TestEntityNotAString struct {
		SortingProperty int
		ID              int
	}
	type fields struct {
		sortingProperty string
		sortedQueues    map[string]*SortedQueue
		mu              sync.Mutex
	}
	type args struct {
		e any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "expect Entity is not a struct error",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues:    nil,
				mu:              sync.Mutex{},
			},
			args: args{
				e: 5,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "expect sorting property not exist error",
			fields: fields{
				sortingProperty: "WrongSortingProperty",
				sortedQueues:    nil,
				mu:              sync.Mutex{},
			},
			args: args{
				e: TestEntity{
					SortingProperty: "A",
					ID:              1,
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "expect sorting property not a string error",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues:    nil,
				mu:              sync.Mutex{},
			},
			args: args{
				e: TestEntityNotAString{
					SortingProperty: 1,
					ID:              1,
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "expect sorting property empty string error",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues:    nil,
				mu:              sync.Mutex{},
			},
			args: args{
				e: TestEntity{
					SortingProperty: "",
					ID:              1,
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "happy case",
			fields: fields{
				sortingProperty: "SortingProperty",
				sortedQueues:    nil,
				mu:              sync.Mutex{},
			},
			args: args{
				e: TestEntity{
					SortingProperty: "A",
					ID:              1,
				},
			},
			want:    "A",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MultiQueue{
				SortingProperty: tt.fields.sortingProperty,
				SortedQueues:    tt.fields.sortedQueues,
				mu:              tt.fields.mu,
			}
			got, err := m.getSortingPropertyValue(tt.args.e)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSortingPropertyValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSortingPropertyValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortedQueue_block(t *testing.T) {
	type fields struct {
		entities         []SortedQueueEntity
		oldestEntityTime time.Time
		isBlocked        bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "block unblocked queue",
			fields: fields{
				entities:         nil,
				oldestEntityTime: time.Time{},
				isBlocked:        false,
			},
			wantErr: false,
		},
		{
			name: "block blocked queue expect ErrSortedQueueAlreadyBlockedError",
			fields: fields{
				entities:         nil,
				oldestEntityTime: time.Time{},
				isBlocked:        true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SortedQueue{
				Entities:         tt.fields.entities,
				OldestEntityTime: tt.fields.oldestEntityTime,
				IsBlocked:        tt.fields.isBlocked,
			}
			if err := s.block(); (err != nil) != tt.wantErr {
				t.Errorf("block() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSortedQueue_dequeue(t *testing.T) {
	type fields struct {
		entities         []SortedQueueEntity
		oldestEntityTime time.Time
		isBlocked        bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "empty queue expect ErrNoEntitiesInQueueError",
			fields: fields{
				entities:         nil,
				oldestEntityTime: time.Time{},
				isBlocked:        false,
			},
			wantErr: true,
		},
		{
			name: "dequeue last Entity expect no error",
			fields: fields{
				entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
					},
				},
				oldestEntityTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
				isBlocked:        false,
			},
			wantErr: false,
		},
		{
			name: "dequeue one Entity expect no error",
			fields: fields{
				entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
					},
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              2,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 12, 0, time.UTC),
					},
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              3,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 15, 0, time.UTC),
					},
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              4,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 18, 0, time.UTC),
					},
				},
				oldestEntityTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
				isBlocked:        false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SortedQueue{
				Entities:         tt.fields.entities,
				OldestEntityTime: tt.fields.oldestEntityTime,
				IsBlocked:        tt.fields.isBlocked,
			}

			err := s.dequeue()
			if (err != nil) != tt.wantErr {
				t.Errorf("dequeue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(s.Entities) == len(tt.fields.entities) && !tt.wantErr {
				t.Errorf("dequeue() Entities were not altered got = %v, want %v", s.Entities, tt.fields.entities)
			}
			if len(s.Entities) >= 1 {
				if tt.fields.entities[0].InsertionTime.Equal(s.OldestEntityTime) {
					t.Errorf("dequeue() insertion time was not altered got = %v, want %v", s.OldestEntityTime, tt.fields.entities[0].InsertionTime)
				}
			}
		})
	}
}

func TestSortedQueue_enqueue(t *testing.T) {
	type fields struct {
		entities         []SortedQueueEntity
		oldestEntityTime time.Time
		isBlocked        bool
	}
	type args struct {
		e  any
		ts time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name: "enqueue first element",
			fields: fields{
				entities:         []SortedQueueEntity{},
				oldestEntityTime: time.Time{},
				isBlocked:        false,
			},
			args: args{
				e: TestEntity{
					SortingProperty: "A",
					ID:              1,
				},
				ts: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
			},
			want: fields{
				entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
					},
				},
				oldestEntityTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
				isBlocked:        false,
			},
		},
		{
			name: "enqueue second element",
			fields: fields{
				entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
					},
				},
				oldestEntityTime: time.Time{},
				isBlocked:        false,
			},
			args: args{
				e: TestEntity{
					SortingProperty: "A",
					ID:              2,
				},
				ts: time.Date(2024, 06, 26, 11, 00, 12, 0, time.UTC),
			},
			want: fields{
				entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
					},
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              2,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 12, 0, time.UTC),
					},
				},
				oldestEntityTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
				isBlocked:        false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SortedQueue{
				Entities:         tt.fields.entities,
				OldestEntityTime: tt.fields.oldestEntityTime,
				IsBlocked:        tt.fields.isBlocked,
			}
			s.enqueue(tt.args.e, tt.args.ts)
			// check correct len of Entity slices
			if len(s.Entities) != len(tt.want.entities) {
				t.Errorf("enqueue() entitiesLen = %v, wantEntitiesLen %v", len(s.Entities), len(tt.want.entities))
				return
			}
			// compare Entities
			for i, _ := range s.Entities {
				// compare insertion times
				if !s.Entities[i].InsertionTime.Equal(tt.want.entities[i].InsertionTime) {
					t.Errorf("enqueue() different insertion times for Entity = %v time = %v, wantTime %v", s.Entities[i], s.Entities[i].InsertionTime.String(), tt.want.entities[i].InsertionTime.String())
					return
				}
				// compare Entity
				if !cmp.Equal(s.Entities[i].Entity, tt.want.entities[i].Entity) {
					t.Errorf("enqueue() different Entities got = %v want = %v with diff = %v", s.Entities[i].Entity, tt.want.entities[i].Entity, cmp.Diff(s.Entities[i].Entity, tt.want.entities[i].Entity))
					return
				}
			}
		})
	}
}

func TestSortedQueue_unblock(t *testing.T) {
	type fields struct {
		entities         []SortedQueueEntity
		oldestEntityTime time.Time
		isBlocked        bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "unblock blocked without errors",
			fields: fields{
				entities:         nil,
				oldestEntityTime: time.Time{},
				isBlocked:        true,
			},
			wantErr: false,
		},
		{
			name: "try unblock already unblocked and return ErrSortedQueueAlreadyUnblockedError",
			fields: fields{
				entities:         nil,
				oldestEntityTime: time.Time{},
				isBlocked:        false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SortedQueue{
				Entities:         tt.fields.entities,
				OldestEntityTime: tt.fields.oldestEntityTime,
				IsBlocked:        tt.fields.isBlocked,
			}
			if err := s.unblock(); (err != nil) != tt.wantErr {
				t.Errorf("unblock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSortedQueue_getNextEntity(t *testing.T) {
	type fields struct {
		entities         []SortedQueueEntity
		oldestEntityTime time.Time
		isBlocked        bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    any
		wantErr bool
	}{
		{
			name: "empty queue expect ErrNoEntitiesInQueueError",
			fields: fields{
				entities:         nil,
				oldestEntityTime: time.Time{},
				isBlocked:        false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get last Entity expect no error",
			fields: fields{
				entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
					},
				},
				oldestEntityTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
				isBlocked:        false,
			},
			want: TestEntity{
				SortingProperty: "A",
				ID:              1,
			},
			wantErr: false,
		},
		{
			name: "get next Entity expect no error",
			fields: fields{
				entities: []SortedQueueEntity{
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              1,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
					},
					{
						Entity: TestEntity{
							SortingProperty: "A",
							ID:              2,
						},
						InsertionTime: time.Date(2024, 06, 26, 11, 00, 12, 0, time.UTC),
					},
				},
				oldestEntityTime: time.Date(2024, 06, 26, 11, 00, 0, 0, time.UTC),
				isBlocked:        false,
			},
			want: TestEntity{
				SortingProperty: "A",
				ID:              1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SortedQueue{
				Entities:         tt.fields.entities,
				OldestEntityTime: tt.fields.oldestEntityTime,
				IsBlocked:        tt.fields.isBlocked,
			}
			got, err := s.getNextEntity()
			if (err != nil) != tt.wantErr {
				t.Errorf("getNextEntity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("getNextEntity() got = %v, want %v with diff = %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}
