package kurento

import (
	LOG "log"
	"reflect"
	"strings"

	"golang.org/x/net/websocket"
)

// IMadiaElement implements some basic methods as getContructorParams or Create()
type IMediaObject interface {
	getConstructorParams(IMediaObject) map[string]interface{}
	Create(IMediaObject)
	getField(name string) interface{}
}

var kurentoclients = make(map[string]*ServerClient)
var currClientID int64

type ServerClient struct {
	MediaObject
	ws *websocket.Conn
}

func GetClient(uri string) (*ServerClient, error) {
	var err error
	if kurentoclients[uri] == nil {
		s := ServerClient{}
		s.ws, err = websocket.Dial(uri, "", "http://127.0.0.1:8080")
		kurentoclients[uri] = &s
	}

	return kurentoclients[uri], err
}

func (m *MediaObject) getField(name string) interface{} {
	val := reflect.ValueOf(name).FieldByName(name)
	if val.IsValid() {
		return val
	}
	return nil
}

// Build a prepared create request
func (m *MediaObject) getCreateRequest() map[string]interface{} {
	currClientID++

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      currClientID,
		"method":  "create",
	}
}

// Build a prepared invoke request
func (m *MediaObject) getInvokeRequest() map[string]interface{} {
	req := m.getCreateRequest()
	req["method"] = "invoke"

	return req
}

// String implement Stringer interface, return ID
func (m *MediaObject) String() string {
	return m.id
}

// Return name of the object
func getMediaElementType(i interface{}) string {
	n := reflect.TypeOf(i).String()
	p := strings.Split(n, ".")
	return p[len(p)-1]
}

var debug = false

func log(i interface{}) {
	if debug {
		LOG.Println(i)
	}
}
