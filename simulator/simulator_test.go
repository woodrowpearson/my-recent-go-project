package simulator

import (
	"testing"
	"bytes"
	"sync"
	"strings"
	"log"
)

func TestRunPrimary(t *testing.T){
	/*
		Single integration test case for proving statistics.
		All orders will be flushed simultaneously.
		Purpose of this test is to smoke test the ingestion pipeline
		+
		
	*/

	orders_as_string := `[
  {
    "id": "a8cfcb76-7f24-4420-a5ba-d46dd77bdffd",
    "name": "Banana Split",
    "temp": "frozen",
    "shelfLife": 20,
    "decayRate": 0.63
  },
  {
    "id": "58e9b5fe-3fde-4a27-8e98-682e58a4a65d",
    "name": "McFlury",
    "temp": "frozen",
    "shelfLife": 375,
    "decayRate": 0.4
  },
  {
    "id": "2ec069e3-576f-48eb-869f-74a540ef840c",
    "name": "Acai Bowl",
    "temp": "cold",
    "shelfLife": 249,
    "decayRate": 0.3
  },
  {
    "id": "690b85f7-8c7d-4337-bd02-04e04454c826",
    "name": "Yogurt",
    "temp": "cold",
    "shelfLife": 263,
    "decayRate": 0.37
  }
]`


	receivedOut,swapOut := bytes.Buffer{},bytes.Buffer{}
	courier_out,courier_err := bytes.Buffer{},bytes.Buffer{}
	dispatch_err,dispatch_out := bytes.Buffer{},bytes.Buffer{}
	inputSource := strings.NewReader(orders_as_string)
	overflowSize,hotSize,coldSize,frozenSize := 15,10,10,10
	// make courier arrive instantaneously
	courierLowerBound,courierUpperBound := 0,0
	ordersPerSecond := 0
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
		&receivedOut,
		&swapOut,
		&courier_out,
		&courier_err,
		&dispatch_out,
		&dispatch_err,
		inputSource,
		false,
	)
	check(err)
	args.getRandRange = mockGetRandRange
	statistics := new(Statistics)
	statistics = Run(args,statistics)

	expectedReceivedOut := `
Received Order a8cfcb76-7f24-4420-a5ba-d46dd77bdffd. Name: Banana Split. Temp: frozen. Shelf Life: 20. Decay Rate: 0.63.

Received Order 58e9b5fe-3fde-4a27-8e98-682e58a4a65d. Name: McFlury. Temp: frozen. Shelf Life: 375. Decay Rate: 0.40.

Received Order 2ec069e3-576f-48eb-869f-74a540ef840c. Name: Acai Bowl. Temp: cold. Shelf Life: 249. Decay Rate: 0.30.

Received Order 690b85f7-8c7d-4337-bd02-04e04454c826. Name: Yogurt. Temp: cold. Shelf Life: 263. Decay Rate: 0.37.
`

	// We cannot guarantee ordering of dispatch_out and courier_out.
	assertStrings(t,courier_err.String(),"")
	assertStrings(t,dispatch_err.String(),"")
	assertStrings(t,receivedOut.String(),expectedReceivedOut)
	assertUint64(t,statistics.GetTotalProcessed(),4)
	assertUint64(t,statistics.GetTotalSuccesses(),4)
	assertUint64(t,statistics.GetColdSuccesses(),2)
	assertUint64(t,statistics.GetFrozenSuccesses(),2)
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
		statistics := Statistics{}
		var wg sync.WaitGroup
		receivedOut,swapOut := bytes.Buffer{},bytes.Buffer{}
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
			&receivedOut,
			&swapOut,
			&courier_out,
			&courier_err,
			&dispatch_out,
			&dispatch_err,
			inputSource,
			false,
		)
		check(err)
		args.getRandRange = mockGetRandRange
		order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:200,DecayRate:1}
		dispatch(&order, args,&statistics,&wg)
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
		statistics := Statistics{}
		var wg sync.WaitGroup
		receivedOut,swapOut := bytes.Buffer{},bytes.Buffer{}
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
			&receivedOut,
			&swapOut,
			&courier_out,
			&courier_err,
			&dispatch_out,
			&dispatch_err,
			inputSource,
			false,
		)
		check(err)
		args.getRandRange = mockGetRandRange
		order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:200,DecayRate:1}
		dispatch(&order, args,&statistics,&wg)
		wg.Wait()
		out_res := dispatch_out.String()
		err_res := dispatch_err.String()
		expected_err := `Order a discarded due to lack of capacity.
`
		assertStrings(t,out_res,"")
		assertStrings(t,err_res,expected_err)
		assertUint64(t,statistics.GetTotalFailures(),1)
		assertUint64(t,statistics.GetTotalDiscarded(),1)
		assertUint64(t,statistics.GetTotalProcessed(),1)
		assertUint64(t,statistics.GetHotDiscarded(),1)

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
		statistics := Statistics{}
		var wg sync.WaitGroup
		wg.Add(1)
		overflow_shelf := buildOrderShelf(1,"overflow",0)
		order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:false,placementTime:mockTimeNow(),
				arrivalTime:mockTimeNow(),shelf:overflow_shelf,DecayScore:1.00}
		overflow_shelf.contents.Set(order.Id,&order)
		courier_out := bytes.Buffer{}
		courier_err := bytes.Buffer{}
		courier_out_log := log.New(&courier_out,"",0)
		courier_err_log := log.New(&courier_err,"",0)
		shelves := orderShelves{overflow:overflow_shelf}
		args := SimulatorConfig{shelves:&shelves,getNow:mockTimeNow,
				courier_out_log:courier_out_log,
				courier_err_log:courier_err_log}
		go courier(&order,overflow_shelf,&statistics,&wg,&args)
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
		assertUint64(t,statistics.GetTotalProcessed(),1)
		assertUint64(t,statistics.GetTotalSuccesses(),1)
		assertUint64(t,statistics.GetHotSuccesses(),1)
	})

	msg = `
Order arrives to courier,
and its a critical order.
Write the failure to the error pipeline.
Wait group should be finished.
`
	t.Run(msg, func(t *testing.T){
		statistics := Statistics{}
		var wg sync.WaitGroup
		wg.Add(1)
		overflow_shelf := buildOrderShelf(1,"overflow",0)
		order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:true,placementTime:mockTimeNow(),
				arrivalTime:mockTimeNow(),shelf:overflow_shelf,DecayScore:0.00}
		overflow_shelf.contents.Set(order.Id,&order)
		courier_out := bytes.Buffer{}
		courier_err := bytes.Buffer{}
		courier_out_log := log.New(&courier_out,"",0)
		courier_err_log := log.New(&courier_err,"",0)
		shelves := orderShelves{overflow:overflow_shelf}
		args := SimulatorConfig{shelves:&shelves,getNow:mockTimeNow,
				courier_out_log:courier_out_log,
				courier_err_log:courier_err_log}
		go courier(&order,overflow_shelf,&statistics,&wg,&args)
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
		assertUint64(t,statistics.GetTotalProcessed(),1)
		assertUint64(t,statistics.GetTotalFailures(),1)
		assertUint64(t,statistics.GetTotalDecayed(),1)
		assertUint64(t,statistics.GetHotDecayed(),1)
	})


}
