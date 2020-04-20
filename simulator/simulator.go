package simulator


import (
	"fmt"
	"time"
	"io"
	"sync"
	"math/rand"
)

func check(e error){
	if e != nil{
		panic(e)
	}
}



func courier(order *Order, shelf *Shelf,overflow *Shelf,
		wg *sync.WaitGroup,
		courier_out io.Writer,
		courier_err io.Writer,getNow timeFunc){
	time.Sleep(time.Until(order.arrivalTime))

	contents := shelf.duplicateContents(order,false)
	/*
	In Linux, thread safety is assured in file access:
	https://stackoverflow.com/questions/29981050/concurrent-writing-to-a-file`
	*/
	if (order.IsCritical){
		fmt.Fprintf(courier_err,PickupErrMsg,
			order.Id,order.DecayScore,
			shelf.name,contents)
	} else {
		fmt.Fprintf(courier_out,PickupSuccessMsg,order.Id,
			order.DecayScore,shelf.name,contents)
	}
	/*
	 Determine if we can move an at-risk order from the overflow shelf
	in order to prevent it from decaying before pickup
	*/
	shelf.swapAssessment(order,overflow,getNow)
	wg.Done()
}



func dispatch(o *Order,  args *SimulatorConfig,
	wg *sync.WaitGroup){
	/* 
	TODO: get a concurrent structure in here
	(a map maybe?) that will let us swap out
	overflow + critical orders to a matching shelf.
	*/

	// TODO: move this to a method on the args struct so it can be mocked
	arrival_seconds := rand.Intn(
		int(args.courier_upper_bound -
		args.courier_lower_bound)) +
		int(args.courier_lower_bound)
	// END BLOCK
	shelf := o.selectShelf(args.shelves,arrival_seconds,args.getNow)
	if shelf != args.shelves.dead {
		wg.Add(1)
		contents := shelf.duplicateContents(o,true)
		fmt.Fprintf(args.dispatch_out,DispatchSuccessMsg,
			o.Id,shelf.name,contents)
		go courier(o,shelf,args.shelves.overflow,wg,
			args.courier_out,args.courier_err,args.getNow)
	} else {
		fmt.Fprintf(args.dispatch_err,
			DispatchErrMsg,o.Id)
	}
}

func Run(args *SimulatorConfig){
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
