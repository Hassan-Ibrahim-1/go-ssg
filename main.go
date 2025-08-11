package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Hassan-Ibrahim-1/go-ssg/markdown"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ssg [directory]")
	}

	md, err := os.ReadFile("content/blog.md")
	if err != nil {
		log.Fatalln(err)
	}

	html := markdown.ToHTML(md)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(html))
}
