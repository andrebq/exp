package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"time"
)

var help = flag.Bool("h", false, "Help")
var rand32 = flag.Bool("rand32", false, "Print a 32 bit integer random number to stdout")
var now32 = flag.Bool("now32", false, "Print the 32 bits of time.Now().Unixnano to stdout")
var echoHttp = flag.Bool("echoHttp", false, "Start the http echo server")
var echoHttpAddr = flag.String("echoHttpAddr", "0.0.0.0:9090", "Address to start the http echo server")
var splitImage = flag.String("splitImage", "", "SpritDecomposer xml output file")

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(1)
	}

	if *rand32 {
		processRand32()
	}
	if *now32 {
		processNow32()
	}
	if *splitImage != "" {
		splitIt(*splitImage)
	}
	if *echoHttp {
		startEchoServer(echoHttpAddr)
	}

	os.Exit(0)
}

func processRand32() {
	val, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		log.Fatalf("Error creating a random 32 bit number. %v", err)
	}
	fmt.Fprintf(os.Stdout, "%v\n", int32(val.Int64()))
}

func processNow32() {
	now := time.Now().UnixNano()
	fmt.Fprintf(os.Stdout, "%v\n", int32(now>>32))
}
