package main

import (
	"flag"
	"fmt"
	"os"

	"./simulator"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	verbose := flag.Bool("v", false, simulator.VerbosePrompt)
	overflowSize := flag.Uint("overflow_size", 15, simulator.ShelfSizePrompt)
	hotSize := flag.Uint("hot_size", 10, simulator.ShelfSizePrompt)
	coldSize := flag.Uint("cold_size", 10, simulator.ShelfSizePrompt)
	frozenSize := flag.Uint("frozen_size", 10, simulator.ShelfSizePrompt)

	overflowModifier := flag.Uint("overflowModifier", 2,
		simulator.ShelfModifierPrompt)
	coldModifier := flag.Uint("cold_modifier", 1,
		simulator.ShelfModifierPrompt)
	hotModifier := flag.Uint("hot_modifier", 1,
		simulator.ShelfModifierPrompt)
	frozenModifier := flag.Uint("frozen_modifier", 1,
		simulator.ShelfModifierPrompt)

	courierLowerBound := flag.Uint("courier_lower_bound", 2, simulator.CourierPrompt)
	courierUpperBound := flag.Uint("courier_upper_bound", 6, simulator.CourierPrompt)
	ordersPerSecond := flag.Uint("orders_per_second", 2, simulator.OrderRatePrompt)
	flag.Parse()
	receivedOut, err := os.Create("received.log")
	check(err)
	defer receivedOut.Close()
	swapOut, err := os.Create("swapped.log")
	check(err)
	defer swapOut.Close()

	courierOut, err := os.Create("courier_out.log")
	check(err)
	defer courierOut.Close()
	courierErr, err := os.Create("courier_err.log")
	check(err)
	defer courierOut.Close()
	dispatchOut, err := os.Create("dispatch_out.log")
	check(err)
	defer dispatchOut.Close()
	dispatchErr, err := os.Create("dispatch_err.log")
	check(err)
	defer courierOut.Close()
	inputSource, err := os.Open("orders.json")
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
		*overflowModifier,
		*coldModifier,
		*hotModifier,
		*frozenModifier,
		receivedOut,
		swapOut,
		courierOut,
		courierErr,
		dispatchOut,
		dispatchErr,
		inputSource,
		*verbose,
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("Configuration: %+v\n", args)
	statistics := new(simulator.Statistics)
	statistics = simulator.Run(args, statistics)
	fmt.Printf("\nOverall Results:\n\n")
	fmt.Printf("Total Processed: %d\n", statistics.GetTotalProcessed())
	fmt.Printf("Total Successes: %d\n", statistics.GetTotalSuccesses())
	fmt.Printf("Total Failures: %d\n", statistics.GetTotalFailures())
	fmt.Printf("Total Swapped: %d\n", statistics.GetTotalSwapped())
	fmt.Printf("\nCold Items:\n\n")
	fmt.Printf("Total Cold Successes:%d\n", statistics.GetColdSuccesses())
	fmt.Printf("Total Cold Decayed:%d\n", statistics.GetColdDecayed())
	fmt.Printf("Total Cold Discarded:%d\n", statistics.GetColdDiscarded())
	fmt.Printf("\nHot Items:\n\n")
	fmt.Printf("Total Hot Successes:%d\n", statistics.GetHotSuccesses())
	fmt.Printf("Total Hot Decayed:%d\n", statistics.GetHotDecayed())
	fmt.Printf("Total Hot Discarded:%d\n", statistics.GetHotDiscarded())
	fmt.Printf("\nFrozen Items:\n\n")
	fmt.Printf("Total Frozen Successes:%d\n", statistics.GetFrozenSuccesses())
	fmt.Printf("Total Frozen Decayed:%d\n", statistics.GetFrozenDecayed())
	fmt.Printf("Total Frozen Discarded:%d\n", statistics.GetFrozenDiscarded())
}
