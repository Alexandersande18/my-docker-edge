package clitool

import (
	"log"
	"testing"
)

func TestJson(t *testing.T) {
	log.Println(GetConfig("mongo.json"))
}

func TestCli(t *testing.T) {
	RunCli()
}
