package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"time"

	"github.com/dreamoeu/domains2providers/raw"
	"github.com/dreamoeu/domains2providers/rule"
)

type RuleSetInfo struct {
	Name     string `json:"name"`
	Behavior string `json:"behavior"`
	Count    int    `json:"count"`
}

type RuleSetCollection struct {
	UpdateTime time.Time     `json:"updateTime"`
	Count      int           `json:"count"`
	RuleSets   []RuleSetInfo `json:"ruleSets"`
}

func main() {
	if len(os.Args) < 3 {
		println("Usage: <v2ray-domains-path> <output-path>")

		os.Exit(1)
	}

	data := path.Join(os.Args[1], "data")
	generated := os.Args[2]

	_ = os.MkdirAll(generated, 0755)

	rulesetCollection := RuleSetCollection{
		UpdateTime: time.Now(),
		Count:      0,
		RuleSets:   make([]RuleSetInfo, 0),
	}

	err := writeRulesets(data, generated, &rulesetCollection)
	if err != nil {
		fmt.Println("Error writing rulesets:", err.Error())

		os.Exit(1)
	}

	err = writeRawRulesets(&rulesetCollection, generated)
	if err != nil {
		fmt.Println("Error writing raw rulesets:", err.Error())

		os.Exit(1)
	}

	// Write RuleSetCollection to json file
	rulesetCollection.Count = len(rulesetCollection.RuleSets)
	sort.Slice(rulesetCollection.RuleSets, func(i, j int) bool {
		if rulesetCollection.RuleSets[i].Behavior == "ipcidr" && rulesetCollection.RuleSets[j].Behavior == "domain" {
			return true
		} else if rulesetCollection.RuleSets[i].Behavior == "domain" && rulesetCollection.RuleSets[j].Behavior == "ipcidr" {
			return false
		} else if rulesetCollection.RuleSets[i].Behavior == rulesetCollection.RuleSets[j].Behavior {
			return rulesetCollection.RuleSets[i].Name < rulesetCollection.RuleSets[j].Name
		} else {
			return rulesetCollection.RuleSets[i].Behavior < rulesetCollection.RuleSets[j].Behavior
		}
	})

	jsonBytes, err := json.Marshal(rulesetCollection)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err.Error())

		os.Exit(1)
	}

	outputPath := path.Join(generated, "rulesets.json")
	errs := ioutil.WriteFile(outputPath, jsonBytes, 0644)
	if errs != nil {
		fmt.Println("Error writing file:", err.Error())

		os.Exit(1)
	}
}

func writeRulesets(data string, generated string, rulesetCollection *RuleSetCollection) error {
	ruleSets, err := rule.ParseDirectory(data)
	if err != nil {
		return fmt.Errorf("parse directory %s: %s", data, err.Error())
	}

	for name := range ruleSets {
		tags, err := rule.Resolve(ruleSets, name)
		if err != nil {
			return fmt.Errorf("resolve error %s: %s", name, err.Error())
		}

		for tag, rules := range tags {
			var outputPath string

			if tag == "" {
				outputPath = path.Join(generated, fmt.Sprintf("%s.yaml", name))

				ruleSetInfo := RuleSetInfo{Name: fmt.Sprintf("%s", name), Behavior: "domain", Count: len(rules)}
				rulesetCollection.RuleSets = append(rulesetCollection.RuleSets, ruleSetInfo)
			} else {
				outputPath = path.Join(generated, fmt.Sprintf("%s@%s.yaml", name, tag))

				ruleSetInfo := RuleSetInfo{Name: fmt.Sprintf("%s@%s", name, tag), Behavior: "domain", Count: len(rules)}
				rulesetCollection.RuleSets = append(rulesetCollection.RuleSets, ruleSetInfo)
			}

			file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("failed to write file %s: %s", outputPath, err.Error())
			}

			_, _ = file.WriteString(fmt.Sprintf("# NAME: %s\n", name))
			_, _ = file.WriteString(fmt.Sprintf("# BEHAVIOR: domain\n"))
			_, _ = file.WriteString(fmt.Sprintf("# SOURCE: https://github.com/v2fly/domain-list-community/tree/master/data/%s\n", name))
			_, _ = file.WriteString(fmt.Sprintf("# COUNT: %d\n", len(rules)))
			_, _ = file.WriteString(fmt.Sprintf("# UPDATED: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
			_, _ = file.WriteString(fmt.Sprintf("payload:\n"))

			for _, domain := range rules {
				_, _ = file.WriteString(fmt.Sprintf("- \"%s\"\n", domain))
			}

			_ = file.Close()
		}
	}

	return nil
}

func writeRawRulesets(rulesetCollection *RuleSetCollection, generated string) error {
	raws, err := raw.LoadRawSources()
	if err != nil {
		return fmt.Errorf("load raw resources %s", err.Error())
	}

	dns, err := raw.LoadDNSPollutedIPs()
	if err != nil {
		return fmt.Errorf("load DNS polluted ips: %s", err.Error())
	}

	tailscaleNodes, err := raw.LoadTailscaleNodes()
	if err != nil {
		return fmt.Errorf("load tailscale node ips: %s", err.Error())
	}

	ianaIPv4Specials, err := raw.LoadIANASpecials(true)
	if err != nil {
		return fmt.Errorf("load inna ipv4 special ips: %s", err.Error())
	}

	ianaIPv6Specials, err := raw.LoadIANASpecials(false)
	if err != nil {
		return fmt.Errorf("load inna ipv6 special ips: %s", err.Error())
	}

	raws = append(raws, dns...)
	raws = append(raws, tailscaleNodes...)
	raws = append(raws, ianaIPv4Specials...)
	raws = append(raws, ianaIPv6Specials...)

	for _, r := range raws {
		ruleSetInfo := RuleSetInfo{Name: r.Name, Behavior: r.Behavior, Count: len(r.Rules)}
		rulesetCollection.RuleSets = append(rulesetCollection.RuleSets, ruleSetInfo)

		outputPath := path.Join(generated, r.Name+".yaml")

		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %s", outputPath, err.Error())
		}

		_, _ = file.WriteString(fmt.Sprintf("# NAME: %s\n", r.Name))
		_, _ = file.WriteString(fmt.Sprintf("# BEHAVIOR: %s\n", r.Behavior))
		_, _ = file.WriteString(fmt.Sprintf("# SOURCE: %s\n", r.SourceUrl))
		_, _ = file.WriteString(fmt.Sprintf("# COUNT: %d\n", len(r.Rules)))
		_, _ = file.WriteString(fmt.Sprintf("# UPDATED: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
		_, _ = file.WriteString(fmt.Sprintf("payload:\n"))

		for _, domain := range r.Rules {
			_, _ = file.WriteString(fmt.Sprintf("- \"%s\"\n", domain))
		}

		_ = file.Close()
	}

	return nil
}
