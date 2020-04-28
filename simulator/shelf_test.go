package simulator

import (
	"bytes"
	"log"
	"testing"
	"time"
)

/*
	Three cases:
		1. We have a critical, eligible value to swap
		2. We have no eligible values to swap due to lack of criticals
		3. swapAssessment's caller shelf is an overflow shelf,
			so there's no freed up space.
	Blocked by: SelectCritical test
*/
func TestSwapAssessment(t *testing.T) {
	msg := `
Eligible Order in overflow
critical map. SwapAssessment should
remove the value from the overflow shelf
contents, remove the value from the overflow
shelf critical list, increase overflow's counter
by one, and move the eligible order into the hot shelf's
contents without changing hot shelf's counter, and
remove the passed in order from the hot shelf (as
it has been picked up).
`

	t.Run(msg, func(t *testing.T) {
		statistics := Statistics{}
		overflowShelf := buildOrderShelf(1, "overflow", 4)
		hotShelf := buildOrderShelf(1, "hot", 1)

		mockNow := mockTimeNow()
		oneSecondAgo := mockNow.Add(time.Second * time.Duration(-1))
		arrivalTime := oneSecondAgo.Add(time.Second * time.Duration(7))
		criticalOrder := foodOrder{Id: "a", Name: "dummy", Temp: "hot", ShelfLife: 12, DecayRate: 1,
			IsCritical: true, placementTime: oneSecondAgo,
			arrivalTime: arrivalTime, shelf: overflowShelf}
		criticalOrder.DecayScore = criticalOrder.computeDecayScore(overflowShelf.modifier, 7*1000)
		overflowShelf.criticals.Set(criticalOrder.Id, &criticalOrder)
		overflowShelf.contents.Set(criticalOrder.Id, &criticalOrder)

		safeOrder := foodOrder{Id: "b", Name: "dummy2", Temp: "hot", ShelfLife: 12, DecayRate: 1,
			IsCritical: false, placementTime: oneSecondAgo,
			arrivalTime: arrivalTime, shelf: hotShelf, DecayScore: 1}
		hotShelf.contents.Set(safeOrder.Id, &safeOrder)
		shelves := orderShelves{overflow: overflowShelf, hot: hotShelf}
		swapOut := bytes.Buffer{}
		swapOutLog := log.New(&swapOut, "", 0)
		args := Config{shelves: &shelves, getNow: mockTimeNow, swapLog: swapOutLog}

		//hot_shelf.swapAssessment(&safe_order,overflow_shelf,&statistics,mockTimeNow)
		hotShelf.swapAssessment(&safeOrder, &statistics, &args)
		hotShelfContents := hotShelf.duplicateContentsToMap(&safeOrder, true)
		hotShelfOrder := castToOrder(hotShelfContents["a"])
		assertOrder(t, hotShelfOrder, &criticalOrder)
		assertInt32(t, hotShelf.counter, int32(1))
		assertInt32(t, overflowShelf.counter, int32(2))
		assertBoolean(t, overflowShelf.contents.IsEmpty(), true)
		assertBoolean(t, overflowShelf.criticals.IsEmpty(), true)
		assertUint64(t, statistics.GetTotalSwapped(), 1)

		expectedSwapOut := `
Swapped Order a from overflow shelf to hot shelf. Old Decay Score: 1.17. New Decay Score: 1.00.
`
		assertStrings(t, swapOut.String(), expectedSwapOut)
	})

	msg = `
No eligible order in overflow.
Hot shelf should have its order removed
and available count increased by one.
`
	t.Run(msg, func(t *testing.T) {
		statistics := Statistics{}
		overflowShelf := buildOrderShelf(1, "overflow", 4)
		hotShelf := buildOrderShelf(1, "hot", 1)

		mockNow := mockTimeNow()
		oneSecondAgo := mockNow.Add(time.Second * time.Duration(-1))
		arrivalTime := oneSecondAgo.Add(time.Second * time.Duration(7))
		safeOrder := foodOrder{Id: "b", Name: "dummy2", Temp: "hot", ShelfLife: 12, DecayRate: 1,
			IsCritical: false, placementTime: oneSecondAgo,
			arrivalTime: arrivalTime, shelf: hotShelf, DecayScore: 1}
		hotShelf.contents.Set(safeOrder.Id, &safeOrder)

		shelves := orderShelves{overflow: overflowShelf, hot: hotShelf}
		swapOut := bytes.Buffer{}
		swapOutLog := log.New(&swapOut, "", 0)
		args := Config{shelves: &shelves, getNow: mockTimeNow, swapLog: swapOutLog}

		hotShelf.swapAssessment(&safeOrder, &statistics, &args)
		assertInt32(t, hotShelf.counter, int32(2))
		assertInt32(t, overflowShelf.counter, int32(1))
		assertBoolean(t, hotShelf.contents.IsEmpty(), true)
		assertUint64(t, statistics.GetTotalSwapped(), 0)
		assertStrings(t, swapOut.String(), "")
	})

	msg = `
Shelf is the overflow shelf.
Overflow shelf should have its order removed
and available count increased by one.
`
	t.Run(msg, func(t *testing.T) {
		statistics := Statistics{}
		overflowShelf := buildOrderShelf(1, "overflow", 4)
		hotShelf := buildOrderShelf(1, "hot", 1)

		mockNow := mockTimeNow()
		oneSecondAgo := mockNow.Add(time.Second * time.Duration(-1))
		arrivalTime := oneSecondAgo.Add(time.Second * time.Duration(7))
		safeOrder := foodOrder{Id: "b", Name: "dummy2", Temp: "cold", ShelfLife: 1000, DecayRate: 1,
			IsCritical: false, placementTime: oneSecondAgo,
			arrivalTime: arrivalTime, shelf: overflowShelf, DecayScore: 1}
		overflowShelf.contents.Set(safeOrder.Id, &safeOrder)

		shelves := orderShelves{overflow: overflowShelf, hot: hotShelf}
		swapOut := bytes.Buffer{}
		swapOutLog := log.New(&swapOut, "", 0)
		args := Config{shelves: &shelves, getNow: mockTimeNow, swapLog: swapOutLog}

		overflowShelf.swapAssessment(&safeOrder, &statistics, &args)
		assertInt32(t, hotShelf.counter, int32(1))
		assertInt32(t, overflowShelf.counter, int32(2))
		assertBoolean(t, hotShelf.contents.IsEmpty(), true)
		assertBoolean(t, overflowShelf.contents.IsEmpty(), true)
		assertUint64(t, statistics.GetTotalSwapped(), 0)
		assertStrings(t, swapOut.String(), "")
	})

}

