package simulator

import (
	"testing"
	"bytes"
	"sync"
)

func TestRunPrimary(t *testing.T){
	/*
		Single test case. needs to run an integration test
		that sends data out to four channels.
		When complete, all shelves should be empty.
		NOTE that we need a successful shelf-swap to occur in this.

	*/

}


func TestDispatch(t *testing.T){
	/*
		Two cases:
			1. discarded due to dead shelf.
			2. sent for courier due to available shelf.
	*/


}


func TestCourier(t *testing.T){

	/*
		Two cases:
			1. order arrives and is critical
			2. order arrives and is not critical.
		Tests for state changes are on TestSwapAssessment
	*/

	msg := `
Order arrives to courier,
and it is a non-critical order.
Write the success to the success pipeline.
Wait group should be finished.
`
	t.Run(msg, func(t *testing.T){
		var wg sync.WaitGroup
		wg.Add(1)
		overflow_shelf := buildShelf(1,"overflow",0)
		order := Order{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:false,placementTime:mockTimeNow(),
				arrivalTime:mockTimeNow(),shelf:overflow_shelf,DecayScore:1.00}
		overflow_shelf.contents.Set(order.Id,&order)
		courier_out := bytes.Buffer{}
		courier_err := bytes.Buffer{}
		go courier(&order, overflow_shelf,overflow_shelf,&wg,&courier_out, &courier_err,mockTimeNow)
		wg.Wait()
		expected_out := `
Courier fetched item a with remaining value of 1.00.
Current shelf: overflow.
Current shelf contents: map[].
`
		expected_err := ""
		out_res := courier_out.String()
		err_res := courier_err.String()
		assertStrings(t,out_res,expected_out)
		assertStrings(t,err_res,expected_err)
		wg.Wait()
	})

	msg = `
Order arrives to courier,
and its a critical order.
Write the failure to the error pipeline.
Wait group should be finished.
`
	t.Run(msg, func(t *testing.T){
		var wg sync.WaitGroup
		wg.Add(1)
		overflow_shelf := buildShelf(1,"overflow",0)
		order := Order{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:true,placementTime:mockTimeNow(),
				arrivalTime:mockTimeNow(),shelf:overflow_shelf,DecayScore:0.00}
		overflow_shelf.contents.Set(order.Id,&order)
		courier_out := bytes.Buffer{}
		courier_err := bytes.Buffer{}
		go courier(&order, overflow_shelf,overflow_shelf,&wg,&courier_out, &courier_err,mockTimeNow)
		wg.Wait()
		expected_out := ""
		expected_err := `
Discarded item with id a due to expiration value of 0.00.
Current shelf: overflow.
Current shelf contents: map[].
`
		out_res := courier_out.String()
		err_res := courier_err.String()
		assertStrings(t,out_res,expected_out)
		assertStrings(t,err_res,expected_err)
	})


}
