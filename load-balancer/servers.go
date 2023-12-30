package main

import "strings"

func (b *backends) Set(val string) error {
	*b = append(*b, val)
	return nil
}

func (b *backends) String() string {
	return strings.Join(*b, " ")
}

type backends []string

var servers backends
