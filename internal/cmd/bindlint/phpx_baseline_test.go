package main

import (
	"regexp"
	"strconv"
)

// phpxBindlintBaseline is the number of findings in the fork's own bindings
// when the gate arrived. Lower it in the commit that lowers the findings.
const phpxBindlintBaseline = 25

var phpxTotal = regexp.MustCompile(`(\d+) bindings point at a runtime value type`)

// phpxBindlintFindings reads the total off the report.
func phpxBindlintFindings(report string) int {
	m := phpxTotal.FindStringSubmatch(report)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}
