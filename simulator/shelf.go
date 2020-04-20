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

func(s *Shelf) selectCritical(overflow *Shelf,getNow timeFunc) *Order{
	/*
		We need to do the casting because the concurrent map
		only deals with interfaces.
	*/
	for _, ptr := range overflow.criticals.Items() {
		order := castToOrder(ptr)
		if s.name == order.Temp && order.swapWillPreserve(s.modifier,getNow){
			return order
		}
	}

	return nil
}

func(s *Shelf) duplicateContents(order *Order, with_order bool) map[string]*Order{
	/*
	 range expression is evaluated once, at the start.
	 we're doing this to make a copy of the current shelf,
	 so that we don't risk weirness in printing shelf contents
	 based on the concurrent maps.
	*/
	contents := make(map[string]*Order)
	for _,v := range s.contents.Items(){
		o := castToOrder(v)
		if with_order || (o.Id != order.Id){
			contents[o.Id] = o
		}
	}
	return contents
}

func(s *Shelf) swapAssessment(o *Order, overflow *Shelf,getNow timeFunc){
	/*
		 In the event that we're freeing up space on
		a non-overflow shelf, we'll want to scan the overflow shelf's
		criticals for the first item that will match the following criteria:
		1) eligible for this shelf due to temperature match
		2) will be saved from decay by moving to the current shelf
		Once the item is found, we swap the item from the matching shelf,
		remove it from criticals, assign it a new decay factor,
		and run incrementAndUpdate on the overflow shelf.
	*/
	// TODO: Print to a logfile when a swap occurs.
	if s != overflow && s.counter == 0{
		to_swap := overflow.selectCritical(s,getNow)
		if to_swap != nil{
			overflow.incrementAndUpdate(to_swap)
			s.contents.Remove(o.Id)
			s.contents.Set(to_swap.Id,to_swap)
		} else {
			s.incrementAndUpdate(o)
		}
	} else {
		s.incrementAndUpdate(o)
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
