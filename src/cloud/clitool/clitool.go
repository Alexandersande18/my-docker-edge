package clitool

import (
	"dockerapigo/src/cloud/cloudcontroller"
	"dockerapigo/src/common/message"
	"fmt"
	"log"
	"strings"
	"time"
)

var (
	command, tag, file string
	cc                 *cloudcontroller.CloudController
)

func handleServiceOptions(cfg message.ConfigFile) {
	for _, service := range cfg.Services {
		log.Println("cli new service")
		cc.NewService(service)
	}
}

func handlePodOptions(cfg message.ConfigFile) {
	for _, podcfg := range cfg.PodCfgs {
		log.Println("Starting pod", podcfg)
		cc.StartPod(podcfg)
		log.Println("pod start returns")
		time.Sleep(2 * time.Second)
	}
}

func nodeStatus(node string) {

}

func podStatus(pod string) {
	sep := strings.Split(pod, "-")
	//cc.PodStatusQuiry(sep[0], sep[1], sep[2])
	cc.AsyncPodStatusQuiry(sep[0], sep[1], sep[2])
}

func RunCli() {
	cc = cloudcontroller.NewCloudController()
	go cc.RunController()
	for {
		//fmt.Print(">")
		n, err := fmt.Scanf("%s %s %s\n", &command, &tag, &file)
		if err != nil {
			fmt.Println(err)
		}
		switch command {
		case "apply":
			if n == 3 && tag == "-f" {
				cfg := GetConfig(file)
				handleServiceOptions(cfg)
				handlePodOptions(cfg)
			}
			break
		case "status":
			if tag == "node" {
				nodeStatus(file)
			} else if tag == "pod" {
				podStatus(file)
			}
			break
		case "list":
			if tag == "node" {
				nodeStatus(file)
			} else if tag == "pod" {
				podStatus(file)
			}
		}
	}
}
