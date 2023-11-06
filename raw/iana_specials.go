package raw

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ianaIPv4Specials = Raw{
	Name:      "iana-ipv4-specials",
	Behavior:  "ipcidr",
	SourceUrl: "https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry-1.csv",
}

var ianaIPv6Specials = Raw{
	Name:      "iana-ipv6-specials",
	Behavior:  "ipcidr",
	SourceUrl: "https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry-1.csv",
}

func LoadIANASpecials(isV4 bool) ([]*RuleSet, error) {
	var ianaSpecials Raw
	if isV4 {
		ianaSpecials = ianaIPv4Specials
	} else {
		ianaSpecials = ianaIPv6Specials
	}

	resp, err := http.Get(ianaSpecials.SourceUrl)
	if err != nil {
		return nil, fmt.Errorf("load %s: %s", ianaSpecials.Name, err.Error())
	}

	if resp.StatusCode/100 != 2 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("load %s: response %s", ianaSpecials.Name, resp.Status)
	}

	defer resp.Body.Close()

	r := csv.NewReader(resp.Body)
	// Skip the header row
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("failed to parse csv file %s: %s", ianaSpecials.Name, err.Error())
	}

	var result []*RuleSet
	var rules []string

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse csv line %s: %s", ianaSpecials.Name, err.Error())
		}

		// Extract the IP CIDR from the first column
		ipCidr := strings.Split(strings.Trim(record[0], "\""), " ")[0]
		if strings.Contains(ipCidr, ",") {
			// There are multiple IP CIDRs separated by commas
			ipCidrs := strings.Split(ipCidr, ",")
			for _, c := range ipCidrs {
				c = strings.TrimSpace(c)
				if c != "" {
					rules = append(rules, c)
				}
			}
		} else {
			ipCidr = strings.TrimSpace(ipCidr)
			if ipCidr != "" {
				rules = append(rules, ipCidr)
			}
		}
	}

	result = append(result, &RuleSet{
		Raw:   &ianaSpecials,
		Rules: rules,
	})

	return result, nil
}
