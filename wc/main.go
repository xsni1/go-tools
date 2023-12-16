package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

func main() {
	c := flag.Bool("c", false, "count bytes")
	l := flag.Bool("l", false, "count lines")
	w := flag.Bool("w", false, "count words")
	m := flag.Bool("m", false, "count chars")

	flag.Parse()
	fileName := flag.Arg(0)
	var file *os.File

	if fileName != "" {
		f, err := os.Open(fileName)

		if err != nil {
			fmt.Printf("err reading file: %s\n", err)
			os.Exit(1)
		}

		file = f
	} else {
		file = os.Stdin
	}

	lines := 0
	bytes := 0
	words := 0
	chars := 0

	f := bufio.NewReader(file)

	for {
		line, err := f.ReadString('\n')
		if err != nil {
			break
		}

		chars += utf8.RuneCountInString(line)
		bytes += len(line)
		lines++
		w := strings.Fields(line)
		words += len(w)
	}

	if !*c && !*l && !*w && !*m {
		fmt.Printf("%d %d %d %s\n", lines, words, bytes, fileName)
	}

	output := ""

	if *c {
		output += fmt.Sprint(bytes)
	}

	if *l {
		output += fmt.Sprintf(" %d", lines)
	}

	if *w {
		output += fmt.Sprintf(" %d", words)
	}

	if *m {
		output += fmt.Sprintf(" %d", chars)
	}

	fmt.Printf("%s %s\n", output, fileName)
}
