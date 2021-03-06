package simulator

import (
	"bufio"
	"encoding/json"
	"io"
	"time"
)

/*
Ingest and parse inputs from an io.Reader in a streaming manner.
Adds an artificial pause between blocks of orders if configured
with an order rate of > 0. An order rate of 0 introduces no pause.
*/
func streamFromSource(inputSource io.Reader, resultChannel chan foodOrder, args *Config) {
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
	_, err := dec.Token()
	check(err)
	ct := uint(0)
	for dec.More() {
		var o foodOrder
		err := dec.Decode(&o)
		check(err)
		resultChannel <- o
		ct += 1
		/*
			Adds an artificial pause for simulating input rates.
			If there is no pause, input rate is controlled by the inputSource's supplier.
		*/
		if args.ordersPerSecond > 0 && ct == args.ordersPerSecond {
			ct = 0
			time.Sleep(1 * time.Second)
		}
	}
	_, err = dec.Token()
	check(err)
	close(resultChannel)
}