/*
	For reference, sese the fixtures in TestSwapWillPreserve
	and read the comments there.

	We're going to verify that the order itself
	gets returned
*/
func TestSelectCritical(t *testing.T) {
	t.Run("Eligible critical value, returns an order.", func(t *testing.T) {

		overflowShelf := buildOrderShelf(1, "overflow", 4)
		hotShelf := buildOrderShelf(1, "hot", 1)
		mockNow := mockTimeNow()
		oneSecondAgo := mockNow.Add(time.Second * time.Duration(-1))
		arrivalTime := oneSecondAgo.Add(time.Second * time.Duration(7))
		order := foodOrder{Id: "a", Name: "dummy", Temp: "hot", ShelfLife: 12, DecayRate: 1,
			IsCritical: true, placementTime: oneSecondAgo,
			arrivalTime: arrivalTime, shelf: overflowShelf}
		order.DecayScore = order.computeDecayScore(overflowShelf.modifier, 7*1000)
		overflowShelf.criticals.Set(order.Id, &order)
		res := hotShelf.selectCritical(overflowShelf, mockTimeNow)
		assertOrder(t, res, &order)
	})

	t.Run("No eligible critical values, returns nil.", func(t *testing.T) {
		/*
			This order will expire because the only
			critical orders are for different shelves
			selectCritical is on a hot shelf, we only
			have a cold order in overflow
		*/
		overflowShelf := buildOrderShelf(1, "overflow", 4)
		hotShelf := buildOrderShelf(1, "hot", 1)
		mockNow := mockTimeNow()
		oneSecondAgo := mockNow.Add(time.Second * time.Duration(-1))
		arrivalTime := oneSecondAgo.Add(time.Second * time.Duration(7))
		order := foodOrder{Id: "a", Name: "dummy", Temp: "cold", ShelfLife: 12, DecayRate: 1,
			IsCritical: true, placementTime: oneSecondAgo,
			arrivalTime: arrivalTime, shelf: overflowShelf}
		order.DecayScore = order.computeDecayScore(overflowShelf.modifier, 7*1000)
		overflowShelf.criticals.Set(order.Id, &order)
		res := hotShelf.selectCritical(overflowShelf, mockTimeNow)
		assertOrder(t, res, nil)

	})
}
