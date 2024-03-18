package main

import (
	"os"
	"strconv"

	"github.com/practigo/gots"
)

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

	if err = gots.ShowPackets(filename, nPackets); err != nil {
		panic(err)
	}
}
