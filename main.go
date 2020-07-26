package main

import (
	"bufio"
	"fmt"
	"github.com/kr328/domains2providers/trie"
	"io"
	"os"
	"path"
	"strings"
)

var generateFiles = map[string][]string{}

func init() {
	generateFiles["cn"] = []string{"cn"}
	generateFiles["!cn"] = []string{"category-scholar-!cn", "geolocation-!cn", "tld-!cn"}
}

func main() {
	data := path.Join(".", "domain-list-community", "data")
	generated := path.Join(".", "generated")

	_ = os.MkdirAll(generated, 0755)

	for output, inputs := range generateFiles {
		t := trie.New()

		for _, input := range inputs {
			if err := appendFileToTrie(data, input, t); err != nil {
				panic(err.Error())
			}
		}

		file, err := os.OpenFile(path.Join(generated, output + ".yaml"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(err.Error())
		}

		if _, err := file.WriteString("payload:\n"); err != nil {
			panic(err.Error())
		}

		domains := t.Dump()

		for _, domain := range domains {
			if _, err := file.WriteString(fmt.Sprintf("  - \"%s\"\n", domain)); err != nil {
				panic(err.Error())
			}
		}

		_ = file.Close()
	}
}

func appendFileToTrie(base, fileName string, t *trie.DomainTrie) error {
	file, err := os.Open(path.Join(base, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		lineBytes, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		line := string(lineBytes)

		line = strings.SplitN(line, "#", 2)[0]
		line = strings.SplitN(line, "@", 2)[0]
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "include:") {
			if err := appendFileToTrie(base, line[len("include:"):], t); err != nil {
				return err
			}
		} else {
			_ = t.Insert(line, true)
		}
	}
}