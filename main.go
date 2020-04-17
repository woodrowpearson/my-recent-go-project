package main

import (
	"flag"
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"math/rand"
)

type Order struct {
	Id string
	Name string
	Temp string// this should be an enum. TODO: Does Go have enums?
	ShelfLife uint
	DecayRate float32
}


// TODO: make a constructor function for this
// and use it to clean up the courier, shelf selection,
// and decrement functions
type Shelf struct {
	Counter int32
	ItemArray []string
	Name string
}


type PrimaryArgs struct {
	overflow_size uint
	hot_size uint
	cold_size uint
	frozen_size uint
	courier_lower_bound uint
	courier_upper_bound uint
	orders_per_second uint
}


func courier(order Order,
		counter *int32, arrival_time int,
		modifier uint, wg *sync.WaitGroup, shelf []string,target_idx int){
	// NOTE: slices are reference types, and due to the specifics of the problem,
	// we cannot have a scenario where two coroutines are attempting to access
	// the same location at the same time.
	time.Sleep(time.Duration(1000*arrival_time)*time.Millisecond)
	a := float32(order.ShelfLife)
	b := order.DecayRate*float32(arrival_time) * float32(modifier)
	value := (a - b)/a
	// remove item from shelf in either scenario
	// need to log the score in either scenario
	atomic.AddInt32(counter,1)
	shelf[target_idx] = ""
	// inform waitgroup that the coro is finished.
	wg.Done()
	fmt.Println("shelf:",shelf)
	if (value <= 0){
		fmt.Printf("Discarded item due to expiration")
		fmt.Printf("current shelf contents: %s\n", shelf)
	} else {
		fmt.Printf("Courier fetched item %s with remaining value of %.2f\n", order.Id, value)
		fmt.Printf("current shelf contents: %s\n", shelf)
	}
}

// TODO: add in a return value pointer for the shelf itself. This is sloppy
func selectShelf(order *Order,over_ct *int32,
		cold_ct *int32,
		hot_ct *int32, frozen_ct *int32,
		dead *int32) *int32{
	// TODO: add in moving average selection here
	// TODO: make this return an enum instead of a number

	if (*over_ct > 0){
		return over_ct
	}
	if (*cold_ct > 0 && order.Temp == "cold"){
		return cold_ct
	} else if (*hot_ct > 0 && order.Temp == "hot"){
		return hot_ct
	} else if (*frozen_ct > 0 && order.Temp == "frozen"){
		return frozen_ct
	} else {
		fmt.Println(order)
		panic("unknown temp")
	}
	return dead
}

func decrement(ct *int32){
	// TODO: add in a pointer to the array
	// so we can store the ID
	atomic.AddInt32(ct,-1)

}

func placeInArray(target_arr []string, value string) int {
	for i := 0; i < len(target_arr); i++ {
		if (target_arr[i] == ""){
			target_arr[i] = value
			return i
		}
	}
	return -1
}

