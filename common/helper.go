package common

import (
	"fmt"
	"strconv"
	"strings"
)

func ConvertRate(rateString string) string {
	rate, _ := strconv.ParseFloat(rateString, 64)
	rate = rate / 1000000000000000000.0
	rateString = fmt.Sprintf("%g", rate)
	if !strings.Contains(rateString, ".") {
		rateString = rateString + ".0"
	}
	return rateString
}

func Include(arr []string, el string) bool {
	for _, elm := range arr {
		if elm == el {
			return true
		}
	}
	return false
}
