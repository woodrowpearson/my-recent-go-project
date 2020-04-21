package simulator

import (
	"testing"
	"time"
)


func TestSwapAssessment(t *testing.T){
	/*
		Three cases:
			1. We have a critical, eligible value to swap
			2. We have no eligible values to swap due to lack of criticals
			3. swapAssessment's caller shelf is an overflow shelf,
				so there's no freed up space.
		Blocked by: SelectCritical test
	*/
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

	t.Run(msg, func(t *testing.T){
		statistics := Statistics{}
		overflow_shelf := buildOrderShelf(1,"overflow",4)
		hot_shelf := buildOrderShelf(1,"hot",1)

		mock_now := mockTimeNow()
		one_second_ago := mock_now.Add(time.Second*time.Duration(-1))
		arrival_time := one_second_ago.Add(time.Second*time.Duration(7))
		critical_order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:true,placementTime:one_second_ago,
				arrivalTime:arrival_time,shelf:overflow_shelf}
		critical_order.DecayScore = critical_order.computeDecayScore(overflow_shelf.modifier,7*1000)
		overflow_shelf.criticals.Set(critical_order.Id,&critical_order)
		overflow_shelf.contents.Set(critical_order.Id,&critical_order)

		safe_order := foodOrder{Id:"b",Name:"dummy2",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:false,placementTime:one_second_ago,
				arrivalTime:arrival_time,shelf:hot_shelf,DecayScore:1}
		hot_shelf.contents.Set(safe_order.Id,&safe_order)

		hot_shelf.swapAssessment(&safe_order,overflow_shelf,&statistics,mockTimeNow)
		hot_shelf_contents := hot_shelf.duplicateContentsToMap(&safe_order,true)
		hot_shelf_order := castToOrder(hot_shelf_contents["a"])
		assertOrder(t,hot_shelf_order,&critical_order)
		assertInt32(t,hot_shelf.counter,int32(1))
		assertInt32(t,overflow_shelf.counter, int32(2))
		assertBoolean(t,overflow_shelf.contents.IsEmpty(),true)
		assertBoolean(t,overflow_shelf.criticals.IsEmpty(),true)
		assertUint64(t,statistics.GetTotalSwapped(),1)
	})


	msg = `
No eligible order in overflow.
Hot shelf should have its order removed
and available count increased by one.
`
	t.Run(msg,func(t *testing.T){
		statistics := Statistics{}
		overflow_shelf := buildOrderShelf(1,"overflow",4)
		hot_shelf := buildOrderShelf(1,"hot",1)

		mock_now := mockTimeNow()
		one_second_ago := mock_now.Add(time.Second*time.Duration(-1))
		arrival_time := one_second_ago.Add(time.Second*time.Duration(7))
		safe_order := foodOrder{Id:"b",Name:"dummy2",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:false,placementTime:one_second_ago,
				arrivalTime:arrival_time,shelf:hot_shelf,DecayScore:1}
		hot_shelf.contents.Set(safe_order.Id,&safe_order)
		hot_shelf.swapAssessment(&safe_order,overflow_shelf,&statistics,mockTimeNow)
		assertInt32(t,hot_shelf.counter,int32(2))
		assertInt32(t,overflow_shelf.counter,int32(1))
		assertBoolean(t,hot_shelf.contents.IsEmpty(),true)
		assertUint64(t,statistics.GetTotalSwapped(),0)
	})

	msg = `
Shelf is the overflow shelf.
Overflow shelf should have its order removed
and available count increased by one.
`
	t.Run(msg, func(t *testing.T){
		statistics := Statistics{}
		overflow_shelf := buildOrderShelf(1,"overflow",4)
		hot_shelf := buildOrderShelf(1,"hot",1)

		mock_now := mockTimeNow()
		one_second_ago := mock_now.Add(time.Second*time.Duration(-1))
		arrival_time := one_second_ago.Add(time.Second*time.Duration(7))
		safe_order := foodOrder{Id:"b",Name:"dummy2",Temp:"cold",ShelfLife:1000,DecayRate:1,
				IsCritical:false,placementTime:one_second_ago,
				arrivalTime:arrival_time,shelf:overflow_shelf,DecayScore:1}
		overflow_shelf.contents.Set(safe_order.Id,&safe_order)
		overflow_shelf.swapAssessment(&safe_order,overflow_shelf,&statistics,mockTimeNow)
		assertInt32(t,hot_shelf.counter,int32(1))
		assertInt32(t,overflow_shelf.counter,int32(2))
		assertBoolean(t,hot_shelf.contents.IsEmpty(),true)
		assertBoolean(t,overflow_shelf.contents.IsEmpty(),true)
		assertUint64(t,statistics.GetTotalSwapped(),0)
	})

}



func TestSelectCritical(t *testing.T){


	t.Run("Eligible critical value, returns an order.", func(t *testing.T){
		/*
			For reference, sese the fixtures in TestSwapWillPreserve
			and read the comments there.

			We're going to verify that the order itself
			gets returned 
		*/
		overflow_shelf := buildOrderShelf(1,"overflow",4)
		hot_shelf := buildOrderShelf(1,"hot",1)
		mock_now := mockTimeNow()
		one_second_ago := mock_now.Add(time.Second*time.Duration(-1))
		arrival_time := one_second_ago.Add(time.Second*time.Duration(7))
		order := foodOrder{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:12,DecayRate:1,
				IsCritical:true,placementTime:one_second_ago,
				arrivalTime:arrival_time,shelf:overflow_shelf}
		order.DecayScore = order.computeDecayScore(overflow_shelf.modifier,7*1000)
		overflow_shelf.criticals.Set(order.Id,&order)
		res := hot_shelf.selectCritical(overflow_shelf,mockTimeNow)
		assertOrder(t,res,&order)
	})

	t.Run("No eligible critical values, returns nil.", func(t *testing.T){
		/*
			This order will expire because the only
			critical orders are for different shelves
			selectCritical is on a hot shelf, we only
			have a cold order in overflow
		*/
		overflow_shelf := buildOrderShelf(1,"overflow",4)
		hot_shelf := buildOrderShelf(1,"hot",1)
		mock_now := mockTimeNow()
		one_second_ago := mock_now.Add(time.Second*time.Duration(-1))
		arrival_time := one_second_ago.Add(time.Second*time.Duration(7))
		order := foodOrder{Id:"a",Name:"dummy",Temp:"cold",ShelfLife:12,DecayRate:1,
				IsCritical:true,placementTime:one_second_ago,
				arrivalTime:arrival_time,shelf:overflow_shelf}
		order.DecayScore = order.computeDecayScore(overflow_shelf.modifier,7*1000)
		overflow_shelf.criticals.Set(order.Id,&order)
		res := hot_shelf.selectCritical(overflow_shelf,mockTimeNow)
		assertOrder(t,res,nil)

	})
}
