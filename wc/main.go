package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
    c := flag.Bool("c", false, "count bytes")
    flag.Parse()
    fileName := flag.Arg(0)

    file, err := os.ReadFile(fileName)
    if err != nil {
        fmt.Printf("err reading file: %s\n", err)
        os.Exit(1)
    }

    if *c {
        fmt.Printf("%d %s\n", len(file), fileName)
    }
}
