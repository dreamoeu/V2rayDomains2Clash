// from https://github.com/Dreamacro/clash/blob/dev/component/trie/trie.go

package trie

import (
	"errors"
	"strings"
)

const (
	wildcard        = "*"
	dotWildcard     = ""
	complexWildcard = "+"
	domainStep      = "."
)

var (
	// ErrInvalidDomain means insert domain is invalid
	ErrInvalidDomain = errors.New("invalid domain")
)

// DomainTrie contains the main logic for adding and searching nodes for domain segments.
// support wildcard domain (e.g *.google.com)
type DomainTrie struct {
	root *Node
}

func validAndSplitDomain(domain string) ([]string, bool) {
	if domain != "" && domain[len(domain)-1] == '.' {
		return nil, false
	}

	parts := strings.Split(domain, domainStep)
	if len(parts) == 1 {
		if parts[0] == "" {
			return nil, false
		}

		return parts, true
	}

	for _, part := range parts[1:] {
		if part == "" {
			return nil, false
		}
	}

	return parts, true
}

// Insert adds a node to the trie.
// Support
// 1. www.example.com
// 2. *.example.com
// 3. subdomain.*.example.com
// 4. .example.com
// 5. +.example.com
func (t *DomainTrie) Insert(domain string, data interface{}) error {
	parts, valid := validAndSplitDomain(domain)
	if !valid {
		return ErrInvalidDomain
	}

	if parts[0] == complexWildcard {
		t.insert(parts[1:], data)
		parts[0] = dotWildcard
		t.insert(parts, data)
	} else {
		t.insert(parts, data)
	}

	return nil
}

func (t *DomainTrie) insert(parts []string, data interface{}) {
	node := t.root
	// reverse storage domain part to save space
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if !node.hasChild(part) {
			node.addChild(part, newNode(nil))
		}

		node = node.getChild(part)
	}

	node.Data = data
}

// Search is the most important part of the Trie.
// Priority as:
// 1. static part
// 2. wildcard domain
// 2. dot wildcard domain
func (t *DomainTrie) Search(domain string) *Node {
	parts, valid := validAndSplitDomain(domain)
	if !valid || parts[0] == "" {
		return nil
	}

	n := t.search(t.root, parts)

	if n == nil || n.Data == nil {
		return nil
	}

	return n
}

func (t *DomainTrie) search(node *Node, parts []string) *Node {
	if len(parts) == 0 {
		return node
	}

	if c := node.getChild(parts[len(parts)-1]); c != nil {
		if n := t.search(c, parts[:len(parts)-1]); n != nil {
			return n
		}
	}

	if c := node.getChild(wildcard); c != nil {
		if n := t.search(c, parts[:len(parts)-1]); n != nil {
			return n
		}
	}

	if c := node.getChild(dotWildcard); c != nil {
		return c
	}

	return nil
}

func (t *DomainTrie) Dump() []string {
	result := make([]string, 0, 1024*10)

	t.dump(&result, "", t.root)

	index := 0

	for _, s := range result {
		if s == "" {
			continue
		}

		result[index] = s[:len(s)-1]

		index++
	}

	return result
}

func (t *DomainTrie) dump(domains *[]string, currentSegment string, node *Node) {
	if node.Data != nil || len(node.children) == 0 {
		*domains = append(*domains, currentSegment)

		return
	}

	for k, v := range node.children {
		t.dump(domains, k+"."+currentSegment, v)
	}
}

// New returns a new, empty Trie.
func New() *DomainTrie {
	return &DomainTrie{root: newNode(nil)}
}
