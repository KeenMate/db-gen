package main

import (
	_ "embed"
	"github.com/keenmate/db-gen/cmd"
)

// due to the way go embed works, we can only embed file from same folder

//go:embed version.txt
var VersionFile string

func main() {
	cmd.Execute(VersionFile)
}
