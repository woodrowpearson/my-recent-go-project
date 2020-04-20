package simulator

import (
	"sync/atomic"
	"github.com/orcaman/concurrent-map"
)



type Shelf struct {
	counter int32
	contents cmap.ConcurrentMap
	name string
	modifier uint
	criticals cmap.ConcurrentMap
}

func buildShelf(array_capacity uint, name string,
		modifier uint) *Shelf {
	// TODO: UNIT TEST THIS
	shelf := new(Shelf)
	shelf.name = name;
	shelf.counter = int32(array_capacity)
	shelf.modifier = modifier
	shelf.contents = cmap.New()
	shelf.criticals = cmap.New()
	return shelf
}

func (s *Shelf) incrementAndUpdate(o *Order){
	/*
		removes item from shelf
	*/
	s.contents.Remove(o.Id)
	if o.IsCritical {
		s.criticals.Remove(o.Id)
	}
	atomic.AddInt32(&s.counter,1)
}


func (s *Shelf) decrementAndUpdate(o *Order) {
	s.contents.Set(o.Id, o)
	if o.IsCritical {
		s.criticals.Set(o.Id,o)
	}
	atomic.AddInt32(&s.counter, -1);
}

func castToOrder(blob interface{}) *Order{
	switch order := blob.(type){
		case *Order:
			return order
		default:
			panic("wrong type!!")
	}
}

func(s *Shelf) selectCritical(overflow *Shelf) *Order{
	/*
		We need to do the casting because the concurrent map
		only deals with interfaces.
	*/
	/*
		 TODO: THE BUG IS HERE. There is some issue with 
		iterating over the map while doing edits on the map.
		error message is 
		"go fatal error concurrent map iteration and map write"
	*/
	for _, ptr := range overflow.criticals.Items() {
		order := castToOrder(ptr)
		if s.name == order.shelf.name && order.swapWillPreserve(s.modifier){
			return order
		}
	}

	return nil
}

// Helper struct for keeping argument lengths reasonable.
type Shelves struct{
	overflow *Shelf
	cold *Shelf
	hot *Shelf
	frozen *Shelf
	dead *Shelf
}