/*
	TODO:
		- use CLI args to inform dispatch, ingestion, and shelves
		- account for order rates that aren't 2
		- move the shelf and count stuff to the Shelf type.

*/
func runQueue(args *PrimaryArgs){

	fmt.Println(args)
	var orders []Order
	inputFile, err := os.Open("orders.json")
	if err != nil{
		panic(err)
	}
	fmt.Println("opened file")
	defer inputFile.Close()
	byteArray, err := ioutil.ReadAll(inputFile)
	if err != nil{
		panic(err)
	}
	json.Unmarshal(byteArray, &orders)
	arrlen := len(orders)

	// waitgroups are for 
	var wg sync.WaitGroup
	over_ct,cold_ct,hot_ct,frozen_ct,dead := int32(15),int32(10),int32(10),int32(10),int32(0)
	overflow,cold,hot,frozen := make([]string,15),make([]string,10),make([]string,10),make([]string,10)
	for i := 0; i < 15; i++ {
		overflow[i] = ""
	}
	for i := 0; i < 10; i++ {
		cold[i] = "";
		hot[i] = "";
		frozen[i] = "";
	}
	for i:= 1; i < arrlen; i += 2 {
		blob_1,blob_2 := orders[i],orders[i-1]
		// TODO: add stuff for shelf contents
		shelf_1 := selectShelf(&blob_1, &over_ct,
				&cold_ct,&hot_ct,&frozen_ct,
				&dead)
		if (shelf_1 != &dead){
			decrement(shelf_1)
		}
		shelf_2 := selectShelf(&blob_2, &over_ct,
				&cold_ct,&hot_ct,&frozen_ct,
				&dead)
		if (shelf_2 != &dead){
			decrement(shelf_2)
		}
		arrival_1 := rand.Intn(6-2)+2
		arrival_2 := rand.Intn(6-2)+2
		// TODO: Add logging for dispatch
		if (shelf_1 != &dead){
			wg.Add(1)
			if (shelf_1 == &over_ct){
				// "go" keyword dispatches a goroutine 
				target_idx := placeInArray(overflow,blob_1.Id);
				fmt.Printf("Dispatched courier for order %s for overflow shelf\n", blob_1.Id)
				go courier(blob_1, shelf_1,arrival_1, 2, &wg,overflow,
					target_idx)
			} else if (shelf_1 == &cold_ct){
				target_idx := placeInArray(cold,blob_1.Id);
				fmt.Printf("Dispatched courier for order %s for cold shelf\n", blob_1.Id)
				go courier(blob_1, shelf_1,arrival_1,1,&wg,cold,target_idx);
			} else if (shelf_1 == &hot_ct){
				target_idx := placeInArray(hot,blob_1.Id);
				fmt.Printf("Dispatched courier for order %s for hot shelf\n", blob_1.Id)
				go courier(blob_1, shelf_1,arrival_1,1,&wg,hot,target_idx);
			} else {
				target_idx := placeInArray(frozen,blob_1.Id);
				fmt.Printf("Dispatched courier for order %s for frozen shelf\n", blob_1.Id)
				go courier(blob_1, shelf_1,arrival_1,1,&wg,frozen,target_idx);

			}
		}
		if (shelf_2 != &dead){
			wg.Add(1)
			if (shelf_2 == &over_ct){
				target_idx := placeInArray(overflow,blob_2.Id);
				fmt.Printf("Dispatched courier for order %s for overflow shelf\n", blob_2.Id)
				go courier(blob_2, shelf_2,
				arrival_2, 2,&wg,overflow,target_idx)
			} else if (shelf_2 == &cold_ct){
				target_idx := placeInArray(cold,blob_2.Id);
				fmt.Printf("Dispatched courier for order %s for cold shelf\n", blob_2.Id)
				go courier(blob_2, shelf_2,arrival_2,1,&wg,cold,target_idx);
			} else if (shelf_2 == &hot_ct){
				target_idx := placeInArray(hot,blob_2.Id);
				fmt.Printf("Dispatched courier for order %s for hot shelf\n", blob_2.Id)
				go courier(blob_2, shelf_2,arrival_2,1,&wg,hot,target_idx);
			} else {
				target_idx := placeInArray(frozen,blob_2.Id);
				fmt.Printf("Dispatched courier for order %s for frozen shelf\n", blob_2.Id)
				go courier(blob_2, shelf_2,arrival_2,1,&wg,frozen, target_idx);

			}
		}
		time.Sleep(2000*time.Millisecond)
	}
	wg.Wait()
	fmt.Println("complete")
}

func main(){
	/*
		TODO: clean up style stuff. I dont know what the rules
		are for formatting.
	*/


	shelf_size_prompt := "Specifies shelf capacity."
	overflowSize := flag.Uint("overflow_size", 15,shelf_size_prompt)
	hotSize := flag.Uint("hot_size", 10,shelf_size_prompt)
	coldSize := flag.Uint("cold_size", 10,shelf_size_prompt)
	frozenSize := flag.Uint("frozen_size", 10,shelf_size_prompt)

	courier_prompt := `
	Specify the timeframe bound for courier arrival.
	courier_lower_bound must be less than or equal to courier_upper_bound.
	courier_lower_bound and courier_upper_bound must be greater than or
	equal to 1.
	`

	courierLowerBound := flag.Uint("courier_lower_bound", 2, courier_prompt)
	courierUpperBound := flag.Uint("courier_upper_bound",6,courier_prompt)

	order_rate_prompt := `
	Specify the number of orders ingested per second.
	Must be greater than zero.
	`
	ordersPerSecond := flag.Uint("orders_per_second",2,order_rate_prompt)
	flag.Parse()
	if (*courierLowerBound > *courierUpperBound || *courierLowerBound < 1 ||
		*courierUpperBound < 1){
		fmt.Println(courier_prompt)
		os.Exit(1)

	}
	if (*ordersPerSecond < 1){
		fmt.Println(order_rate_prompt)
		os.Exit(1)

	}


	fmt.Println("overflow_size:",*overflowSize);
	fmt.Println("hot_size:",*hotSize);
	fmt.Println("cold_size:",*coldSize);
	fmt.Println("frozen_size:",*frozenSize);
	fmt.Println("courier_lower_bound", *courierLowerBound)
	fmt.Println("courier_upper_bound", *courierUpperBound)
	fmt.Println("orders_per_second", *ordersPerSecond)

	args := new(PrimaryArgs)
	args.overflow_size = *overflowSize
	args.hot_size = *hotSize
	args.cold_size = *coldSize
	args.frozen_size = *frozenSize
	args.courier_lower_bound = *courierLowerBound
	args.courier_upper_bound = *courierUpperBound
	args.orders_per_second = *ordersPerSecond

	runQueue(args)

}
