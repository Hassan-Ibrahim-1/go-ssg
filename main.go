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
		fmt.Println("usage: ssg [directory]")
		return
	}

	nodes, err := site.Build(os.Args[1])
	if err != nil {
		log.Fatalln("Failed to build site:", err)
	}

	fmt.Println(nodes)

	addr := ":4200"
	s := server.New(addr, nodes)
	defer s.Close()

	fmt.Println("listening on", addr)
	err = s.ListenAndServe()

	if err != nil {
		log.Fatalln("server failed:", err)
	}
}
