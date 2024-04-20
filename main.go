package main

import (
	_ "embed"
	"github.com/keenmate/db-gen/cmd"
)

//go:embed version.txt
var VersionFile string

func main() {
	cmd.Execute(VersionFile)
}
