package main

import (
	"github.com/LogsExporter/cmd"
	_ "github.com/LogsExporter/cmd/docker"
	_ "github.com/LogsExporter/cmd/elasticsearch"
)

func main() {
	cmd.Execute()
}
