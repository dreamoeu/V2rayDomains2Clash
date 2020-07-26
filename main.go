package main

import (
	"bufio"
	"fmt"
	"github.com/kr328/domains2providers/trie"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
)

func main() {
	data := path.Join(".", "domain-list-community", "data")
	generated := path.Join(".", "generated")

	_ = os.MkdirAll(generated, 0755)

	files, err := ioutil.ReadDir(data)
	if err != nil {
		panic(err.Error())
	}

	for _, file := range files {
		t := trie.New()

		if err := appendFileToTrie(data, file.Name(), t); err != nil {
			panic(err.Error())
		}

		output, err := os.OpenFile(path.Join(generated, file.Name()+".yaml"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(err.Error())
		}

		header := fmt.Sprintf("# Generated from https://github.com/v2ray/domain-list-community/blob/master/data/%s\n\npayload:\n", file.Name())

		if _, err := output.WriteString(header); err != nil {
			panic(err.Error())
		}

		domains := t.Dump()

		sort.Strings(domains)

		for _, domain := range domains {
			if _, err := output.WriteString(fmt.Sprintf("  - \"%s\"\n", domain)); err != nil {
				panic(err.Error())
			}
		}

		_ = output.Close()
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

		seg := strings.SplitN(line, ":", 2)
		if len(seg) == 1 {
			_ = t.Insert(seg[0], "+.")
		} else {
			switch seg[0] {
			case "include":
				if err := appendFileToTrie(base, seg[1], t); err != nil {
					return err
				}
			case "full":
				_ = t.Insert(seg[1], "")
			default:
				println("Ignore unsupported " + line)
			}
		}
	}
}
