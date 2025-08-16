package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Hassan-Ibrahim-1/go-ssg/server"
	"github.com/Hassan-Ibrahim-1/go-ssg/site"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(sprintUsage())
		return
	}

	action, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch action.typ {
	case DevServer:
		opts := action.devServerOpts
		addr := fmt.Sprintf(":%d", opts.port)
		s, err := server.New(addr, action.siteDir, opts.buildDrafts)
		if err != nil {
			log.Fatalln("failed to create server", err)
		}
		defer func() {
			if err := s.Close(); err != nil {
				log.Println("failed to close server:", err)
			}
		}()

		fmt.Println("listening on", addr)
		err = s.ListenAndServe()
		if err != nil {
			log.Fatalln("server failed:", err)
		}

	case BuildSite:
		opts := action.buildSiteOpts

		buildOpts := site.BuildOptions{BuildDrafts: opts.buildDrafts}
		s, err := site.Build(action.siteDir, buildOpts)
		if err != nil {
			log.Fatalln(err)
		}

		err = site.BuildSite(s, opts.buildDir)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
