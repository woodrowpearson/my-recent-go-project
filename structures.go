package main

import (
	"time"
	"sync/atomic"
	"io"
)

type Order struct {
	Id string
	Name string
	Temp string// this should be an enum. TODO: Does Go have enums?
	ShelfLife uint
	DecayRate float32
}


type Shelf struct {
	counter int32
	item_array []string
	name string
	modifier uint
}

func (s *Shelf) incrementAndUpdate(shelf_idx int){
	// TODO: unit test this. account for thread safety.
	atomic.AddInt32(&s.counter,1)
	s.item_array[shelf_idx] = ""
}

func buildShelf(array_capacity uint, name string,
		modifier uint) *Shelf {
	// TODO: UNIT TEST THIS
	shelf := new(Shelf)
	shelf.item_array = make([]string, array_capacity)
	shelf.name = name;
	shelf.counter = int32(array_capacity)
	shelf.modifier = modifier
	for i := uint(0); i < array_capacity; i++ {
		shelf.item_array[i] = ""
	}
	return shelf
}
// TODO: MAKE THIS THROW AN ERROR ON PATHOLOGICAL RESPONSE OF NONE FOUND
func (s *Shelf) decrementAndUpdate(id string) int {
	atomic.AddInt32(&s.counter, -1);
	// TODO: make this smarter based on the counter value
	// TODO: UNIT TEST THIS
	for i := 0; i < len(s.item_array); i++ {
		if (s.item_array[i] == ""){
			s.item_array[i] = id
			return i
		}
	}
	// Due to where this is called in the worflow,
	// This will never occur
	return -1
}

type Shelves struct{
	overflow *Shelf
	cold *Shelf
	hot *Shelf
	frozen *Shelf
	dead *Shelf
}
type PrimaryArgs struct {
	overflow_size uint
	hot_size uint
	cold_size uint
	frozen_size uint
	courier_lower_bound uint
	courier_upper_bound uint
	orders_per_second uint
	overflow_modifier uint
	cold_modifier uint
	hot_modifier uint
	frozen_modifier uint
	courier_out io.Writer
	courier_err io.Writer
	dispatch_out io.Writer
	dispatch_err io.Writer
	// normally it's 1, but for tests we'll want it at 0.
	// refers to the value of a second
	second_value time.Duration
	shelves *Shelves
}
