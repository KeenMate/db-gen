package version

import (
	_ "embed"
	"fmt"
	"github.com/keenmate/db-gen/private/helpers"
	"log"
	"strings"
)

type BuildInformation struct {
	Builder    string
	Version    string
	CommitHash string
}

const localVersion = "LOCAL"
const localBuilder = "LOCAL"

var info = BuildInformation{
	Builder:    localBuilder,
	Version:    localVersion,
	CommitHash: "",
}

func ParseBuildInformation(versionFileText string) error {
	if strings.HasPrefix(versionFileText, "LOCAL") {
		helpers.LogWarn("THIS IS LOCAL BUILD")
		helpers.LogWarn("USE ONLY FOR TESTING!!!!")
		return nil
	}

	parts := strings.Split(versionFileText, " ")

	if len(parts) < 2 {
		return fmt.Errorf("error with build, Version file has incorrect format ")
	}

	info = BuildInformation{
		Builder:    parts[0],
		Version:    parts[1],
		CommitHash: parts[2],
	}

	return nil
}

func PrintVersion() {
	if info.Builder == localBuilder {
		log.Printf("Locally build Version")
		return
	}
	log.Printf("db-gen build by %s ", info.Builder)
	log.Printf("Version %s ", info.Version)
	log.Printf("last commit hash %s ", info.CommitHash)
}

func GetVersion() string {
	return info.Version
}

func GetBuildInfo() *BuildInformation {
	return &info
}

func IsLocalBuild() bool {
	return info.Version == localVersion
}
