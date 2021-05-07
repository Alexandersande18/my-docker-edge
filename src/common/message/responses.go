package message

import "encoding/json"

type PodCreateResponse struct {
	Cid     string `json:"cid"`
	Success bool   `json:"success"`
}

type PodQueryResponse struct {
	PodName string `json:"pod_name"`
	Status  string `json:"status"`
	Image   string `json:"image"`
	PortMap string `json:"port_map"`
	ID      string `json:"id"`
}

type PodListResponse struct {
	NodeID string   `json:"node_id"`
	Pods   []string `json:"pods"`
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

func ReadPodQueryResponse(message *Message) PodQueryResponse {
	configMap := message.GetContent()
	jsonString, _ := json.Marshal(configMap)
	res := PodQueryResponse{}
	json.Unmarshal(jsonString, &res)
	return res
}

func ReadPodListResponse(message *Message) PodListResponse {
	configMap := message.GetContent()
	jsonString, _ := json.Marshal(configMap)
	res := PodListResponse{}
	json.Unmarshal(jsonString, &res)
	return res
}
