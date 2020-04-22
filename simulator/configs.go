package simulator

import (
	"io"
	"time"
	"errors"
	"math/rand"
	"log"
	"os"
)

type timeFunc func() time.Time

type randFunc func(lower_bound int, upper_bound int) int

/*
 Helper struct for keeping argument lengths reasonable.
Access with BuildConfig.
*/
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
	receivedOutLog *log.Logger
	swapLog *log.Logger
	courier_out_log *log.Logger
	courier_err_log *log.Logger
	dispatch_out_log *log.Logger
	dispatch_err_log *log.Logger
	verbose_log *log.Logger
	inputSource io.Reader
	shelves *orderShelves
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
		frozen_modifier uint,receivedOut,swapOut,
		courier_out,courier_err,
		dispatch_out,dispatch_err io.Writer,
		inputSource io.Reader,verbose bool)(*SimulatorConfig, error){
	if (courier_lower_bound > courier_upper_bound){

		return nil,errors.New(CourierPrompt)
	}
	overflow := buildOrderShelf(overflow_size,"overflow",
			overflow_modifier)
	cold := buildOrderShelf(cold_size, "cold",cold_modifier)
	hot := buildOrderShelf(hot_size,"hot",hot_modifier)
	frozen := buildOrderShelf(frozen_size,"frozen",frozen_modifier)
	dead := buildOrderShelf(1,"dead",0)
	shelves := orderShelves{overflow:overflow,cold:cold,frozen:frozen,
			hot:hot,dead:dead}
	/*
	We need to wrap the logs in a log.Logger object.
	fmt.Fprintf, etc, are not thread-safe. Logger is.
	*/
	receivedOutLog := log.New(receivedOut,"",0)
	swapOutLog := log.New(swapOut,"",0)
	courier_out_log := log.New(courier_out,"",0)
	courier_err_log := log.New(courier_err,"",0)
	dispatch_out_log := log.New(dispatch_out,"",0)
	dispatch_err_log := log.New(dispatch_err,"",0)
	verbose_log := log.New(os.Stdout,"Ingested order:",log.Ltime)
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
		receivedOutLog:receivedOutLog,
		swapLog:swapOutLog,
		courier_out_log:courier_out_log,
		courier_err_log:courier_err_log,
		dispatch_out_log:dispatch_out_log,
		dispatch_err_log:dispatch_err_log,
		verbose_log: verbose_log,
		inputSource:inputSource,
		shelves: &shelves,
		verbose: verbose,
		getNow: time.Now,
		getRandRange: getRandRange,
	}
	return &config,nil
}
