package main

import (
	"bufio"
	"encoding/json"
	"io"
	"fmt"
	"time"
)

// TODO: write up documentation for this indicating
// that a 0 for orders_per_second must be passed in
// if the entirety of the orders is not inside a file
// (i.e. its from an arbitrary source
// TODO: have woody figure out a cleaner way to represent
// the orders per second from non-controllable sources
func streamFromSource(inputSource io.Reader, resultChannel chan Order, args *SimulatorConfig){
	/*
		a websocket can be represented by an io.Reader
		For the purposes of the default, it will be a file.
		For the unit tests, we'll use a bytes.Buffer

		We need to use a stream because parsing an entire file
		in memory could cause the box to run out of RAM
		(i.e. the JSON array in the file is 8gb).
		Additionally, by using a stream, a separate program could
		hook things in via a websocket.

		https://stackoverflow.com/questions/31794355/stream-large-json
	*/

	dec := json.NewDecoder(bufio.NewReader(inputSource))
	t, err := dec.Token()
	fmt.Println(t)
	check(err)
	ct := uint(0)
	for dec.More(){
		var o Order
		err := dec.Decode(&o)
		check(err)
		fmt.Printf("generated blob: %+v\n",o)
		resultChannel <- o
		ct += 1
		if ct == args.orders_per_second {
			ct = 0
			time.Sleep(time.Duration(
				args.second_value)*time.Second)
			fmt.Println("sleeping")
		}
	}
	t, err = dec.Token()
	fmt.Println(t)
	check(err)
	fmt.Println("closing input channel")
	close(resultChannel)
}
