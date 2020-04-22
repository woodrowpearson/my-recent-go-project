package simulator

import (
	"bytes"
	"log"
	"strings"
	"sync"
	"testing"
)

func TestRunPrimary(t *testing.T){
	/*
		Single integration test case for proving statistics.
		All orders will be flushed simultaneously.
		Purpose of this test is to smoke test the ingestion pipeline
		+

	*/

	ordersAsString := `[
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


	courierOut, courierErr := bytes.Buffer{},bytes.Buffer{}
	dispatchErr, dispatchOut := bytes.Buffer{},bytes.Buffer{}
	inputSource := strings.NewReader(ordersAsString)
	overflowSize,hotSize,coldSize,frozenSize := 15,10,10,10
	// make courier arrive instantaneously
	courierLowerBound,courierUpperBound := 0,0
	ordersPerSecond := 0
	overflowModifier, coldModifier, hotModifier, frozenModifier := 1,1,1,1
	args, err := BuildConfig(
		uint(overflowSize),
		uint(hotSize),
		uint(coldSize),
		uint(frozenSize),
		uint(courierLowerBound),
		uint(courierUpperBound),
		uint(ordersPerSecond),
		uint(overflowModifier),
		uint(coldModifier),
		uint(hotModifier),
		uint(frozenModifier),
		&courierOut,
		&courierErr,
		&dispatchOut,
		&dispatchErr,
		inputSource,
		false,
	)
	check(err)
	args.getRandRange = mockGetRandRange
	statistics := new(Statistics)
	statistics = Run(args,statistics)
	assertStrings(t, courierErr.String(),"")
	assertStrings(t, dispatchErr.String(),"")
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
		courierOut, courierErr := bytes.Buffer{},bytes.Buffer{}
		dispatchErr, dispatchOut := bytes.Buffer{},bytes.Buffer{}
		inputSource := strings.NewReader("Dummy")
		overflowSize,hotSize,coldSize,frozenSize := 1,1,1,1
		// make courier arrive instantaneously
		courierLowerBound,courierUpperBound := 0,0
		ordersPerSecond := 1
		overflowModifier, coldModifier, hotModifier, frozenModifier := 1,1,1,1
		args, err := BuildConfig(
			uint(overflowSize),
			uint(hotSize),
			uint(coldSize),
			uint(frozenSize),
			uint(courierLowerBound),
			uint(courierUpperBound),
			uint(ordersPerSecond),
			uint(overflowModifier),
			uint(coldModifier),
			uint(hotModifier),
			uint(frozenModifier),
			&courierOut,
			&courierErr,
			&dispatchOut,
			&dispatchErr,
			inputSource,
			false,
		)
		check(err)
		args.getRandRange = mockGetRandRange
		order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:200,DecayRate:1}
		dispatch(&order, args,&statistics,&wg)
		wg.Wait()
		outRes := dispatchOut.String()
		errRes := dispatchErr.String()
		expectedOut := `
Dispatched order a to courier.
Current shelf: overflow.
Current shelf contents: [a].
`
		assertStrings(t, errRes,"")
		assertStrings(t, outRes, expectedOut)

	})

	msg = `
Order arrives, but all shelves have 0
capacity. dispatch() logs the discard
message to the dispatch_err io.Writer.
`
	t.Run(msg,func(t *testing.T){
		statistics := Statistics{}
		var wg sync.WaitGroup
		courierOut, courierErr := bytes.Buffer{},bytes.Buffer{}
		dispatchErr, dispatchOut := bytes.Buffer{},bytes.Buffer{}
		inputSource := strings.NewReader("Dummy")
		overflowSize,hotSize,coldSize,frozenSize := 0,0,0,0
		// make courier arrive instantaneously
		courierLowerBound,courierUpperBound := 0,0
		ordersPerSecond := 1
		overflowModifier, coldModifier, hotModifier, frozenModifier := 0,0,0,0
		args, err := BuildConfig(
			uint(overflowSize),
			uint(hotSize),
			uint(coldSize),
			uint(frozenSize),
			uint(courierLowerBound),
			uint(courierUpperBound),
			uint(ordersPerSecond),
			uint(overflowModifier),
			uint(coldModifier),
			uint(hotModifier),
			uint(frozenModifier),
			&courierOut,
			&courierErr,
			&dispatchOut,
			&dispatchErr,
			inputSource,
			false,
		)
		check(err)
		args.getRandRange = mockGetRandRange
		order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:200,DecayRate:1}
		dispatch(&order, args,&statistics,&wg)
		wg.Wait()
		outRes := dispatchOut.String()
		errRes := dispatchErr.String()
		expectedErr := `Order a discarded due to lack of capacity.
`
		assertStrings(t, outRes,"")
		assertStrings(t, errRes, expectedErr)
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
		overflowShelf := buildOrderShelf(1,"overflow",0)
		order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:false,placementTime:mockTimeNow(),
				arrivalTime:mockTimeNow(),shelf: overflowShelf,DecayScore:1.00}
		overflowShelf.contents.Set(order.Id,&order)
		courierOut := bytes.Buffer{}
		courierErr := bytes.Buffer{}
		courierOutLog := log.New(&courierOut,"",0)
		courierErrLog := log.New(&courierErr,"",0)
		go courier(&order, overflowShelf, overflowShelf,
			&statistics,&wg,
			courierOutLog,
			courierErrLog,mockTimeNow)
		wg.Wait()
		expectedOut := `
Courier fetched item a with remaining value of 1.00.
Current shelf: overflow.
Current shelf contents: [].
`
		expectedErr := ""
		outRes := courierOut.String()
		errRes := courierErr.String()
		assertStrings(t, outRes, expectedOut)
		assertStrings(t, errRes, expectedErr)
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
		overflowShelf := buildOrderShelf(1,"overflow",0)
		order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:true,placementTime:mockTimeNow(),
				arrivalTime:mockTimeNow(),shelf: overflowShelf,DecayScore:0.00}
		overflowShelf.contents.Set(order.Id,&order)
		courierOut := bytes.Buffer{}
		courierErr := bytes.Buffer{}
		courierOutLog := log.New(&courierOut,"",0)
		courierErrLog := log.New(&courierErr,"",0)
		go courier(&order, overflowShelf, overflowShelf,
			&statistics,&wg,
			courierOutLog,
			courierErrLog,
			mockTimeNow)
		wg.Wait()
		expectedOut := ""
		expectedErr := `
Discarded item with id a due to expiration value of 0.00.
Current shelf: overflow.
Current shelf contents: [].
`
		outRes := courierOut.String()
		errRes := courierErr.String()
		assertStrings(t, outRes, expectedOut)
		assertStrings(t, errRes, expectedErr)
		assertUint64(t,statistics.GetTotalProcessed(),1)
		assertUint64(t,statistics.GetTotalFailures(),1)
		assertUint64(t,statistics.GetTotalDecayed(),1)
		assertUint64(t,statistics.GetHotDecayed(),1)
	})


}
