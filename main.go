package main


import (
	"fmt"
	"flag"
	"os"
	"./simulator"
)

func check(e error){
	if e != nil{
		panic(e)
	}
}

func main(){
	/*
		TODO: clean up style stuff. I dont know what the rules
		are for formatting and camelcase vs snakecase.
	*/

	overflowSize := flag.Uint("overflow_size", 15,simulator.ShelfSizePrompt)
	hotSize := flag.Uint("hot_size", 10,simulator.ShelfSizePrompt)
	coldSize := flag.Uint("cold_size", 10,simulator.ShelfSizePrompt)
	frozenSize := flag.Uint("frozen_size", 10,simulator.ShelfSizePrompt)

	overflow_modifier := flag.Uint("overflow_modifier",2,
			simulator.ShelfModifierPrompt)
	cold_modifier := flag.Uint("cold_modifier",1,
			simulator.ShelfModifierPrompt)
	hot_modifier := flag.Uint("hot_modifier",1,
			simulator.ShelfModifierPrompt)
	frozen_modifier := flag.Uint("frozen_modifier",1,
			simulator.ShelfModifierPrompt)

	courierLowerBound := flag.Uint("courier_lower_bound", 2, simulator.CourierPrompt)
	courierUpperBound := flag.Uint("courier_upper_bound",6,simulator.CourierPrompt)
	ordersPerSecond := flag.Uint("orders_per_second",2,simulator.OrderRatePrompt)
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
	args, err := simulator.BuildConfig(
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
	simulator.Run(args)
//	runQueue(args)
}
