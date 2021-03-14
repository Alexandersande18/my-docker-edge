package message

import (
	"encoding/json"
)

type PodConfig struct {
	Node      string   `json:"node"`
	PodName   string   `json:"pod_name"`
	Group     string   `json:"group"`
	ImageName string   `json:"image_name"`
	PortsMap  []string `json:"ports_map"`
	MountsMap []string `json:"mounts_map"`
	EnvCfg    []string `json:"env"`
	HostsCfg  []string `json:"hosts_cfg,omitempty"`
}

type NodeRegisterInfo struct {
	LocalIP string `json:"local_ip"`
}

type ConfigFile struct {
	Services []Service   `json:"services"`
	PodCfgs  []PodConfig `json:"pods"`
}

type Service struct {
	Name     string `json:"name"`
	Node     string `json:"node"`
	Group    string `json:"group"`
	Pod      string `json:"pod"`
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
	LocalIP  string `json:"local_ip"`
}

func NewPodConfigMap(podname string, imagename string, portsmap []string, mountsmap []string, envcfg []string) *PodConfig {
	return &PodConfig{
		PodName:   podname,
		ImageName: imagename,
		PortsMap:  portsmap,
		MountsMap: mountsmap,
		EnvCfg:    envcfg,
	}
}
func ReadNodeRegister(message *Message) NodeRegisterInfo {
	configMap := message.GetContent()
	//log.Println(configMap)
	jsonString, _ := json.Marshal(configMap)
	res := NodeRegisterInfo{}
	json.Unmarshal(jsonString, &res)
	//log.Println(res)
	return res
}

func ReadPodConfigMap(message *Message) PodConfig {
	configMap := message.GetContent()
	jsonString, _ := json.Marshal(configMap)
	res := PodConfig{}
	json.Unmarshal(jsonString, &res)
	return res
}

func ReadServiceConfigMap(message *Message) Service {
	configMap := message.GetContent()
	jsonString, _ := json.Marshal(configMap)
	res := Service{}
	json.Unmarshal(jsonString, &res)
	return res
}
