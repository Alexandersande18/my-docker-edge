package message

import "encoding/json"

type PodCreateResponse struct {
	Cid     string `json:"cid"`
	Success bool   `json:"success"`
}

type PodQuiryResponse struct {
	PodName string `json:"pod_name"`
	Status  string `json:"status"`
	Image   string `json:"image"`
	PortMap string `json:"port_map"`
}

type ErrorResponse struct {
	ErrorString string
}

func ReadPodCreateResponse(message *Message) PodCreateResponse {
	configMap := message.GetContent()
	jsonString, _ := json.Marshal(configMap)
	res := PodCreateResponse{}
	json.Unmarshal(jsonString, &res)
	return res
}

func ReadPodQuiryResponse(message *Message) PodQuiryResponse {
	configMap := message.GetContent()
	jsonString, _ := json.Marshal(configMap)
	res := PodQuiryResponse{}
	json.Unmarshal(jsonString, &res)
	return res
}
