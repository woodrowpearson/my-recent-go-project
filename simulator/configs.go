package simulator

import (
	"errors"
	"io"
	"log"
	"math/rand"
	"os"
	"time"
)

type timeFunc func() time.Time

type randFunc func(lowerBound int, upperBound int) int

/*
 Helper struct for keeping argument lengths reasonable.
Access with BuildConfig.
*/
type Config struct {
	overflowSize      uint
	hotSize           uint
	coldSize          uint
	frozenSize        uint
	courierLowerBound uint
	courierUpperBound uint
	ordersPerSecond   uint
	overflowModifier  uint
	coldModifier      uint
	hotModifier       uint
	frozenModifier    uint
	receivedOutLog    *log.Logger
	swapLog           *log.Logger
	courierOutLog     *log.Logger
	courierErrLog     *log.Logger
	dispatchOutLog    *log.Logger
	dispatchErrLog    *log.Logger
	verboseLog        *log.Logger
	inputSource       io.Reader
	shelves           *orderShelves
	verbose           bool
	// necessary for mocks.
	getNow timeFunc
	// necessary for mocks.
	getRandRange randFunc
}

func getRandRange(lowerBound int, upperBound int) int {
	return rand.Intn(upperBound-lowerBound) + lowerBound
}

/*
Allows config to be built from code by other projects,
as opposed to just CLI args.
*/
func BuildConfig(overflowSize, hotSize,
	coldSize, frozenSize, courierLowerBound,
	courierUpperBound, ordersPerSecond,
	overflowModifier, coldModifier, hotModifier,
	frozenModifier uint, receivedOut, swapOut,
	courierOut, courierErr,
	dispatchOut, dispatchErr io.Writer,
	inputSource io.Reader, verbose bool) (*Config, error) {
	if courierLowerBound > courierUpperBound {

		return nil, errors.New(CourierPrompt)
	}
	overflow := buildOrderShelf(overflowSize, "overflow",
		overflowModifier)
	cold := buildOrderShelf(coldSize, "cold", coldModifier)
	hot := buildOrderShelf(hotSize, "hot", hotModifier)
	frozen := buildOrderShelf(frozenSize, "frozen", frozenModifier)
	dead := buildOrderShelf(1, "dead", 0)
	shelves := orderShelves{overflow: overflow, cold: cold, frozen: frozen,
		hot: hot, dead: dead}
	/*
		We need to wrap the logs in a log.Logger object.
		fmt.Fprintf, etc, are not thread-safe. Logger is.
	*/
	receivedOutLog := log.New(receivedOut, "", 0)
	swapOutLog := log.New(swapOut, "", 0)
	courierOutLog := log.New(courierOut, "", 0)
	courierErrLog := log.New(courierErr, "", 0)
	dispatchOutLog := log.New(dispatchOut, "", 0)
	dispatchErrLog := log.New(dispatchErr, "", 0)
	verboseLog := log.New(os.Stdout, "Ingested order:", log.Ltime)
	config := Config{
		overflowSize:      overflowSize,
		hotSize:           hotSize,
		coldSize:          coldSize,
		frozenSize:        frozenSize,
		courierLowerBound: courierLowerBound,
		courierUpperBound: courierUpperBound,
		ordersPerSecond:   ordersPerSecond,
		overflowModifier:  overflowModifier,
		coldModifier:      coldModifier,
		hotModifier:       hotModifier,
		frozenModifier:    frozenModifier,
		receivedOutLog:    receivedOutLog,
		swapLog:           swapOutLog,
		courierOutLog:     courierOutLog,
		courierErrLog:     courierErrLog,
		dispatchOutLog:    dispatchOutLog,
		dispatchErrLog:    dispatchErrLog,
		verboseLog:        verboseLog,
		inputSource:       inputSource,
		shelves:           &shelves,
		verbose:           verbose,
		getNow:            time.Now,
		getRandRange:      getRandRange,
	}
	return &config, nil
}
