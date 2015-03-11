package kurento

import (
	"testing"
)

func TestCreateMediaPipeline(t *testing.T) {

	m := ServerClient{}
	pipeline := &MediaPipeline{}
	m.Create(pipeline)

}
