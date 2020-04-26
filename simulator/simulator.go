package simulator

import (
	"math/rand"
	"sync"
	"time"
	//	"log"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
Simulates an order being picked up by a courier.
Attempts to trigger shelf-swapping from overflow.
Logs whether the order is successfully picked up or decayed out.
*/
func courier(order *foodOrder, shelf *orderShelf,
	statistics *Statistics,
	wg *sync.WaitGroup,
	args *Config) {
	time.Sleep(time.Until(order.arrivalTime))

	contents := shelf.duplicateContentsToSlice(order, false)
	if order.IsCritical {
		// Decayed
		statistics.update(order, false, true)
		args.courierErrLog.Printf(PickupErrMsg, order.Id, order.DecayScore, shelf.name, contents)
	} else {
		// Success
		statistics.update(order, true, false)
		args.courierOutLog.Printf(PickupSuccessMsg, order.Id, order.DecayScore, shelf.name, contents)
	}
	shelf.swapAssessment(order, statistics, args)
	wg.Done()
}

/*
Determines when the courier will arrive, selects a shelf for placement,
places the item on the shelf, and dispatches a courier.
If there is no shelf space, logs a dispatching error message.
*/
func dispatch(o *foodOrder, args *Config,
	statistics *Statistics,
	wg *sync.WaitGroup) {
	/*
		Arrival seconds needs to be mocked, as
		rand.Intn will not accept a range of 0,
		which we need for the tests.
	*/
	arrivalSeconds := args.getRandRange(int(args.courierLowerBound), int(args.courierUpperBound))
	shelf := o.selectShelf(args.shelves, arrivalSeconds, args.getNow)
	if shelf != args.shelves.dead {
		wg.Add(1)
		contents := shelf.duplicateContentsToSlice(o, true)
		args.dispatchOutLog.Printf(DispatchSuccessMsg, o.Id, shelf.name, contents)
		go courier(o, shelf,
			statistics, wg,
			args)
	} else {
		// Discarded due to space.
		statistics.update(o, false, false)
		args.dispatchErrLog.Printf(DispatchErrMsg, o.Id)
	}
}

/*
Runs a simulator and returns statistics upon completion.
Statistics struct is passed in to allow for access to statistics
while the process is running.
*/
func Run(args *Config, statistics *Statistics) *Statistics {
	// Need to seed the rand global to get proper randomness.
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	resultChannel := make(chan foodOrder)
	go streamFromSource(args.inputSource, resultChannel, args)
ioLoop:
	for {
		select {
		case v, ok := <-resultChannel:
			if ok {
				if args.verbose {
					args.verboseLog.Printf("%+v\n", &v)
				}
				args.receivedOutLog.Printf(OrderReceivedMsg, v.Id, v.Name, v.Temp, v.ShelfLife, v.DecayRate)
				dispatch(&v, args, statistics, &wg)
			} else {
				break ioLoop
			}
		}
	}
	wg.Wait()
	return statistics
}
