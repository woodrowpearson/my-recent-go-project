package simulator

import (
	"testing"
	"time"
//	"fmt"
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

}



func TestSelectCritical(t *testing.T){


	t.Run("Eligible critical value, returns an order.", func(t *testing.T){
		/*
			For reference, sese the fixtures in TestSwapWillPreserve
			and read the comments there.

			We're going to verify that the order itself
			gets returned 
		*/
		overflow_shelf := buildShelf(1,"overflow",4)
		hot_shelf := buildShelf(1,"hot",1)
		mock_now := mockTimeNow()
		one_second_ago := mock_now.Add(time.Second*time.Duration(-1))
		arrival_time := one_second_ago.Add(time.Second*time.Duration(7))
		order := Order{Id:"a",Name:"dummy",Temp:"hot",ShelfLife:12,DecayRate:1,
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
		overflow_shelf := buildShelf(1,"overflow",4)
		hot_shelf := buildShelf(1,"hot",1)
		mock_now := mockTimeNow()
		one_second_ago := mock_now.Add(time.Second*time.Duration(-1))
		arrival_time := one_second_ago.Add(time.Second*time.Duration(7))
		order := Order{Id:"a",Name:"dummy",Temp:"cold",ShelfLife:12,DecayRate:1,
				IsCritical:true,placementTime:one_second_ago,
				arrivalTime:arrival_time,shelf:overflow_shelf}
		order.DecayScore = order.computeDecayScore(overflow_shelf.modifier,7*1000)
		overflow_shelf.criticals.Set(order.Id,&order)
		res := hot_shelf.selectCritical(overflow_shelf,mockTimeNow)
		assertOrder(t,res,nil)

	})


}
