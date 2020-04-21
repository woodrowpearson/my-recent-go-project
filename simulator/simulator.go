package simulator


import (
	"time"
	"sync"
	"math/rand"
	"log"
)

func check(e error){
	if e != nil{
		panic(e)
	}
}

/*
Simulates an order being picked up by a courier.
Attempts to trigger shelf-swapping from overflow.
Logs whether the order is successfully picked up or decayed out.
*/
func courier(order *foodOrder, shelf *orderShelf,
		overflow *orderShelf,
		statistics *Statistics,
		wg *sync.WaitGroup,
		courier_out_log *log.Logger,
		courier_err_log *log.Logger,
		getNow timeFunc){
	time.Sleep(time.Until(order.arrivalTime))

	contents := shelf.duplicateContentsToSlice(order,false)
	if (order.IsCritical){
		// Decayed
		statistics.update(order,false,true)
		courier_err_log.Printf(PickupErrMsg,order.Id,order.DecayScore,shelf.name,contents)
	} else {
		// Success
		statistics.update(order,true,false)
		courier_out_log.Printf(PickupSuccessMsg,order.Id,order.DecayScore,shelf.name,contents)
	}
	shelf.swapAssessment(order,overflow,statistics,getNow)
	wg.Done()
}

/*
Determines when the courier will arrive, selects a shelf for placement,
places the item on the shelf, and dispatches a courier.
If there is no shelf space, logs a dispatching error message.
*/
func dispatch(o *foodOrder,  args *SimulatorConfig,
	statistics *Statistics,
	wg *sync.WaitGroup){
	/*
		Arrival seconds needs to be mocked, as
		rand.Intn will not accept a range of 0,
		which we need for the tests.
	*/
	arrival_seconds := args.getRandRange(int(args.courier_lower_bound),int(args.courier_upper_bound))
	shelf := o.selectShelf(args.shelves,arrival_seconds,args.getNow)
	if shelf != args.shelves.dead {
		wg.Add(1)
		contents := shelf.duplicateContentsToSlice(o,true)
		args.dispatch_out_log.Printf(DispatchSuccessMsg,o.Id,shelf.name,contents)
		go courier(o,shelf,args.shelves.overflow,
			statistics,wg,
			args.courier_out_log,
			args.courier_err_log,
			args.getNow)
	} else {
		// Discarded due to space.
		statistics.update(o,false,false)
		args.dispatch_err_log.Printf(DispatchErrMsg,o.Id)
	}
}
/*
Runs a simulator and returns statistics upon completion.
Statistics struct is passed in to allow for access to statistics
while the process is running.
*/
func Run(args *SimulatorConfig,statistics *Statistics) *Statistics {
	// Need to seed the rand global to get proper randomness.
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	resultChannel := make(chan foodOrder)
	go streamFromSource(args.inputSource,resultChannel,args)
	ioLoop:
	for {
		select {
		case v,ok := <-resultChannel:
			if ok {
				if args.verbose {
					args.verbose_log.Printf("%+v\n",&v)
				}
				dispatch(&v,args,statistics,&wg)
			} else {
				break ioLoop
			}
		}
	}
	wg.Wait()
	return statistics
}
