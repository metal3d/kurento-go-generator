package kurento

import (
	"encoding/json"
	"testing"
)

func nicePrint(i interface{}, t *testing.T) {
	p, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(p))
}

func TestCreateMediaPipeline(t *testing.T) {

	debug = true
	m := ServerClient{}
	pipeline := &MediaPipeline{}
	m.Create(pipeline, nil)

	pipeline.id = "uuid_of_mediapipeline"

	nicePrint(pipeline, t)

	wrtc := &WebRtcEndpoint{}
	pipeline.Create(wrtc, map[string]interface{}{"test": "val"})
	nicePrint(wrtc, t)

}
