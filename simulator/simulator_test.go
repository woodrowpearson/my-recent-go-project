package simulator

import (
	"testing"
	"bytes"
	"sync"
	"strings"
//	"fmt"
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
		Only tests logging. State change tests
		are handled in the shelf and order test suites.
	*/

	msg := `
Order arrives with an available shelf.
dispatch() logs the contents of the shelf
to the dispatch_out io.Writer.
`
	t.Run(msg, func(t *testing.T){
		var wg sync.WaitGroup
		courier_out,courier_err := bytes.Buffer{},bytes.Buffer{}
		dispatch_err,dispatch_out := bytes.Buffer{},bytes.Buffer{}
		inputSource := strings.NewReader("Dummy")
		overflowSize,hotSize,coldSize,frozenSize := 1,1,1,1
		// make courier arrive instantaneously
		courierLowerBound,courierUpperBound := 0,0
		ordersPerSecond := 1
		overflow_modifier,cold_modifier,hot_modifier,frozen_modifier := 1,1,1,1
		args, err := BuildConfig(
			uint(overflowSize),
			uint(hotSize),
			uint(coldSize),
			uint(frozenSize),
			uint(courierLowerBound),
			uint(courierUpperBound),
			uint(ordersPerSecond),
			uint(overflow_modifier),
			uint(cold_modifier),
			uint(hot_modifier),
			uint(frozen_modifier),
			&courier_out,
			&courier_err,
			&dispatch_out,
			&dispatch_err,
			inputSource,
			1,
		)
		check(err)
		args.getRandRange = mockGetRandRange
		order := Order{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:200,DecayRate:1}
		dispatch(&order, args,&wg)
		wg.Wait()
		out_res := dispatch_out.String()
		err_res := dispatch_err.String()
		expected_out := `
Dispatched order a to courier.
Current shelf: overflow.
Current shelf contents: [a].
`
		assertStrings(t,err_res,"")
		assertStrings(t,out_res,expected_out)

	})

	msg = `
Order arrives, but all shelves have 0
capacity. dispatch() logs the discard
message to the dispatch_err io.Writer.
`
	t.Run(msg,func(t *testing.T){
		var wg sync.WaitGroup
		courier_out,courier_err := bytes.Buffer{},bytes.Buffer{}
		dispatch_err,dispatch_out := bytes.Buffer{},bytes.Buffer{}
		inputSource := strings.NewReader("Dummy")
		overflowSize,hotSize,coldSize,frozenSize := 0,0,0,0
		// make courier arrive instantaneously
		courierLowerBound,courierUpperBound := 0,0
		ordersPerSecond := 1
		overflow_modifier,cold_modifier,hot_modifier,frozen_modifier := 0,0,0,0
		args, err := BuildConfig(
			uint(overflowSize),
			uint(hotSize),
			uint(coldSize),
			uint(frozenSize),
			uint(courierLowerBound),
			uint(courierUpperBound),
			uint(ordersPerSecond),
			uint(overflow_modifier),
			uint(cold_modifier),
			uint(hot_modifier),
			uint(frozen_modifier),
			&courier_out,
			&courier_err,
			&dispatch_out,
			&dispatch_err,
			inputSource,
			1,
		)
		check(err)
		args.getRandRange = mockGetRandRange
		order := Order{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:200,DecayRate:1}
		dispatch(&order, args,&wg)
		wg.Wait()
		out_res := dispatch_out.String()
		err_res := dispatch_err.String()
		expected_err := `Order a discarded due to lack of capacity.
`
		assertStrings(t,out_res,"")
		assertStrings(t,err_res,expected_err)

	})

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
Current shelf contents: [].
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
Current shelf contents: [].
`
		out_res := courier_out.String()
		err_res := courier_err.String()
		assertStrings(t,out_res,expected_out)
		assertStrings(t,err_res,expected_err)
	})


}
