package main



type Order struct {
	Id string
	Name string
	Temp string
	ShelfLife uint32
	DecayRate float32
	DecayScore float32
	IsCritical bool
	shelf *Shelf
	// TODO: add an initial age here
	// that we can use to recompute the decayscore in the event of a swap
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

func(o *Order) Snapshot(modifier uint) float32 {
	/*
		Needs to compute prospective decay score
		based on current elapsed score + remaining elapsed score 
	*/
	return 1
}

func (o *Order) selectShelf(s *Shelves,arrival_time int) *Shelf {
	/*
	TODO: add in a criticality score for the order.
	If the order is not safe for overflow, don't stick it 
	in overflow unless matching shelf is empty.

	*/
	matchingShelf := s.overflow
	switch o.Temp{
		case "cold":
			matchingShelf = s.cold
		case "hot":
			matchingShelf = s.hot
		case "frozen":
			matchingShelf = s.frozen
	}

	overflowDecayScore := o.computeDecayScore(s.overflow,
					arrival_time)
	matchingDecayScore := o.computeDecayScore(matchingShelf,
				arrival_time)
	if (s.overflow.counter < 1 && matchingShelf.counter < 1){
		// nowhere to place, must discard.
		o.shelf = s.dead
		return s.dead
	}
	if (overflowDecayScore <= 0 && matchingDecayScore <= 0){
		// will die no matter what due to expiration.
		o.shelf = s.dead
		return s.dead
	}

	if (overflowDecayScore > 0 && s.overflow.counter > 0){
		// will survive overflow
		o.shelf = s.overflow
		o.DecayScore = overflowDecayScore
		o.shelf.decrementAndUpdate(o)
		return s.overflow
	}
	if (matchingDecayScore > 0 && matchingShelf.counter > 0){
		o.shelf = matchingShelf
		o.DecayScore = matchingDecayScore
		o.shelf.decrementAndUpdate(o)
		return matchingShelf
	}
	/*
	 The only case not accounted for here
	is when we've got a <= 0 and it'll only go into overflow.
	*/
	o.IsCritical = true
	o.DecayScore = overflowDecayScore
	o.shelf = s.overflow
	o.shelf.decrementAndUpdate(o)
	return s.overflow
}
