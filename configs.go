package main

import (
	"io"
	"time"
	"errors"
)

// Helper struct for keeping argument lengths reasonable.
type SimulatorConfig struct {
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
	courier_out io.Writer
	courier_err io.Writer
	dispatch_out io.Writer
	dispatch_err io.Writer
	// normally it's 1, but for tests we'll want it at 0.
	// refers to the value of a second
	second_value time.Duration
	shelves *Shelves
}

// Allows config to be built from code by other projects,
// as opposed to just CLI args.
// TODO: add in defaults.
func BuildConfig (overflow_size,hot_size,
		cold_size,frozen_size,courier_lower_bound,
		courier_upper_bound,orders_per_second,
		overflow_modifier,cold_modifier,hot_modifier,
		frozen_modifier uint,courier_out,courier_err,
		dispatch_out,dispatch_err io.Writer,
		second_value time.Duration)(*SimulatorConfig, error){
	if (courier_lower_bound > courier_upper_bound ||
		courier_lower_bound < 1 ||
		courier_upper_bound < 1){

		return nil,errors.New(CourierPrompt)
	}
	if (orders_per_second < 1){
		return nil, errors.New(OrderRatePrompt)
	}
	overflow := buildShelf(overflow_size,"overflow",
			overflow_modifier)
	cold := buildShelf(cold_size, "cold",cold_modifier)
	hot := buildShelf(hot_size,"hot",hot_modifier)
	frozen := buildShelf(frozen_size,"frozen",frozen_modifier)
	dead := buildShelf(1,"dead",0)
	shelves := Shelves{overflow:overflow,cold:cold,frozen:frozen,
			hot:hot,dead:dead}
	config := SimulatorConfig{
		overflow_size:overflow_size,
		hot_size: hot_size,
		cold_size: cold_size,
		frozen_size: frozen_size,
		courier_lower_bound: courier_lower_bound,
		courier_upper_bound: courier_upper_bound,
		orders_per_second: orders_per_second,
		overflow_modifier: overflow_modifier,
		cold_modifier: cold_modifier,
		hot_modifier: hot_modifier,
		frozen_modifier: frozen_modifier,
		courier_out:courier_out,
		courier_err:courier_err,
		dispatch_out:dispatch_out,
		dispatch_err:dispatch_err,
		second_value: 1000,
		shelves: &shelves,
	}
	return &config,nil
}
