package main

import (
	"log"
	"vk-training/pkg/cmd"
	"vk-training/pkg/config"
	"vk-training/pkg/db"
	"vk-training/pkg/util"
)

func main() {
	if err := util.InitLocalStorage(); err != nil {
		log.Fatalf("init storage: %v", err)
	}

	t := &db.Training{}
	
	if err := t.LoadFromFile(config.LocalFile); err != nil {
		log.Fatalf("load from file: %v", err)
	}

	cmd.Start(t)
}