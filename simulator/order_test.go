package simulator

import (
	"testing"
	"math"
	"fmt"
)


func assertBoolean(t *testing.T, res bool, expected bool){
	t.Helper()
	if res != expected{
		t.Errorf("res: %t, expected: %t", res,expected)
	}
}

func assertShelf(t *testing.T, res *Shelf, expected *Shelf){
	t.Helper()
	if res != expected{
		t.Errorf("received %+v, expected %+v",res,expected)
	}
}

func assertStrings(t *testing.T, res string, expected string){
	t.Helper()
	if res != expected{
		t.Errorf("received %q, expected %q", res,expected)
	}
}

func assertInt32(t *testing.T, res int32, expected int32){
	t.Helper()
	if res != expected{
		t.Errorf("received %d, expected %d",res,expected)
	}
}
func assertFloat32(t *testing.T, res float32, expected float32){
	t.Helper()
	almost_equal := math.Abs(float64(res)-float64(expected)) <= 1e-9
	if !almost_equal{
		t.Errorf("received %.3f, expected %.3f",res,expected)
	}
}


func TestSwapWillPreserve(t *testing.T){
	/*
		Only two cases for this.
		1. swap WILL preserve it, so the new generated score must be updated.
		2. swap WONT preserve it. order is not updated, function returns false.
		Need to figure out a way to mock the times on this.
	*/

}

func TestComputeDecayScore(t *testing.T){
	/*
		Three cases:
			1. zero shelf life
			2. B greater than A
			3. A greater than B
	*/

	order := Order{Id:"a",Name:"dummy",Temp:"hot",
			ShelfLife:200,DecayRate:0.25}

	t.Run("Returns zero when order shelf life is zero", 
		func(t *testing.T){
		order.ShelfLife = 0
		res := order.computeDecayScore(1,1)
		expected := float32(0)
		assertFloat32(t, res, expected)
	})

	msg := `
Returns a negative result when the decay rate, 
modifier, and arrival time outweigh shelf life.
`
	t.Run(msg,func(t *testing.T){
		order.ShelfLife = 10
		res := order.computeDecayScore(2,1000)
		expected := float32(-49)
		assertFloat32(t,res,expected)
	})


	msg = `
Returns a positive result when shelf life 
outweighs decay factors.
`
	t.Run(msg, func(t *testing.T){
		order.ShelfLife = 200
		res := order.computeDecayScore(1,2)
		expected := float32(0.9975)
		assertFloat32(t,res,expected)
	})
}

func TestSelectShelf(t *testing.T){
	order := Order{Id:"a",Name:"dummy",Temp:"hot",
		ShelfLife: 200, DecayRate: 0.25}
	overflow := buildShelf(1,"overflow",
			0)
	cold := buildShelf(1, "cold",0)
	hot := buildShelf(1,"hot",0)
	frozen := buildShelf(1,"frozen",0)
	dead := buildShelf(0,"dead",0)
	shelves := Shelves{overflow:overflow,cold:cold,
			hot:hot,frozen:frozen,dead:dead}

	t.Run("returns dead if matchingScore and overflorScore are both less than zero",
		func(t *testing.T){
		order.ShelfLife = 0
		res := order.selectShelf(&shelves,100)
		expected := dead
		assertShelf(t,res,expected)
	})

	t.Run("returns dead if no space in matching and overflow shelves",
		func(t *testing.T){
		order.ShelfLife = 200
		overflow.counter = 0
		hot.counter = 0
		res := order.selectShelf(&shelves,100)
		expected := dead
		assertShelf(t,res,expected)

	})

	msg := `
Returns overflow if overflow space is available
and item will survive storage in overflow. 
Ensures that order's shelf is set to overflow,
and that its decay score is set.
`
	t.Run(msg, func(t *testing.T){
		overflow.counter = 1
		order.ShelfLife = 200
		overflow.modifier = 2
		res := order.selectShelf(&shelves,2)
		expected := overflow
		expected_overflow_counter := int32(0)
		expected_decay_score := order.computeDecayScore(overflow.modifier,2)
		assertShelf(t,res,expected)
		assertInt32(t,res.counter,expected_overflow_counter)
		assertFloat32(t,order.DecayScore,expected_decay_score)
	})
	msg = `
Returns matching shelf if eligible for matching shelf
and no space is available in overflow shelf.
Ensures that order's shelf is set to matching shelf,
and that its decay score is set.
`
	t.Run(msg,func(t *testing.T){
		overflow.counter = 0
		hot.counter = 1
		overflow.modifier = 2
		hot.modifier = 1
		res := order.selectShelf(&shelves,2)
		expected := hot
		expected_hot_counter := int32(0)
		expected_decay_score := order.computeDecayScore(hot.modifier,2)
		assertShelf(t,res,expected)
		assertInt32(t,res.counter,expected_hot_counter)
		assertFloat32(t,order.DecayScore,expected_decay_score)
	})

	msg = `
Returns overflow if overflow space is available,
no matching space is available, even if item
will expire in overflow region. Ensures
that the order is set to critical, its decay score is set,
and that its shelf is set to overflow.
`
	t.Run(msg,func(t *testing.T){
		overflow.counter = 1
		hot.counter = 0
		overflow.modifier = 1000
		hot.modifier = 0
		res := order.selectShelf(&shelves,1000)
		expected := overflow
		expected_overflow_counter := int32(0)
		expected_decay_score := order.computeDecayScore(overflow.modifier,1000)
		assertShelf(t,res,expected)
		assertInt32(t,res.counter,expected_overflow_counter)
		assertFloat32(t,order.DecayScore,expected_decay_score)
		assertBoolean(t,order.IsCritical,true)
	})
}
