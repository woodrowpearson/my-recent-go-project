package main

import (
	"flag"
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"io"
	"os"
	"sync"
	"math/rand"
)

func check(e error){
	if e != nil{
		panic(e)
	}
}


func computeDecayStatus(order *Order,shelf *Shelf, arrival_time int) float32{
	a := float32(order.ShelfLife)
	b := order.DecayRate*float32(arrival_time)*float32(shelf.modifier)
	if (a <= 0){
		return a
	}
	value := (a-b)/a
	return value
}


func courier(order Order, shelf *Shelf, arrival_time int,
		wg *sync.WaitGroup,shelf_idx int,
		courier_out io.Writer,courier_err io.Writer){
	// TODO: ON TESTS, PASS IN ARRIVAL TIME AS 0
	time.Sleep(time.Duration(1000*arrival_time)*time.Millisecond)
	// END BLOCK
	value := computeDecayStatus(&order,shelf,arrival_time)
	shelf.incrementAndUpdate(shelf_idx)
	/*
	In Linux, thread safety is assured in file access:
	https://stackoverflow.com/questions/29981050/concurrent-writing-to-a-file`
	*/
	if (value <= 0){
		fmt.Fprintf(courier_err,PickupErrMsg,order.Id,value,shelf.name,shelf.item_array)
	} else {
		fmt.Fprintf(courier_out,PickupSuccessMsg,order.Id,value,shelf.name,shelf.item_array)
	}
	// END BLOCK
	wg.Done()
}


func selectShelf(o *Order,s *Shelves) *Shelf {
	if (s.overflow.counter > 0){
		return s.overflow
	} else if (o.Temp == "cold" && s.cold.counter > 0){
		return s.cold
	} else if (o.Temp == "hot" && s.hot.counter > 0){
		return s.hot
	} else if (o.Temp == "frozen" &&
		s.frozen.counter > 0){
		return s.frozen
	}
	return s.dead
}



func runQueue(args *SimulatorConfig){

	fmt.Println(args)
	var orders []Order
	// TODO: move this to a streaming json parse
	inputFile, err := os.Open("orders.json")
	check(err)
	fmt.Println("opened file")
	defer inputFile.Close()
	byteArray, err := ioutil.ReadAll(inputFile)
	check(err)
	json.Unmarshal(byteArray, &orders)
	arrlen := uint(len(orders))
	// END BLOCK
	var wg sync.WaitGroup
	criticality_arr := make([]Order, args.orders_per_second)
	for i := uint(0); i < arrlen; i += args.orders_per_second {
		/*
			TODO: before dispatching, sort the items
			by criticality (i.e. longest arrival time)
			We'll want to compute the score for the order 
			at instantiation. 
			TODO: find an equivalent of python's bisect
			function for inserting into the array in a sorted manner
			TODO: MAKE THE CRITICALITY SORT A SEPARATE FUNCTION AND TEST IT
			NOTE: the problem with sorting by criticality here is
			that it makes our loop n^2 instead of o(n), since we iterate over
			each item twice effectively.
		*/
		for j := uint(0); j < args.orders_per_second && i+j < arrlen; j++ {
			order := orders[i+j]
			shelf_idx := -1
			arrival := rand.Intn(
				int(args.courier_upper_bound -
				args.courier_lower_bound)) +
				int(args.courier_lower_bound)
			shelf := selectShelf(&order,args.shelves)
			// TODO: MOVE THIS TO OUTSIDE OF THE J LOOP.
			if (shelf != args.shelves.dead){
				wg.Add(1)
				shelf_idx,err = shelf.decrementAndUpdate(order.Id)
				check(err)
				fmt.Fprintf(args.dispatch_out,DispatchSuccessMsg, order.Id,
					shelf.name, shelf.item_array)
				go courier(order,shelf,arrival,&wg,shelf_idx,args.courier_out,args.courier_err)
			} else {
				fmt.Fprintf(args.dispatch_err,DispatchErrMsg,order.Id)
			}
		}
		fmt.Println(criticality_arr)
		time.Sleep(args.second_value*time.Millisecond)
	}
	wg.Wait()
	fmt.Println("complete")
}

func main(){
	/*
		TODO: clean up style stuff. I dont know what the rules
		are for formatting and camelcase vs snakecase.
	*/

	overflowSize := flag.Uint("overflow_size", 15,ShelfSizePrompt)
	hotSize := flag.Uint("hot_size", 10,ShelfSizePrompt)
	coldSize := flag.Uint("cold_size", 10,ShelfSizePrompt)
	frozenSize := flag.Uint("frozen_size", 10,ShelfSizePrompt)

	overflow_modifier := flag.Uint("overflow_modifier",2,
			ShelfModifierPrompt)
	cold_modifier := flag.Uint("cold_modifier",1,
			ShelfModifierPrompt)
	hot_modifier := flag.Uint("hot_modifier",1,
			ShelfModifierPrompt)
	frozen_modifier := flag.Uint("frozen_modifier",1,
			ShelfModifierPrompt)



	courierLowerBound := flag.Uint("courier_lower_bound", 2, CourierPrompt)
	courierUpperBound := flag.Uint("courier_upper_bound",6,CourierPrompt)

	ordersPerSecond := flag.Uint("orders_per_second",2,OrderRatePrompt)
	flag.Parse()
	if (*courierLowerBound > *courierUpperBound ||
		 *courierLowerBound < 1 ||
		*courierUpperBound < 1){
		fmt.Println(CourierPrompt)
		os.Exit(1)
	}
	if (*ordersPerSecond < 1){
		fmt.Println(OrderRatePrompt)
		os.Exit(1)
	}
	// TODO: add CLI args for logfile locations.
	courier_out, err := os.Create("courier_out.log")
	check(err)
	defer courier_out.Close()
	courier_err, err := os.Create("courier_err.log")
	check(err)
	defer courier_out.Close()
	dispatch_out, err := os.Create("dispatch_out.log")
	check(err)
	defer dispatch_out.Close()
	dispatch_err, err := os.Create("dispatch_err.log")
	check(err)
	defer courier_out.Close()
	// END BLOCK
	overflow := buildShelf(*overflowSize,"overflow",
			*overflow_modifier)
	cold := buildShelf(*coldSize, "cold",*cold_modifier)
	hot := buildShelf(*hotSize,"hot",*hot_modifier)
	frozen := buildShelf(*frozenSize,"frozen",*frozen_modifier)
	dead := buildShelf(1,"dead",0)
	shelves := Shelves{overflow:overflow,cold:cold,frozen:frozen,
			hot:hot,dead:dead}
	args := SimulatorConfig{
		overflow_size:*overflowSize,
		hot_size: *hotSize,
		cold_size: *coldSize,
		frozen_size: *frozenSize,
		courier_lower_bound: *courierLowerBound,
		courier_upper_bound: *courierUpperBound,
		orders_per_second: *ordersPerSecond,
		overflow_modifier: *overflow_modifier,
		cold_modifier: *cold_modifier,
		hot_modifier: *hot_modifier,
		frozen_modifier: *frozen_modifier,
		courier_out:courier_out,
		courier_err:courier_err,
		dispatch_out:dispatch_out,
		dispatch_err:dispatch_err,
		second_value: 1000,
		shelves: &shelves,
	}
	fmt.Printf("Configuration: %+v\n", args)
	runQueue(&args)
}
