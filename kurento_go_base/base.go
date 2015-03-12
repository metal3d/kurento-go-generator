package kurento

import (
	LOG "log"
	"reflect"
	"strings"

	"golang.org/x/net/websocket"
)

// IMadiaElement implements some basic methods as getConstructorParams or Create()
type IMediaObject interface {

	// Return the constructor parameters
	getConstructorParams(IMediaObject, map[string]interface{}) map[string]interface{}

	// Each media object should be able to create another object
	// Those options are sent to getConstructorParams
	Create(IMediaObject, map[string]interface{})

	// Return the value of a given field
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

// Create object "m" with given "options"
func (elem *MediaObject) Create(m IMediaObject, options map[string]interface{}) {
	req := elem.getCreateRequest()
	req["type"] = getMediaElementType(m)
	params := m.getConstructorParams(elem, options)
	req["constructorParameters"] = params
	if debug {
		log(req)
	}
}

func (m *MediaObject) getField(name string) interface{} {
	val := reflect.ValueOf(m).FieldByName(name)
	if val.IsValid() {
		return val.String()
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

func mergeOptions(a, b map[string]interface{}) {
	for key, val := range b {
		a[key] = val
	}
}
