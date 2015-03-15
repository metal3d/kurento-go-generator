package kurento

import (
	"encoding/json"
	"testing"
)

var T *testing.T

func nicePrint(i interface{}, t *testing.T) {
	p, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(p))
}

// change requestKMS for tests
func init() {
	requestKMS = func(req map[string]interface{}) <-chan Response {
		clientId++
		req["id"] = clientId
		response := Response{}
		response.Id = req["id"].(float64)

		nicePrint(req, T)

		if req["method"] == "create" {
			params := req["params"].(map[string]interface{})
			if params != nil && params["type"].(string) == "WebRtcEndpoint" {
				cp := params["constructorParams"].(map[string]interface{})
				response.Result = map[string]string{
					"value": cp["mediaPipeline"].(string) + "/ID_FOR_WEBRTCENDPOINT",
				}
				// Build a webrtc response
			}
		}

		ret := make(chan Response)
		go func() {
			ret <- response
		}()
		return ret
	}
}

// Test MediaPipeline creation
func TestCreateMediaPipeline(t *testing.T) {

	T = t
	debug = true
	pipeline := NewMediaPipeline()
	pipeline.Id = "uuid_of_mediapipeline"
	nicePrint(pipeline, t)

}

// Test WebRTCEndpoint creation
func TestWebRtcEndpoint(t *testing.T) {
	T = t
	debug = true
	pipeline := NewMediaPipeline()
	pipeline.Id = "uuid_of_mediapipeline"

	wrtc := &WebRtcEndpoint{}
	pipeline.Create(wrtc, map[string]interface{}{"test": "val"})

	nicePrint(wrtc, t)

}

//Test to connect sink on webrtc
func TestConnectWebRTCEndpointSink(t *testing.T) {
	T = t
	debug = true
	pipeline := NewMediaPipeline()
	pipeline.Id = "uuid_of_mediapipeline"

	wrtc := &WebRtcEndpoint{}
	pipeline.Create(wrtc, map[string]interface{}{"test": "val"})

	wrtc.Connect(wrtc, "", "", "")
	nicePrint(wrtc, t)

}
