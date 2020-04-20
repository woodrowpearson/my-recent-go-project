package simulator

import "testing"

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


//func TestCourier(t *testing.T){
//
//	/*
//		Two cases:
//			1. order arrives and is critical
//			2. order arrives and is not critical.
//		Tests for state changes are on TestSwapAssessment
//	*/
//
//
//	t.Run("happy path. successful pickup", func(t *testing.T){
//		var wg sync.WaitGroup
//		wg.Add(1)
//		order := Order{Id:"a",Name:"dummy",Temp:"hot",
//			ShelfLife: 200, DecayRate: 0.25}
//		hot := buildShelf(1,"hot",0)
//		arrival_time,shelf_idx := 0,0
//		courier_out := bytes.Buffer{}
//		courier_err := bytes.Buffer{}
//		courier(order, hot,arrival_time,
//			&wg, shelf_idx,
//			&courier_out, &courier_err)
//		expected_out := `
//Courier fetched item a with remaining value of 1.00.
//Current shelf: hot.
//Current shelf contents: [].
//`
//		expected_err := ""
//		out_res := courier_out.String()
//		err_res := courier_err.String()
//		assertStrings(t,out_res,expected_out)
//		assertStrings(t,err_res,expected_err)
//		wg.Wait()
//	})
//	t.Run("sad path. decayed out", func(t *testing.T){
//		var wg sync.WaitGroup
//		wg.Add(1)
//		order := Order{Id:"a",Name:"dummy",Temp:"hot",
//			ShelfLife: 0, DecayRate: 0.25}
//		hot := buildShelf(1,"hot",0)
//		arrival_time,shelf_idx := 0,0
//		courier_out := bytes.Buffer{}
//		courier_err := bytes.Buffer{}
//		courier(order, hot,arrival_time,
//			&wg, shelf_idx,
//			&courier_out, &courier_err)
//		expected_out := ""
//		expected_err := `
//Discarded item with id a due to expiration value of 0.00.
//Current shelf: hot.
//Current shelf contents: [].
//`
//		out_res := courier_out.String()
//		err_res := courier_err.String()
//		assertStrings(t,out_res,expected_out)
//		assertStrings(t,err_res,expected_err)
//	})
//
//
//}
