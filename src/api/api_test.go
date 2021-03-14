package api

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"log"
	"testing"
)

func TestApi(t *testing.T) {
	//hostUrl := "tcp://192.168.116.131:2375"
	hostUrl := "tcp://localhost:2375"
	cli, err := client.NewClientWithOpts(client.WithHost(hostUrl), client.WithAPIVersionNegotiation())
	ctx := context.Background()
	log.Println(err)
	cid, err := RunCbyName(ctx, cli, "mongo-express", "cc01", []string{"8081:8081"}, []string{},
		[]string{"ME_CONFIG_MONGODB_SERVER=dbnode",
			"ME_CONFIG_MONGODB_ADMINUSERNAME=root",
			"ME_CONFIG_MONGODB_ADMINPASSWORD=123"},
		[]string{"dbnode:192.168.116.128"})
	r, _ := GetContainerStatus(ctx, cli, "cc01")
	log.Println(GetPortMapString(r.HostConfig.PortBindings))
	fmt.Println(cid)
}
