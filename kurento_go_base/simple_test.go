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
			if params != nil {
				if t, ok := params["type"].(string); ok {
					if t == "WebRtcEndpoint" {
						cp := params["constructorParams"].(map[string]interface{})
						response.Result = map[string]string{
							"value": cp["mediaPipeline"].(string) + "/ID_FOR_WEBRTCENDPOINT",
						}
					}
					if t == "MediaPipeline" {
						response.Result = map[string]string{
							"value": "ID_FOR_MEDIAPIPELINE",
						}
					}
				}
			}
		}

		// fake KMS response in a chanel
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
	nicePrint(pipeline, t)

}

// Test WebRTCEndpoint creation
func TestWebRtcEndpoint(t *testing.T) {
	T = t
	debug = true
	pipeline := NewMediaPipeline()

	wrtc := &WebRtcEndpoint{}
	pipeline.Create(wrtc, nil)

	nicePrint(wrtc, t)

}

//Test to connect sink on webrtc
func TestConnectWebRTCEndpointSink(t *testing.T) {
	T = t
	debug = true
	pipeline := NewMediaPipeline()

	wrtc := &WebRtcEndpoint{}
	pipeline.Create(wrtc, nil)
	nicePrint(wrtc, t)

	wrtc.Connect(wrtc, "", "", "")
	nicePrint(wrtc, t)

}
