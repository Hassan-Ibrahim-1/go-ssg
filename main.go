package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Hassan-Ibrahim-1/go-ssg/markdown"
)

func main() {
	md, err := os.ReadFile("content/blog.md")
	if err != nil {
		log.Fatal(err)
	}

	html, err := markdown.ToHTML(md)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(html)
}
