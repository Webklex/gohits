package main

import (
	"./server"
	"./utils/config"
	"flag"
	"fmt"
	_ "github.com/elazarl/go-bindata-assetfs"
)

var buildNumber string
var buildVersion string

func main() {
	c := config.DefaultConfig()
	c.AddFlags(flag.CommandLine)

	sv := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	c.Build = config.Build{
		Number:  buildNumber,
		Version: buildVersion,
	}

	if *sv {
		fmt.Printf("GoHits version: %s\n", c.Build.Version)
		fmt.Printf("GoHits build number: %s\n", c.Build.Number)
		return
	}

	c.Load(c.File)

	if c.SaveConfigFlag {
		if _, err := c.Save(); err != nil {
			print(err)
		}
	}

	s := server.NewServerConfig(c, assetFS())
	s.Start()
}
