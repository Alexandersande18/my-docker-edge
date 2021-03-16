package main

import (
	"context"
	"dockerapigo/src/common/config"
	"dockerapigo/src/edge/edgecontroller"
	"flag"
	"fmt"
)

func main() {
	hostApiUrl := "tcp://localhost:2375"
	masterIP := flag.String("m", "127.0.0.1", "master IP")
	nodeID := flag.String("n", "n01", "nodeID")
	flag.Parse()
	fmt.Println(*masterIP, *nodeID)
	ec := edgecontroller.NewEdgeController(*masterIP, config.MasterPort, *nodeID, hostApiUrl, context.Background())
	ec.Run()
}
