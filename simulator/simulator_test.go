package simulator

import (
	"testing"
	"bytes"
//	"reflect"
//	"fmt"
	"sync"
)

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
	t.Run("selects overflow aggressively", func(t *testing.T){
		expected := overflow
		res := selectShelf(&order,&shelves)
		assertShelf(t,res,expected)
	})
	t.Run("moves to matching shelf if overflow is empty",
		func(t *testing.T){
		overflow.counter = 0
		expected := hot
		res := selectShelf(&order,&shelves)
		assertShelf(t,res,expected)
	})
	t.Run("returns dead if no capacity", func(t *testing.T){
		overflow.counter = 0
		hot.counter = 0
		expected := dead
		res := selectShelf(&order,&shelves)
		assertShelf(t,res,expected)
	})
	// TODO: add more tests for the other options beyond hot assuming
	// that we don't end up screwing with selectShelf's heuristics
}

func TestCourier(t *testing.T){
	t.Run("happy path. successful pickup", func(t *testing.T){
		var wg sync.WaitGroup
		wg.Add(1)
		order := Order{Id:"a",Name:"dummy",Temp:"hot",
			ShelfLife: 200, DecayRate: 0.25}
		hot := buildShelf(1,"hot",0)
		arrival_time,shelf_idx := 0,0
		courier_out := bytes.Buffer{}
		courier_err := bytes.Buffer{}
		courier(order, hot,arrival_time,
			&wg, shelf_idx,
			&courier_out, &courier_err)
		expected_out := `
Courier fetched item a with remaining value of 1.00.
Current shelf: hot.
Current shelf contents: [].
`
		expected_err := ""
		out_res := courier_out.String()
		err_res := courier_err.String()
		assertStrings(t,out_res,expected_out)
		assertStrings(t,err_res,expected_err)
		wg.Wait()
	})
	t.Run("sad path. decayed out", func(t *testing.T){
		var wg sync.WaitGroup
		wg.Add(1)
		order := Order{Id:"a",Name:"dummy",Temp:"hot",
			ShelfLife: 0, DecayRate: 0.25}
		hot := buildShelf(1,"hot",0)
		arrival_time,shelf_idx := 0,0
		courier_out := bytes.Buffer{}
		courier_err := bytes.Buffer{}
		courier(order, hot,arrival_time,
			&wg, shelf_idx,
			&courier_out, &courier_err)
		expected_out := ""
		expected_err := `
Discarded item with id a due to expiration value of 0.00.
Current shelf: hot.
Current shelf contents: [].
`
		out_res := courier_out.String()
		err_res := courier_err.String()
		assertStrings(t,out_res,expected_out)
		assertStrings(t,err_res,expected_err)
	})


}
