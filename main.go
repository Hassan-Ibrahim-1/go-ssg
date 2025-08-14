package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Hassan-Ibrahim-1/go-ssg/site"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ssg [directory]")
		return
	}

	dir := os.Args[1]
	s, err := site.Build(dir)
	if err != nil {
		log.Fatalln(err)
	}

	err = site.BuildSite(s, "ssg-build")
	if err != nil {
		log.Fatalln(err)
	}

	// addr := ":4200"
	// s, err := server.New(addr, os.Args[1])
	// if err != nil {
	// 	log.Fatalln("failed to create server", err)
	// }
	// defer func() {
	// 	if err := s.Close(); err != nil {
	// 		log.Println("failed to close server:", err)
	// 	}
	// }()
	//
	// fmt.Println("listening on", addr)
	// err = s.ListenAndServe()
	// if err != nil {
	// 	log.Fatalln("server failed:", err)
	// }
}
