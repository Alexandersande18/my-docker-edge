package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
	"log"
	"os"
	"strings"
)

func NewApiClient(hostUrl string) (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.WithHost(hostUrl), client.WithAPIVersionNegotiation())
	return cli, err
}

func getAllImages(cli *client.Client) []types.ImageSummary {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	logger(err)
	for _, image := range images {
		fmt.Printf("%+v\n", image)
	}
	return images
}

func getAllContainers(ctx context.Context, cli *client.Client) []types.Container {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Quiet:   false,
		Size:    false,
		All:     true,
		Latest:  false,
		Since:   "",
		Before:  "",
		Limit:   0,
		Filters: filters.Args{},
	})

	if err != nil {
		log.Println(err)
	}
	return containers
}

func MakeNameToContainerMap(ctx context.Context, cli *client.Client) map[string]types.Container {
	ret := make(map[string]types.Container)
	var containers []types.Container
	containers = getAllContainers(ctx, cli)
	for _, c := range containers {
		var splitName = strings.Split(c.Names[0], "/")
		var lastName = splitName[len(splitName)-1]
		ret[lastName] = c
	}
	return ret
}

func FindContainerByName(ctx context.Context, cli *client.Client, containerName string) types.Container {
	var containerList = getAllContainers(ctx, cli)
	for _, ctr := range containerList {
		var splitName = strings.Split(ctr.Names[0], "/")
		var lastName = splitName[len(splitName)-1]
		if containerName == lastName {
			return ctr
		}
	}
	return types.Container{ID: "0"}
}

func GetContainerIP(container types.Container) string {
	settings := container.NetworkSettings.Networks["bridge"]
	return settings.IPAddress
}

func runContainer(ctx context.Context, cli *client.Client, cfg *container.Config, hostCfg *container.HostConfig, netCfg *network.NetworkingConfig, name string) (string, error) {
	fmt.Println(*cfg)
	resp, err := cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, nil, name)
	if err != nil {
		logger(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	return resp.ID, err
}

func GetContainerStatus(ctx context.Context, cli *client.Client, containerName string) (types.ContainerJSON, error) {
	ctr := FindContainerByName(ctx, cli, containerName)
	if ctr.ID == "0" {
		return types.ContainerJSON{}, errors.New("Container doesnot Exist")
	}
	res, err := cli.ContainerInspect(ctx, ctr.ID)
	log.Println(res.Config)
	return res, err
}

func RunCbyName(ctx context.Context, cli *client.Client, imageName string, containerName string, portsMap []string, mountsMap []string, envs []string, extraHost []string) (string, error) {
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		logger(err)
		return "", errors.New("image Pull Backoff")
	}
	io.Copy(os.Stdout, out)

	var cfg = container.Config{}
	var hostCfg = container.HostConfig{}
	cfg.Image = imageName

	ConfigPorts(&cfg, &hostCfg, portsMap)
	ConfigEnv(&cfg, envs)
	ConfigMounts(&hostCfg, mountsMap)
	ConfigHosts(&hostCfg, extraHost)
	hostCfg.RestartPolicy = container.RestartPolicy{
		Name:              "on-failure",
		MaximumRetryCount: 10,
	}

	cid, err := runContainer(ctx, cli, &cfg, &hostCfg, nil, containerName)
	if err != nil {
		logger(err)
	}
	return cid, err
}

func StopCbyName(ctx context.Context, cli *client.Client, containerName string) (string, error) {
	var ctnr = FindContainerByName(ctx, cli, containerName)
	if ctnr.ID == "0" {
		return "", errors.New("container Not Exist")
	}
	var err = cli.ContainerStop(ctx, ctnr.ID, nil)
	if err != nil {
		return ctnr.ID, err
	}
	return ctnr.ID, nil
}

func RemoveC(ctx context.Context, cli *client.Client, containerID string) error {
	var err = cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		RemoveLinks:   false,
		RemoveVolumes: false,
		Force:         true,
	})
	if err != nil {
		logger(err)
	}
	return err
}

func ConfigPorts(cfg *container.Config, hostCfg *container.HostConfig, portsM []string) {
	if len(portsM) == 0 {
		return
	}
	var exposedSet = make(nat.PortSet)
	var Maping = make(nat.PortMap)
	for _, portM := range portsM {
		var arr = strings.Split(portM, ":")
		cpt, err := nat.NewPort("tcp", arr[1])
		if err != nil {
			panic(err)
		}
		var portBinding = nat.PortBinding{HostIP: "0.0.0.0", HostPort: arr[0]}
		exposedSet[cpt] = struct{}{}
		Maping[cpt] = []nat.PortBinding{portBinding}
	}
	cfg.ExposedPorts = exposedSet
	hostCfg.PortBindings = Maping
}

func ConfigMounts(hostCfg *container.HostConfig, mountsM []string) {
	if len(mountsM) == 0 {
		return
	}
	var mountsHost []mount.Mount
	for _, v := range mountsM {
		var l = strings.Split(v, ":")
		var m = mount.Mount{
			Type:   "bind",
			Source: l[0],
			Target: l[1],
		}
		mountsHost = append(mountsHost, m)
	}
	hostCfg.Mounts = mountsHost
}

func ConfigHosts(hostCfg *container.HostConfig, extraHosts []string) {
	if len(extraHosts) == 0 {
		return
	}
	hostCfg.ExtraHosts = extraHosts
}

func ConfigEnv(cfg *container.Config, envs []string) {
	if len(envs) == 0 {
		return
	}
	cfg.Env = envs
}

func GetPortMapString(portMap nat.PortMap) string {
	ret := ""
	for hp, cp := range portMap {
		t := hp.Port() + ":" + cp[0].HostPort
		ret = ret + "," + t
	}
	return ret
}

func logger(err error) {
	if err != nil {
		log.Printf("%v\n", err)
		panic(err)
	}
}
