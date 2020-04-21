package simulator

import (
	"testing"
	"math"
	"time"
)


func mockTimeNow() time.Time{
	location,err := time.LoadLocation("UTC")
	check(err)
	return time.Date(2020,1,0,0,0,0,0,location)
}


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

func assertOrder(t *testing.T, res *Order, expected *Order){
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