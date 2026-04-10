package main

import (
	"mini-porter/cmd"
	"mini-porter/internal/logger"
)

func main() {
	logger.Init()
	cmd.Execute()

}
