
# Build

- go build main.go
- ./main

# via go run

- go run main.go

# test

### Standard:
- cd simulator && go test

### Race Conditions
- cd simulator && go test -race

It is best to run the race conditions test a few times for any given test run. 

## installation of shit

- (go 1.14)
- sudo snap install go --classic (ubuntu) 
- brew install go (os X)
- go get "github.com/orcaman/concurrent-map"
- go build main.go

## API Usage

Two structs are available:
	- SimulatorConfig
	- Statistics

Two functions are available:
	- BuildConfig
	- Run

Run() activates the process. It requires a pointer to a Statistics struct (which can be instantiated with a literal)
and a SimulatorConfig. Instantiate a SimulatorConfig with BuildConfig.

The Statistics struct may be inspected at any point during the running of the process to get current counts.
Use the public methods on the Statistics to access the values.

## Narrative:

1. foodOrder structs are pushed to a channel(foodOrder) from a generic io.Reader in streamFromSource. The streamFromSource function reads and parses the JSON structs in a streaming fashion. Streaming the ingestion allows us to limit memory usage (i.e. if we have a file of 100M orders, all orders are not in memory at a given time) and also allows us to ingest the JSON from sources beyond a file, for example from a websocket.
2. the ioLoop in Run() reads from the channel pushed to by streamFromSource, calling dispatch() with the received orders, until the channel is closed. Run() finishes when all dispatched goroutines are completed.
3. dispatch() determines the courier arrival time with a random range call, selects a shelf based on availability and suitability (i.e. will the shelf expire if it sits in overflow? If there is no space but in overflow for an order, and the order will expire in overflow, the order is added to both the shelf's contents map as well as a separate map on the shelf struct containing orders that are at-risk for expiration (the "criticals" map)
4. If there is no shelf space anywhere, or the order will expire no matter where it is placed, the dispatcher logs out a discard message. Otherwise, the dispatcher calls a goroutine on the courier function, simulating a courier pickup.
4. The courier function sleeps until the courier's arrival date. If the order has decayed out, it is logged as a decayed message. If the order has not decayed out, it is logged as a success. In either scenario, the order is removed from the shelf. 
5. as a final action in the courier function, the courier determines if the order was on the overflow shelf. if it was not, the courier looks for at-risk orders on the overflow shelf that can be saved from decay by a shelf swap, and moves them to the newly updated non-overflow shelf with an updated decay score.


## Weaknesses

1. This implementation does not account for the scenario where the overflow shelf has a lower decay multiplier than the non-overflow shelves.
2. This implementation only accounts for having four shelves in total. It does not account for having a wide range of temperatures.
