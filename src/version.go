package dbGen

import (
	_ "embed"
	"fmt"
	"log"
	"strings"
)

type BuildInformation struct {
	builder    string
	version    string
	commitHash string
}

var info *BuildInformation = nil

func ParseBuildInformation(versionFileText string) error {
	if strings.HasPrefix(versionFileText, "LOCAL") {
		log.Println("Running locally build version, be careful")
		return nil
	}

	splitted := strings.Split(versionFileText, " ")

	if len(splitted) < 2 {
		return fmt.Errorf("error with build, version file has incorrect format ")
	}

	info = &BuildInformation{
		builder:    splitted[0],
		version:    splitted[1],
		commitHash: splitted[2],
	}

	return nil
}

func PrintVersion() {
	if info == nil {
		log.Printf("Locally build version")
		return
	}
	log.Printf("db-gen build by %s ", info.builder)
	log.Printf("version %s ", info.version)
	log.Printf("last commit hash %s ", info.commitHash)
}

func GetVersion() string {
	if info == nil {
		return "LOCAL"
	} else {
		return info.version
	}
}
