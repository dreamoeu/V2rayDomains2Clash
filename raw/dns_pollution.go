package raw

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var dns = Raw{
	Name:      "dns-polluted-ips",
	Behavior:  "ipcidr",
	SourceUrl: "https://zh.m.wikiversity.org/zh-hans/%E9%98%B2%E7%81%AB%E9%95%BF%E5%9F%8E%E5%9F%9F%E5%90%8D%E6%9C%8D%E5%8A%A1%E5%99%A8%E7%BC%93%E5%AD%98%E6%B1%A1%E6%9F%93IP%E5%88%97%E8%A1%A8",
}

func LoadDNSPollutedIPs() ([]*RuleSet, error) {
	resp, err := http.Get(dns.SourceUrl)
	if err != nil {
		return nil, fmt.Errorf("load %s: %s", dns.Name, err.Error())
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("load %s: response %s", dns.Name, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("create doc %s: %s", dns.Name, err.Error())
	}

	_ = resp.Body.Close()

	var mf = doc.Find(".mf-section-1")
	if len(mf.Nodes) != 1 {
		return nil, fmt.Errorf("invalid mf %s: %d", dns.Name, len(mf.Nodes))
	}

	var mw = mf.First().Find(".mw-highlight")
	if len(mw.Nodes) != 2 {
		return nil, fmt.Errorf("invalid mw %s: %d", dns.Name, len(mw.Nodes))
	}

	var result []*RuleSet
	var rules []string

	mw.Each(func(j int, sel *goquery.Selection) {
		lines := strings.Split(sel.Find("pre").Text(), "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			sections := strings.Split(line, " ")

			if j == 0 {
				rules = append(rules, sections[0]+"/32")
			} else {
				rules = append(rules, sections[0]+"/128")
			}
		}
	})

	result = append(result, &RuleSet{
		Raw:   &dns,
		Rules: rules,
	})

	return result, nil
}
