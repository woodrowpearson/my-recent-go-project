package simulator

import (
	"io"
	"time"
	"errors"
	"math/rand"
)

type timeFunc func() time.Time

type randFunc func(lower_bound int, upper_bound int) int

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
	inputSource io.Reader
	shelves *Shelves
	verbose bool
	// necessary for mocks.
	getNow timeFunc
	// necessary for mocks.
	getRandRange randFunc
}

func getRandRange(lower_bound int, upper_bound int) int {
	return rand.Intn(upper_bound - lower_bound)+lower_bound
}

/*
Allows config to be built from code by other projects,
as opposed to just CLI args.
*/
func BuildConfig (overflow_size,hot_size,
		cold_size,frozen_size,courier_lower_bound,
		courier_upper_bound,orders_per_second,
		overflow_modifier,cold_modifier,hot_modifier,
		frozen_modifier uint,courier_out,courier_err,
		dispatch_out,dispatch_err io.Writer,
		inputSource io.Reader,verbose bool)(*SimulatorConfig, error){
	if (courier_lower_bound > courier_upper_bound){

		return nil,errors.New(CourierPrompt)
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
		inputSource:inputSource,
		shelves: &shelves,
		verbose: verbose,
		getNow: time.Now,
		getRandRange: getRandRange,
	}
	return &config,nil
}
