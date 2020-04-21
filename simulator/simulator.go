package simulator


import (
	"fmt"
	"time"
	"io"
	"sync"
	"sync/atomic"
	"math/rand"
)

type Statistics struct {
	hotDiscarded uint64
	coldDiscarded uint64
	frozenDiscarded uint64
	hotDecayed uint64
	coldDecayed uint64
	frozenDecayed uint64
	hotSuccess uint64
	coldSuccess uint64
	frozenSuccess uint64
	totalDiscarded uint64
	totalDecayed uint64
	totalFailures uint64
	totalSuccesses uint64
	totalProcessed uint64
	totalSwapped uint64
}

func(stat *Statistics) update(o *Order, success bool, decayed bool){
	if success {
		atomic.AddUint64(&stat.totalSuccesses,1)
		switch o.Temp {
		case "hot":
			atomic.AddUint64(&stat.hotSuccess,1)
		case "cold":
			atomic.AddUint64(&stat.coldSuccess,1)
		case "frozen":
			atomic.AddUint64(&stat.frozenSuccess,1)
		}
	} else {
		atomic.AddUint64(&stat.totalFailures,1)
		switch o.Temp {
		case "hot":
			if decayed {
				atomic.AddUint64(&stat.hotDecayed,1)
				atomic.AddUint64(&stat.totalDecayed,1)
			} else {
				atomic.AddUint64(&stat.hotDiscarded,1)
				atomic.AddUint64(&stat.totalDiscarded,1)
			}
		case "cold":
			if decayed {
				atomic.AddUint64(&stat.coldDecayed,1)
				atomic.AddUint64(&stat.totalDecayed,1)
			} else {
				atomic.AddUint64(&stat.coldDiscarded,1)
				atomic.AddUint64(&stat.totalDiscarded,1)
			}
		case "frozen":
			if decayed {
				atomic.AddUint64(&stat.frozenDecayed,1)
				atomic.AddUint64(&stat.totalDecayed,1)
			} else {
				atomic.AddUint64(&stat.frozenDiscarded,1)
				atomic.AddUint64(&stat.totalDiscarded,1)
			}
		}
	}
	atomic.AddUint64(&stat.totalProcessed,1)
}

func(s *Statistics) updateSwapped(){
	atomic.AddUint64(&s.totalSwapped,1)
}

func(s *Statistics) GetHotDiscarded() uint64 {return s.hotDiscarded}
func(s *Statistics) GetColdDiscarded() uint64 {return s.coldDiscarded}
func(s *Statistics) GetFrozenDiscarded() uint64 {return s.frozenDiscarded}
func(s *Statistics) GetHotDecayed() uint64 {return s.hotDecayed}
func(s *Statistics) GetColdDecayed() uint64 {return s.coldDecayed}
func(s *Statistics) GetFrozenDecayed() uint64 {return s.frozenDecayed}
func(s *Statistics) GetHotSuccess() uint64 {return s.hotSuccess}
func(s *Statistics) GetColdSuccess() uint64 {return s.coldSuccess}
func(s *Statistics) GetFrozenSuccess() uint64 {return s.frozenSuccess}
func(s *Statistics) GetTotalDiscarded() uint64 {return s.totalDiscarded}
func(s *Statistics) GetTotalDecayed() uint64 {return s.totalDecayed}
func(s *Statistics) GetTotalFailures() uint64 {return s.totalFailures}
func(s *Statistics) GetTotalSuccesses() uint64 {return s.totalSuccesses}
func(s *Statistics) GetTotalProcessed() uint64 {return s.totalProcessed}
func(s *Statistics) GetTotalSwapped() uint64 {return s.totalSwapped}


func check(e error){
	if e != nil{
		panic(e)
	}
}



func courier(order *Order, shelf *Shelf,
		overflow *Shelf,
		statistics *Statistics,
		wg *sync.WaitGroup,
		courier_out io.Writer,
		courier_err io.Writer,getNow timeFunc){
	time.Sleep(time.Until(order.arrivalTime))

	contents := shelf.duplicateContentsToSlice(order,false)
	/*
	In Linux, thread safety is assured in file access:
	https://stackoverflow.com/questions/29981050/concurrent-writing-to-a-file`

	TODO: move these Fprintfs to using logger.
	Go's logger is concurrent
	https://stackoverflow.com/questions/18361750/correct-approach-to-global-logging-in-golang/18362952#18362952
	*/
	if (order.IsCritical){
		// Decayed
		statistics.update(order,false,true)
		fmt.Fprintf(courier_err,PickupErrMsg,
			order.Id,order.DecayScore,
			shelf.name,contents)
	} else {
		// Success
		statistics.update(order,true,false)
		fmt.Fprintf(courier_out,PickupSuccessMsg,order.Id,
			order.DecayScore,shelf.name,contents)
	}
	/*
	 Determine if we can move an at-risk order from the overflow shelf
	in order to prevent it from decaying before pickup
	*/
	shelf.swapAssessment(order,overflow,statistics,getNow)
	wg.Done()
}



func dispatch(o *Order,  args *SimulatorConfig,
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
		// TODO: Move this to a Logger. Logger is concurrent.
		// Fprintf is causing a race condition.
		fmt.Fprintf(args.dispatch_out,DispatchSuccessMsg,
			o.Id,shelf.name,contents)
		go courier(o,shelf,args.shelves.overflow,
			statistics,wg,
			args.courier_out,args.courier_err,args.getNow)
	} else {
		// Discarded due to space.
		statistics.update(o,false,false)
		fmt.Fprintf(args.dispatch_err,
			DispatchErrMsg,o.Id)
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
	resultChannel := make(chan Order)
	go streamFromSource(args.inputSource,resultChannel,args)
	ioLoop:
	for {
		select {
		case v,ok := <-resultChannel:
			if ok {
//				if args.verbose {
//					fmt.Printf("Received order: %+v\n", &v)
//				}
				dispatch(&v,args,statistics,&wg)
			} else {
				break ioLoop
			}
		}
	}
	wg.Wait()
	return statistics
}
