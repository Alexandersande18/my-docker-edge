package clitool

import (
	"dockerapigo/src/cloud/cloudcontroller"
	"fmt"
)

var (
	command, tag, file string
)

func RunCli() {
	cc := cloudcontroller.NewCloudController()
	go cc.RunController()
	for {
		//fmt.Print(">")
		n, err := fmt.Scanf("%s %s %s\n", &command, &tag, &file)
		if err != nil {
			fmt.Println(err)
		}
		if n == 3 {
			if tag == "-f" {
				cfg := GetPodConfig(file)
				cc.StartPod("0", "c01", cfg)
			}
		}
	}
}
