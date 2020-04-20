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

func getUnixTimestampMilliseconds() int64{
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func courier(order *Order, shelf *Shelf,overflow *Shelf,
//		arrival_time int,
		wg *sync.WaitGroup,
		courier_out io.Writer,courier_err io.Writer){
//	time.Sleep(time.Duration(1000*arrival_time)*time.Millisecond)
	time.After(time.Until(order.arrivalTime))
	fmt.Printf("completed: ")
	fmt.Println(time.Now().String())
//	value := order.computeDecayScore(shelf,arrival_time)

	/*
	In Linux, thread safety is assured in file access:
	https://stackoverflow.com/questions/29981050/concurrent-writing-to-a-file`
	*/
	if (order.IsCritical){
		fmt.Fprintf(courier_err,PickupErrMsg,order.Id,order.DecayScore,
			shelf.name,shelf.contents)
	} else {
		fmt.Fprintf(courier_out,PickupSuccessMsg,order.Id,
			order.DecayScore,shelf.name,shelf.contents)
	}
	/*
		 In the event that we're freeing up space on
		a non-overflow shelf, we'll want to scan the overflow shelf's
		criticals for the first item that will match the following criteria:
		1) eligible for this shelf due to temperature match
		2) will be saved from decay by moving to the current shelf
		Once the item is found, we swap the item from the matching shelf,
		remove it from criticals, assign it a new decay factor,
		and run incrementAndUpdate on the overflow shelf.
	*/
	if shelf != overflow && shelf.counter == 0{
		to_swap := overflow.selectCritical(shelf)
		if to_swap != nil{
			overflow.incrementAndUpdate(to_swap)
			shelf.contents.Remove(order.Id)
			shelf.contents.Set(to_swap.Id,to_swap)
		} else {
			shelf.incrementAndUpdate(order)
		}
	} else {
		shelf.incrementAndUpdate(order)
	}
	wg.Done()
}



func dispatch(o *Order,  args *SimulatorConfig,
	wg *sync.WaitGroup){
	/* 
	TODO: get a concurrent structure in here
	(a map maybe?) that will let us swap out
	overflow + critical orders to a matching shelf.
	*/
	arrival_seconds := rand.Intn(
		int(args.courier_upper_bound -
		args.courier_lower_bound)) +
		int(args.courier_lower_bound)
	shelf := o.selectShelf(args.shelves,arrival_seconds)
	if shelf != args.shelves.dead {
		wg.Add(1)
		fmt.Fprintf(args.dispatch_out,DispatchSuccessMsg,
			o.Id,shelf.name,shelf.contents)
		go courier(o,shelf,args.shelves.overflow,wg,
			args.courier_out,args.courier_err)
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
