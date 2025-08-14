package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Hassan-Ibrahim-1/go-ssg/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ssg [directory]")
		return
	}

	addr := ":4200"
	s, err := server.New(addr, os.Args[1])
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
}
