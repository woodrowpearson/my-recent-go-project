package main

import (
//	"github.com/francoispqt/gojay"
	"sync/atomic"
	"errors"
)

type Order struct {
	Id string
	Name string
	Temp string
	ShelfLife uint32
	DecayRate float32
	DecayCritical bool
	DecayScore float32
}

func(o *Order) computeDecayScore(s *Shelf,
	arrival_time int) float32{
	a := float32(o.ShelfLife)
	b := o.DecayRate*float32(arrival_time)*float32(s.modifier)
	if a == b {
		return 0
	}
	return (a-b)/a
}

type Shelf struct {
	counter int32
	item_array []string
	name string
	modifier uint
	last_updated_idx uint32
}

func buildShelf(array_capacity uint, name string,
		modifier uint) *Shelf {
	// TODO: UNIT TEST THIS
	shelf := new(Shelf)
	shelf.item_array = make([]string, array_capacity)
	shelf.name = name;
	shelf.counter = int32(array_capacity)
	shelf.modifier = modifier
	shelf.last_updated_idx = 0
	for i := uint(0); i < array_capacity; i++ {
		shelf.item_array[i] = ""
	}
	return shelf
}

func (s *Shelf) incrementAndUpdate(shelf_idx int){
	// TODO: unit test this. account for thread safety.
	/*
		Explanation: before setting the value
		indicating the shelf has available space,
		we want to clear the value out.
		This prevents a scenario where decrementAndUpdate
		overwrites an ID that has not been cleared yet.

	*/
	s.item_array[shelf_idx] = ""
	atomic.StoreUint32(&s.last_updated_idx,uint32(shelf_idx))
	atomic.AddInt32(&s.counter,1)
}
func (s *Shelf) decrementAndUpdate(id string) (int,error) {
	atomic.AddInt32(&s.counter, -1);
	if s.item_array[s.last_updated_idx] != "" {
		// if the spot is taken, we've got to scan for a new
		for i := 0; i < len(s.item_array); i++ {
			if (s.item_array[i] == ""){
				s.item_array[i] = id
				return i,nil
			}
		}
		return -1,errors.New("pathological case on decrementAndUpdate")
	} else {
		// we can avoid a scan if the spot isn't taken
		s.item_array[s.last_updated_idx] = id
		return int(s.last_updated_idx),nil
	}
}

// Helper struct for keeping argument lengths reasonable.
type Shelves struct{
	overflow *Shelf
	cold *Shelf
	hot *Shelf
	frozen *Shelf
	dead *Shelf
}

