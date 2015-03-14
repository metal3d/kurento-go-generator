package kurento

import (
	LOG "log"
	"reflect"
	"strings"
)

// IMadiaElement implements some basic methods as getConstructorParams or Create()
type IMediaObject interface {

	// Return the constructor parameters
	getConstructorParams(IMediaObject, map[string]interface{}) map[string]interface{}

	// Each media object should be able to create another object
	// Those options are sent to getConstructorParams
	Create(IMediaObject, map[string]interface{})

	// Set ID of the element
	setId(string)
}

// Create object "m" with given "options"
func (elem *MediaObject) Create(m IMediaObject, options map[string]interface{}) {
	req := elem.getCreateRequest()
	constparams := m.getConstructorParams(elem, options)
	// TODO params["sessionId"]
	req["params"] = map[string]interface{}{
		"type":              getMediaElementType(m),
		"constructorParams": constparams,
	}
	if debug {
		log("request to be sent: ")
		log(req)
	}
	res := <-requestKMS(req)
	if res.Result["value"] != "" {
		m.setId(res.Result["value"])
	}
}

// setId set object id from a KMS response
func (m *MediaObject) setId(id string) {
	m.Id = id
}

// Build a prepared create request
func (m *MediaObject) getCreateRequest() map[string]interface{} {

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "create",
		"params":  make(map[string]interface{}),
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
	return m.Id
}

// NewMediaPipeline creation
func NewMediaPipeline() *MediaPipeline {
	m := &MediaObject{}
	ret := &MediaPipeline{}
	m.Create(ret, nil)
	return ret
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
