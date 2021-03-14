package clitool

import (
	"dockerapigo/src/common/message"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func readFile(fileName string) []byte {
	// Open our jsonFile
	jsonFile, err := os.Open(fileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	ret, _ := ioutil.ReadAll(jsonFile)
	return ret
}

func GetPodConfig(fileName string) message.PodConfig {
	var podCfg message.PodConfig
	content := readFile(fileName)
	err := json.Unmarshal(content, &podCfg)
	if err != nil {
		log.Println(err)
	}
	return podCfg
}

func GetConfig(fileName string) message.ConfigFile {
	var config message.ConfigFile
	content := readFile(fileName)
	err := json.Unmarshal(content, &config)
	if err != nil {
		log.Println(err)
	}
	return config
}
