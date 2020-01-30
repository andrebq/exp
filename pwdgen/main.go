// pwdgen generates passwords for humans or machines
package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	diceware "github.com/sethvargo/go-diceware/diceware"
)

func main() {
	target := flag.String("to", "machine", "Generate passwords for machines or humans?")
	size := flag.Int("size", 32, "How many bytes/words should be generated")
	b64 := flag.Bool("b26", true, "When generating passwords for machines, encode it as base64 instead of hex")
	sep := flag.String("sep", " ", "Character used to join words. Whitespace gives the best interactive experience but might be tricky to use in scripts (use - for such cases)")
	flag.Parse()

	switch *target {
	case "machine", "m":
		generateMachine(*size, *b64)
	case "human", "h":
		generateHuman(*size, *sep)
	default:
		flag.Usage()
	}
}

func generateMachine(sz int, b64 bool) {
	buf := make([]byte, sz)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	if b64 {
		fmt.Fprintf(os.Stdout, "%v\n", base64.StdEncoding.EncodeToString(buf))
		return
	}

	fmt.Fprintf(os.Stdout, "%v\n", hex.EncodeToString(buf))
}

func generateHuman(sz int, sep string) {
	words, err := diceware.GenerateWithWordList(sz, diceware.WordListEffLarge())
	if err != nil {
		panic(err)
	}
	phrase := strings.Join(words, sep)
	fmt.Fprintf(os.Stdout, "%v\n", phrase)
}
