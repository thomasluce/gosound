package main

import (
	"go/build"
	"log"
	"os"

	"github.com/thomasluce/gosound"
	_ "github.com/thomasluce/gosound/wav"
	"github.com/thomasluce/util"
)

func importPathToDir(importPath string) (string, error) {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}

func main() {
	util.SetLoggingLevel(util.DEBUG)
	filename := os.Args[1]

	dir, err := importPathToDir("github.com/thomasluce/gosound")
	if err != nil {
		log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	}
	util.Debugf("Changing to %s\n", dir)
	err = os.Chdir(dir)
	if err != nil {
		log.Panicln("os.Chdir:", err)
	}

	err = gosound.Init()
	if err != nil {
		panic(err)
	}

	err = gosound.Play(filename)
	if err != nil {
		panic(err)
	}

	for {
		if !gosound.Playing() {
			break
		}
	}
}
