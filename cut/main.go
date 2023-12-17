package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	f := flag.String("f", "", "field to cut")
	delimiter := flag.String("d", "\t", "delimiter")
	flag.Parse()
	fileName := flag.Arg(0)
	var file *os.File

	if fileName == "" || fileName == "-" {
		file = os.Stdin
	} else {
		fi, err := os.Open(fileName)

		if err != nil {
			fmt.Printf("err opening file: %s", err)
			return
		}

		file = fi
	}

	fields := []int{}
	field, err := strconv.Atoi(*f)
	if err != nil {
		if strings.Contains(*f, ",") {
			for _, v := range strings.Split(*f, ",") {
				num, err := strconv.Atoi(v)
				if err != nil {
					fmt.Printf("err during conv: %s", err)
					return
				}
				fields = append(fields, num)
			}
		} else {
			for _, v := range strings.Split(*f, " ") {
				num, err := strconv.Atoi(v)
				if err != nil {
					fmt.Printf("err during conv: %s", err)
					return
				}
				fields = append(fields, num)
			}
		}
	} else {
		fields = append(fields, field)
	}

	reader := bufio.NewScanner(file)

	for reader.Scan() {
		line := reader.Text()
		var res []string
		res = strings.Split(line, *delimiter)
		if len(res) == 1 {
			fmt.Print(line)
			continue
		}

		a := []string{}

		for _, v := range fields {
			a = append(a, res[v-1])
		}

		fmt.Println(strings.Join(a, *delimiter))
	}
}
