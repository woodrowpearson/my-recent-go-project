package simulator

import (
	"sync/atomic"
	"github.com/orcaman/concurrent-map"
)



type orderShelf struct {
	counter int32
	contents cmap.ConcurrentMap
	name string
	modifier uint
	criticals cmap.ConcurrentMap
}

/*
Helper function for constructing shelf struct.
*/
func buildOrderShelf(array_capacity uint, name string,
		modifier uint) *orderShelf {
	shelf := new(orderShelf)
	shelf.name = name;
	shelf.counter = int32(array_capacity)
	shelf.modifier = modifier
	shelf.contents = cmap.New()
	shelf.criticals = cmap.New()
	return shelf
}

/*
Removes an order from a shelf and updates capacity counter in a threadsafe manner.
Removes order from at-risk map if order is at-risk.
*/
func (s *orderShelf) incrementAndUpdate(o *foodOrder,remove_from_criticals bool){
	/*
		removes item from shelf
	*/
	s.contents.Remove(o.Id)
	if remove_from_criticals {
		s.criticals.Remove(o.Id)
	}
	atomic.AddInt32(&s.counter,1)
}

/*
Adds a new order to a shelf and updates capacity counter in a threadsafe manner.
Adds order from at-risk map if order is at-risk.
*/
func (s *orderShelf) decrementAndUpdate(o *foodOrder) {
	s.contents.Set(o.Id, o)
	if o.IsCritical {
		s.criticals.Set(o.Id,o)
	}
	atomic.AddInt32(&s.counter, -1);
}

/*
Casts a value to an Order pointer. Necessary for accessing values from concurrent hashmap.
*/
func castToOrder(blob interface{}) *foodOrder{
	switch order := blob.(type){
		case *foodOrder:
			return order
		default:
			panic("wrong type!!")
	}
}

/*
Scans shelf's map of at-risk orders and returns an eligible order for swapping
to a safe shelf.
*/
func(s *orderShelf) selectCritical(overflow *orderShelf,getNow timeFunc) *foodOrder{
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

/*
Pushes keys of shelf contents to a slice in a threadsafe manner for logging purposes. 
*/
func(s *orderShelf) duplicateContentsToMap(order *foodOrder,with_order bool) map[string]*foodOrder {
	/*
	Range expression is evaluated once, at the start.
	we're doing this to make a copy of the current shelf,
	so that we don't risk weirdness in printing shelf contents
	based on the concurrent maps.
	*/
	contents := make(map[string]*foodOrder)
	for _, v := range s.contents.Items(){
		o := castToOrder(v)
		if with_order || (o.Id != order.Id){
			contents[o.Id] = o
		}
	}
	return contents
}

/*
Pushes keys of shelf contents to a slice in a threadsafe manner for logging purposes. 
*/
func(s *orderShelf) duplicateContentsToSlice(order *foodOrder, with_order bool) []string{
	/*
	Range expression is evaluated once, at the start.
	we're doing this to make a copy of the current shelf,
	so that we don't risk weirdness in printing shelf contents
	based on the concurrent maps.
	*/
	contents := []string{}
	for _,v := range s.contents.Items(){
		o := castToOrder(v)
		if with_order || (o.Id != order.Id){
			contents = append(contents,o.Id)
		}
	}
	return contents
}


/*
Determine if we can move an at-risk order from the overflow shelf
in order to prevent it from decaying before pickup
*/
func(s *orderShelf) swapAssessment(o *foodOrder, overflow *orderShelf,statistics *Statistics,getNow timeFunc){
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
	if s != overflow {
		to_swap := s.selectCritical(overflow,getNow)
		if to_swap != nil{
			overflow.incrementAndUpdate(to_swap,true)
			s.contents.Remove(o.Id)
			s.contents.Set(to_swap.Id,to_swap)
			statistics.updateSwapped()
		} else {
			/*
				Any order not in overflow is categorically not critical.
			*/
			s.incrementAndUpdate(o,false)
		}
	} else {
		// We could be removing a critical order on overflow
		s.incrementAndUpdate(o,o.IsCritical)
	}
}

// Helper struct for keeping argument lengths reasonable.
type orderShelves struct{
	overflow *orderShelf
	cold *orderShelf
	hot *orderShelf
	frozen *orderShelf
	dead *orderShelf
}
