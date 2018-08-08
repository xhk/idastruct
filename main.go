package main

import (
	"fmt"
	"os"
)

// 自动补充不完全struct的小工具

func usage() {

}

func main() {
	fmt.Println("start main")
	fmt.Printf("arg count:%d\n", len(os.Args))
	if len(os.Args) != 2 {
		usage()
		return
	}

	filePath := os.Args[1]

	var sp StructParser
	sp.ParseFile(filePath)
	//sp.PrintWords()
	sp.DumpFixedStructs()
}
