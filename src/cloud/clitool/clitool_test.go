package clitool

import (
	"log"
	"testing"
)

func TestJson(t *testing.T) {
	log.Println(GetConfig("1.json"))
}

func TestCli(t *testing.T) {
	RunCli()
}
