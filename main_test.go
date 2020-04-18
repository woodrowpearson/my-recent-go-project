package main

import (
	"testing"
//	"reflect"
//	"fmt"
)

func assertShelf(t *testing.T, res *Shelf, expected *Shelf){
	t.Helper()
	if res != expected{
		t.Errorf("received %+v, expected %+v",res,expected)
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
}
