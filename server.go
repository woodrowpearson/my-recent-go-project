package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/trace"
	"golang.org/x/net/websocket"
)

var (
	port = flag.Int("port", 8000, "The server port")
)


Redundant said order existed
//type Order struct {
	// The fields of this struct must be exported so that the json module will be
	// able to write into them. Therefore we need field tags to specify the names
	// by which these fields go in the JSON representation of events.
//	id str `json:"id"`
///	name str `json:"name"`
//	temp str `json:"temp"`
//	shelfLife uint32 `json:"shelfLife"`
//	decayRate float32 `json:"decayRate"`
//}

// handleWebsocketOrderRequest handles the message arriving on connection ws
// from the CLI or is will it be the client? or the CLI is invoking from CLI?
func handleWebsocketOrderRequest(ws *websocket.Conn, e Event) error {
	// Send the event as JSON
	err := websocket.JSON.Send(ws, e)
	if err != nil {
		return fmt.Errorf("Can't send: %s", err.Error())
	}
	return nil
}

// websocketServerConnection handles the ws connection as long as orders are coming
func websocketServerConnection(ws *websocket.Conn) {
	log.Printf("Client connected from %s", ws.RemoteAddr())
	for {
		var order Order
		err := websocket.JSON.Receive(ws, &order)
		if err != nil {
			log.Printf("Receive failed: %s; closing connection...", err.Error())
			if err = ws.Close(); err != nil {
				log.Println("Error closing connection:", err.Error())
			}
			break
		} else {
			if err := handleWebsocketOrderRequest(ws, order); err != nil {
				log.Println(err.Error())
				break
			}
		}
	}
}

// websocketClientConnection handles the single websocket time connection - ws.
func websocketClientConnection(ws *websocket.Conn) {
	for range time.Tick(2 * time.Second) {
		// Once every 2 seconds, send two json objects (as a string).
		websocket.Message.Send(ws, os.Open("orders.json"))
	}
}

func main() {
	flag.Parse()
	// Set up websocket servers and static file server. In addition, we're using
	// net/trace for debugging - it will be available at /debug/requests.
	http.Handle("/wsserver", websocket.Handler(websocketServerConnection))
	http.Handle("/wsclient", websocket.Handler(websocketClientConnection))

	log.Printf("Server listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
