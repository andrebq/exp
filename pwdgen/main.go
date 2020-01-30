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
	enc := flag.String("enc", "b64", "When generating passwords for machines, encode it as base64 instead of hex")
	sep := flag.String("sep", " ", "Character used to join words. Whitespace gives the best interactive experience but might be tricky to use in scripts (use - for such cases)")
	flag.Parse()

	switch *target {
	case "machine", "m":
		generateMachine(*size, encodingForName(*enc))
	case "human", "h":
		generateHuman(*size, *sep)
	default:
		flag.Usage()
	}
}

func encodingForName(name string) func([]byte) string {
	switch name {
	case "b64", "64":
		return base64.StdEncoding.EncodeToString
	case "hex", "bytes":
		return hex.EncodeToString
	default:
		fmt.Fprintf(os.Stderr, `Invalid encoding option: %v.
		Please use one of: b64|64|hex|bytes
`, name)
		flag.Usage()
		os.Exit(1)
	}
	panic("not reached")
}

func generateMachine(sz int, enc func([]byte) string) {
	buf := make([]byte, sz)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stdout, "%v\n", enc(buf))
}

func generateHuman(sz int, sep string) {
	words, err := diceware.GenerateWithWordList(sz, diceware.WordListEffLarge())
	if err != nil {
		panic(err)
	}
	phrase := strings.Join(words, sep)
	fmt.Fprintf(os.Stdout, "%v\n", phrase)
}
