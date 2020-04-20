package simulator

import (
	"time"
)

type Order struct {
	Id string
	Name string
	Temp string
	ShelfLife uint32
	DecayRate float32
	DecayScore float32
	IsCritical bool
	shelf *Shelf
	placementTime time.Time
	arrivalTime time.Time
}

func(o *Order) computeDecayScore(modifier uint,arrival_time_ms int64) float32{
	// TODO: please fix up the type coercions, they're nasty
	a := float32(o.ShelfLife)
	b := o.DecayRate*(float32(arrival_time_ms)/1000)*float32(modifier)
	if a == b  || a == 0{
		return 0
	}
	return (a-b)/a
}

func(o *Order) swapWillPreserve(modifier uint, getNow timeFunc) bool {
	/*
		Needs to compute prospective decay score
		based on current elapsed score + remaining elapsed score 

		if new prospective score is greater than 0, update the decay score
		to be the new prospective score.

		1. we have the initially computed decay score
		2. we have the timestamp of when it was placed on the shelf.
		3. we have the distance in seconds from when it was placed to when 
			it will be picked up
		4. we have the current timestamp.
		The formula for this is then:
		elapsed = computeScore(o.shelf.modifier,current_time-initial_time)
		on_new_shelf = computeScore(new_modifier,arrival_time-current_time)
		prospective_score = elapsed + on_new_shelf

		TODO: Please make the types stop using all this coercion and casting.
		it's ugly
	*/
	//currentTimeMS := time.Now().UnixNano()/int64(time.Millisecond)
	currentTimeMS := getNow().UnixNano()/int64(time.Millisecond)
	initialTimeMS := o.placementTime.UnixNano()/int64(time.Millisecond)
	arrivalTimeMS := o.arrivalTime.UnixNano()/int64(time.Millisecond)
	elapsedMS := currentTimeMS - initialTimeMS
	prospectiveMS := arrivalTimeMS - currentTimeMS
	elapsedScore := o.computeDecayScore(o.shelf.modifier,elapsedMS)
	newShelfScore := o.computeDecayScore(modifier,prospectiveMS)
	prospectiveScore := newShelfScore + elapsedScore
	if prospectiveScore > 0{
		o.IsCritical = false
		o.DecayScore = prospectiveScore
		return true
	}
	return false
}

func (o *Order) selectShelf(s *Shelves,arrival_delay int,getNow timeFunc) *Shelf {
	/*
	TODO: Add a narrative for this.

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

	overflowDecayScore := o.computeDecayScore(s.overflow.modifier,
					int64(arrival_delay*1000))
	matchingDecayScore := o.computeDecayScore(matchingShelf.modifier,
				int64(arrival_delay*1000))
//	o.placementTime = time.Now()
	o.placementTime = getNow()
	o.arrivalTime = o.placementTime.Add(time.Second*time.Duration(arrival_delay))
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
