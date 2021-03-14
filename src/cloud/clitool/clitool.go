package clitool

import (
	"dockerapigo/src/cloud/cloudcontroller"
	"dockerapigo/src/common/message"
	"fmt"
	"log"
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
	}
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
		if n == 3 {
			if tag == "-f" {
				cfg := GetConfig(file)
				handleServiceOptions(cfg)
				handlePodOptions(cfg)
			}
		}
	}
}
