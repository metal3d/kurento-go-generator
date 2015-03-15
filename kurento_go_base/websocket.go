package kurento

import (
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

// Error that can be filled in response
type Error struct {
	Code    float64
	Message string
	Data    string
}

// Implements error built-in interface
func (e Error) Error() string {
	return fmt.Sprintf("[%d] %s %s", e.Code, e.Message, e.Data)
}

// Response represents server response
type Response struct {
	Jsonrpc string
	Id      float64
	Result  map[string]string // should change if result has no several form
	Error   Error
}

// the current websocket connection
// TODO: manage websocket connection lose
var ws *websocket.Conn

// client store Response chanels
var clients = make(map[float64]chan Response)
var clientId float64

// requesterFunc is a type to be injected in request methods (create, invoke, ...)
// The requestKMS is that type. The main goal is to allow tests to implement a
// fake request
type requesterFunc func(map[string]interface{}) <-chan Response

// requestKMS send a request to server and return the waiting chanel (read only)
// that contains Response.
// Note: type of requestKMS is requestFunc
var requestKMS requesterFunc = func(req map[string]interface{}) <-chan Response {
	//clientId = len(clients) + 1
	clientId++
	req["id"] = clientId

	clients[clientId] = make(chan Response)

	websocket.JSON.Send(ws, req)
	return clients[clientId]
}

// initiate connection to KMS
func connect(host string) {
	if ws != nil {
		// Connection already made
		return
	}
	var err error
	ws, err = websocket.Dial(host, "", "http://127.0.0.1")
	if err != nil {
		log.Fatal(err)
	}

	//manage response
	go func(ws *websocket.Conn) {
		for { // run forever
			r := Response{}
			websocket.JSON.Receive(ws, &r)
			// if webscocket client exists, send response to the chanel
			if clients[r.Id] != nil {
				clients[r.Id] <- r
				// chanel is read, we can delete it
				delete(clients, r.Id)
			}
		}
	}(ws)
}
