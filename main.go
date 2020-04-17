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

var dispatch_success_msg = `
Dispatched order %s to courier.
Current shelf: %s.
Current shelf contents: %s.
`
var dispatch_error_msg = "Order %s discarded due to lack of capacity\n"
var pickup_success_msg = `
Courier fetched item %d with remaining value of %.2f.
Current shelf: %s.
Current shelf contents: %s.
`
var pickup_error_msg = `
Discarded item with id %s due to expiration.
Current shelf: %s.
Current shelf contents: %s.
`

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
	counter int32
	item_array []string
	name string
	modifier uint
}

func buildShelf(array_capacity uint, name string,
		modifier uint) *Shelf {
	shelf := new(Shelf)
	shelf.item_array = make([]string, array_capacity)
	shelf.name = name;
	shelf.counter = int32(array_capacity)
	shelf.modifier = modifier
	for i := uint(0); i < array_capacity; i++ {
		shelf.item_array[i] = ""
	}
	return shelf
}

type PrimaryArgs struct {
	overflow_size uint
	hot_size uint
	cold_size uint
	frozen_size uint
	courier_lower_bound uint
	courier_upper_bound uint
	orders_per_second uint
	overflow_modifier uint
	cold_modifier uint
	hot_modifier uint
	frozen_modifier uint
}



func courier(order Order, shelf *Shelf, arrival_time int,
		wg *sync.WaitGroup,shelf_idx int){
	time.Sleep(time.Duration(1000*arrival_time)*time.Millisecond)
	a := float32(order.ShelfLife)
	b := order.DecayRate*float32(arrival_time)*float32(shelf.modifier)
	value := (a-b)/a
	atomic.AddInt32(&shelf.counter,1)
	shelf.item_array[shelf_idx] = ""
	wg.Done()
	if (value <= 0){
		fmt.Printf(pickup_error_msg,order.Id,shelf.name,shelf.item_array)
	} else {
		fmt.Printf(pickup_success_msg,order.Id,value,shelf.name,shelf.item_array)
	}
}


func selectShelf(order *Order, overflow_shelf *Shelf,
		cold_shelf *Shelf, hot_shelf *Shelf,
		frozen_shelf *Shelf,
		dead_shelf *Shelf) *Shelf {
	if (overflow_shelf.counter > 0){
		return overflow_shelf
	} else if (order.Temp == "cold" && cold_shelf.counter > 0){
		return cold_shelf
	} else if (order.Temp == "hot" && hot_shelf.counter > 0){
		return hot_shelf
	} else if (order.Temp == "frozen" &&
		frozen_shelf.counter > 0){
		return frozen_shelf
	}
	return dead_shelf
}


func decrementAndUpdate(shelf *Shelf, id string) int {
	atomic.AddInt32(&shelf.counter, -1);
	// TODO: make this smarter based on the counter value
	for i := 0; i < len(shelf.item_array); i++ {
		if (shelf.item_array[i] == ""){
			shelf.item_array[i] = id
			return i
		}
	}
	// Due to where this is called in the worflow,
	// This will never occur
	return -1
}

func runQueue(args *PrimaryArgs){

	fmt.Println(args)
	var orders []Order
	// TODO: move this to a streaming json parse
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
	arrlen := uint(len(orders))

	// waitgroups are for 
	var wg sync.WaitGroup
	overflow := buildShelf(args.overflow_size,"overflow",
			args.overflow_modifier)
	cold := buildShelf(args.cold_size, "cold",args.cold_modifier)
	hot := buildShelf(args.hot_size,"hot",args.hot_modifier)
	frozen := buildShelf(args.frozen_size,"frozen",args.frozen_modifier)
	dead := buildShelf(1,"dead",0)
	for i := uint(0); i < arrlen; i += args.orders_per_second {
		/*
			TODO: before dispatching, sort the items
			by criticality (i.e. longest arrival time)
		*/
		for j := uint(0); j < args.orders_per_second && i+j < arrlen; j++ {
			order := orders[i+j]
			shelf_idx := -1
			arrival := rand.Intn(
				int(args.courier_upper_bound -
				args.courier_lower_bound)) +
				int(args.courier_lower_bound)
			shelf := selectShelf(&order, overflow,
				cold,hot,frozen,dead)
			if (shelf != dead){
				wg.Add(1)
				shelf_idx = decrementAndUpdate(shelf,order.Id)
				fmt.Printf(dispatch_success_msg, order.Id,
					shelf.name, shelf.item_array) 
				go courier(order,shelf,arrival,&wg,shelf_idx)
			} else {
				fmt.Printf(dispatch_error_msg,order.Id)
			}

		}
		time.Sleep(1000*time.Millisecond)
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

	shelf_modifier_prompt := "Specifies shelf decay modifier"
	overflow_modifier := flag.Uint("overflow_modifier",2,
			shelf_modifier_prompt)
	cold_modifier := flag.Uint("cold_modifier",1,
			shelf_modifier_prompt)
	hot_modifier := flag.Uint("hot_modifier",1,
			shelf_modifier_prompt)
	frozen_modifier := flag.Uint("frozen_modifier",1,
			shelf_modifier_prompt)


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
	if (*courierLowerBound > *courierUpperBound ||
		 *courierLowerBound < 1 ||
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
	fmt.Println("overflow_modifier", *overflow_modifier)
	fmt.Println("cold_modifier", *cold_modifier)
	fmt.Println("hot_modifier", *hot_modifier)
	fmt.Println("frozen_modifier", *frozen_modifier)
	args := new(PrimaryArgs)
	args.overflow_size = *overflowSize
	args.hot_size = *hotSize
	args.cold_size = *coldSize
	args.frozen_size = *frozenSize
	args.courier_lower_bound = *courierLowerBound
	args.courier_upper_bound = *courierUpperBound
	args.orders_per_second = *ordersPerSecond
	args.overflow_modifier = *overflow_modifier
	args.cold_modifier = *cold_modifier
	args.hot_modifier = *hot_modifier
	args.frozen_modifier = *frozen_modifier
	runQueue(args)
}
