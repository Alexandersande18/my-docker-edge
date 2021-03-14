package edgecontroller

import (
	"context"
	"dockerapigo/src/common/config"
	"testing"
)

func TestName(t *testing.T) {
	//hostApiUrl := "tcp://192.168.116.131:2375"
	hostApiUrl := "tcp://localhost:2375"
	ec := NewEdgeController("127.0.0.1", config.MasterPort, "c01", hostApiUrl, context.Background())
	ec.Run()
}
