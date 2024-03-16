package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/practigo/gots"
)

func showNPackets(r gots.Reader, n int) {
	for i := 0; i < n; i++ {
		p, err := r.Next()
		if err != nil {
			panic(err)
		}

		header := p.ParseHeader()
		fmt.Printf("%v\n", header)
	}
}

func main() {
	filename := os.Args[1]

	var (
		err      error
		nPackets = 30
	)
	if len(os.Args) > 2 {
		nPackets, err = strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}
	}

	r, err := gots.NewReader(filename)
	if err != nil {
		panic(err)
	}

	showNPackets(r, nPackets)
}