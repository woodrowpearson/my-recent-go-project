package main

import (
	"flag"
	"fmt"
	"time"
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

func courier(order *Order, shelf *Shelf, arrival_time int,
		wg *sync.WaitGroup,shelf_idx int,
		courier_out io.Writer,courier_err io.Writer){
	time.Sleep(time.Duration(1000*arrival_time)*time.Millisecond)
	value := order.computeDecayScore(shelf,arrival_time)
	shelf.incrementAndUpdate(shelf_idx)
	/*
	In Linux, thread safety is assured in file access:
	https://stackoverflow.com/questions/29981050/concurrent-writing-to-a-file`
	*/
	if (order.DecayCritical){
		fmt.Fprintf(courier_err,PickupErrMsg,order.Id,value,shelf.name,shelf.item_array)
	} else {
		fmt.Fprintf(courier_out,PickupSuccessMsg,order.Id,value,shelf.name,shelf.item_array)
	}
	wg.Done()
}


func selectShelf(o *Order,s *Shelves,arrival_time int) *Shelf {
	/*
	TODO: add in a criticality score for the order.
	If the order is not safe for overflow, don't stick it 
	in overflow unless matching shelf is empty.

	*/
	matchingShelf := s.overflow
	switch o.Temp{
		case "cold":
			matchingShelf = s.cold
		case "hot":
			matchingShelf = s.hot
		case "frozen":
			matchingShelf = s.frozen
	}
	if (s.overflow.counter < 1 && matchingShelf.counter < 1){
		// nowhere to place, must discard.
		return s.dead
	}
	overflowDecayScore := o.computeDecayScore(s.overflow,
					arrival_time)
	matchingDecayScore := o.computeDecayScore(matchingShelf,
				arrival_time)

	if (overflowDecayScore > 0 && s.overflow.counter > 0){
		// will survive overflow
		return s.overflow
	}
	if (matchingDecayScore <= 0){
		// will die no matter where we put it.
		return s.dead
	}
	// TODO:
	// add in check for swapping out critical orders
	// if critical order found, compute 
	// chance of saving. if saveable,  
	if (matchingShelf.counter > 0){
		/*
		place in matching shelf - we're safe.
		*/
		o.DecayCritical = false
		o.DecayScore = matchingDecayScore
		return matchingShelf
	}
	// last resort:
	// 	place in overflow and hope we can swap
	// 	later.
	o.DecayScore = overflowDecayScore
	o.DecayCritical = true
	return s.overflow
}

func dispatch(o *Order,  args *SimulatorConfig,
	wg *sync.WaitGroup){
	/* 
	TODO: get a concurrent structure in here
	(a map maybe?) that will let us swap out
	overflow + critical orders to a matching shelf.
	*/
	arrival := rand.Intn(
		int(args.courier_upper_bound -
		args.courier_lower_bound)) +
		int(args.courier_lower_bound)
	shelf := selectShelf(o,args.shelves,arrival)
	if shelf != args.shelves.dead {
		wg.Add(1)
		shelf_idx, err := shelf.decrementAndUpdate(o.Id)
		check(err)
		fmt.Fprintf(args.dispatch_out,DispatchSuccessMsg,
			o.Id,shelf.name,shelf.item_array)
		go courier(o,shelf,arrival,wg,shelf_idx,
			args.courier_out,args.courier_err)
		check(err)
	} else {
		fmt.Fprintf(args.dispatch_err,
			DispatchErrMsg,o.Id)
	}
}

func runQueue(args *SimulatorConfig){
	fmt.Println(args)
	var wg sync.WaitGroup
	resultChannel := make(chan Order)
	go streamFromSource(args.inputSource,resultChannel,args)
	ioLoop:
	for {
		select {
		case v,ok := <-resultChannel:
			if ok {
				dispatch(&v,args,&wg)
			} else {
				break ioLoop
			}
		}
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
	inputSource,err := os.Open("orders.json")
	check(err)
	defer inputSource.Close()
	args, err := BuildConfig(
		*overflowSize,
		*hotSize,
		*coldSize,
		*frozenSize,
		*courierLowerBound,
		*courierUpperBound,
		*ordersPerSecond,
		*overflow_modifier,
		*cold_modifier,
		*hot_modifier,
		*frozen_modifier,
		courier_out,
		courier_err,
		dispatch_out,
		dispatch_err,
		inputSource,
		1,
	)
	if err != nil {
		fmt.Println(err.Error());
		os.Exit(1)
	}
	fmt.Printf("Configuration: %+v\n", args)
	runQueue(args)
}
