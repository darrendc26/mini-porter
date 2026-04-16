package main

import (
	"github.com/darrendc26/mini-porter/cmd"
	"github.com/darrendc26/mini-porter/internal/logger"
)

func main() {
	logger.Init()
	cmd.Execute()
}
