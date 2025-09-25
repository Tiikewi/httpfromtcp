// Package headers is package for headers, duh
package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

const crfl = "\r\n"

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	endlineIndex := strings.Index(string(data), crfl)
	if endlineIndex == -1 {
		return 0, false, nil
	}

	if string(data[0:len(crfl)]) == crfl {
		return len(crfl), true, nil
	}

	trimmedData := strings.Trim(string(data[:endlineIndex]), " ")
	firstSemi := strings.Index(trimmedData, ":")
	key := trimmedData[:firstSemi]
	value := strings.Trim(trimmedData[firstSemi+1:], " ")

	if strings.Contains(key, " ") {
		return 0, false, fmt.Errorf("malformed header, invalid whitespace, data: %s, key: %s, value: %s", trimmedData, key, value)
	}

	h[key] = value
	consumedData := len(data[:endlineIndex]) + len(crfl)
	return consumedData, false, nil
}
