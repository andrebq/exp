package main

import (
	"flag"
	"os"
	"math/big"
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"time"
)

var help = flag.Bool("h", false, "Help")
var rand32 = flag.Bool("rand32", false, "Print a 32 bit integer random number to stdout")
var now32 = flag.Bool("now32", false, "Print the 32 bits of time.Now().Unixnano to stdout")

func main() {
	flag.Parse()
	if (*help) {
		flag.Usage()
		os.Exit(1)
	}

	if *rand32 {
		processRand32()
	}
	if *now32 {
		processNow32()
	}

	os.Exit(0)
}

func processRand32() {
	val, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		log.Fatalf("Error creating a random 32 bit number. %v", err)
	}
	fmt.Fprintf(os.Stdout, "%v", int32(val.Int64()))
}

func processNow32() {
	now := time.Now().UnixNano()
	fmt.Fprintf(os.Stdout, "%v", int32(now >> 32))
}
