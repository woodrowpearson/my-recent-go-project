package simulator

import (
	"math"
	"testing"
	"time"
)


func mockTimeNow() time.Time{
	location,err := time.LoadLocation("UTC")
	check(err)
	return time.Date(2020,1,0,0,0,0,0,location)
}

func mockGetRandRange(int, int) int {
	return 0
}

func assertBoolean(t *testing.T, res bool, expected bool){
	t.Helper()
	if res != expected{
		t.Errorf("res: %t, expected: %t", res,expected)
	}
}

func assertShelf(t *testing.T, res *orderShelf, expected *orderShelf){
	t.Helper()
	if res != expected{
		t.Errorf("received %+v, expected %+v",res,expected)
	}
}

func assertOrder(t *testing.T, res *foodOrder, expected *foodOrder){
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
	almostEqual := math.Abs(float64(res)-float64(expected)) <= 1e-9
	if !almostEqual {
		t.Errorf("received %.3f, expected %.3f",res,expected)
	}
}

func assertUint64(t *testing.T, res uint64, expected uint64){
	t.Helper()
	if res != expected{
		t.Errorf("received %d, expected %d",res,expected)
	}
}
