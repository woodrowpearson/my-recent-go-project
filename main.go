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

const dispatch_success_msg = `
Dispatched order %s to courier.
Current shelf: %s.
Current shelf contents: %s.
`
const dispatch_error_msg = "Order %s discarded due to lack of capacity\n"
const pickup_success_msg = `
Courier fetched item %d with remaining value of %.2f.
Current shelf: %s.
Current shelf contents: %s.
`
const pickup_error_msg = `
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


type Shelf struct {
	counter int32
	item_array []string
	name string
	modifier uint
}

func buildShelf(array_capacity uint, name string,
		modifier uint) *Shelf {
	// TODO: UNIT TEST THIS
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
// TODO: MAKE THIS THROW AN ERROR ON PATHOLOGICAL RESPONSE OF NONE FOUND
func (s *Shelf) decrementAndUpdate(id string) int {
	atomic.AddInt32(&s.counter, -1);
	// TODO: make this smarter based on the counter value
	// TODO: UNIT TEST THIS
	for i := 0; i < len(s.item_array); i++ {
		if (s.item_array[i] == ""){
			s.item_array[i] = id
			return i
		}
	}
	// Due to where this is called in the worflow,
	// This will never occur
	return -1
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
	// TODO: MAKE THIS READ FROM A CONSTANT, SO THAT IT CAN BE MOCKED TO ZERO
	time.Sleep(time.Duration(1000*arrival_time)*time.Millisecond)
	// END BLOCK
	// TODO: MOVE THIS BLOCK TO A FUNCTION THAT CAN BE MOCKED
	a := float32(order.ShelfLife)
	b := order.DecayRate*float32(arrival_time)*float32(shelf.modifier)
	value := (a-b)/a
	// END BLOCK
	atomic.AddInt32(&shelf.counter,1)
	shelf.item_array[shelf_idx] = ""
	wg.Done()
	/*
	 TODO: PASS IN AN io.Writer HANDLE TO THIS INSTEAD OF WRITING TO STDOUT.
		ENSURE THREAD SAFETY	
	*/
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
	// END BLOCK
	// TODO: add explanation and link for waitgroup behavior
	var wg sync.WaitGroup
	// TODO: MOVE THESE FIVE TO THE INITIAL ARGUMENT PARSING STRUCT
	overflow := buildShelf(args.overflow_size,"overflow",
			args.overflow_modifier)
	cold := buildShelf(args.cold_size, "cold",args.cold_modifier)
	hot := buildShelf(args.hot_size,"hot",args.hot_modifier)
	frozen := buildShelf(args.frozen_size,"frozen",args.frozen_modifier)
	dead := buildShelf(1,"dead",0)
	// END BLOCK
	for i := uint(0); i < arrlen; i += args.orders_per_second {
		/*
			TODO: before dispatching, sort the items
			by criticality (i.e. longest arrival time)
			We'll want to compute the score for the order 
			at instantiation. 
			TODO: find an equivalent of python's bisect
			function for inserting into the array in a sorted manner
			TODO: MAKE THE CRITICALITY SORT A SEPARATE FUNCTION AND TEST IT
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
			// TODO: MOVE THIS TO OUTSIDE OF THE J LOOP.
			// TODO: PASS IN AN IO.Writer INSTEAD OF 
			// PRINTING TO STDOUT(FOR TESTING PURPOSES)
			if (shelf != dead){
				wg.Add(1)
				shelf_idx = shelf.decrementAndUpdate(order.Id)
				fmt.Printf(dispatch_success_msg, order.Id,
					shelf.name, shelf.item_array)
				go courier(order,shelf,arrival,&wg,shelf_idx)
			} else {
				fmt.Printf(dispatch_error_msg,order.Id)
			}
		}
		// TODO: HAVE THIS READ FROM AN ARGUMENT PASSED IN, SO THAT IT CAN BE MOCKED
		time.Sleep(1000*time.Millisecond)
	}
	wg.Wait()
	fmt.Println("complete")
}

func main(){
	/*
		TODO: clean up style stuff. I dont know what the rules
		are for formatting and camelcase vs snakecase.
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
	args := PrimaryArgs{
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
	}
	fmt.Printf("Configuration: %+v\n", args)
	runQueue(&args)
}
