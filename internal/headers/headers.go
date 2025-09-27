// Package headers is package for headers, duh
package headers

import (
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

const crfl = "\r\n"

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	validCharacters, err := regexp.Compile("^[A-Za-z0-9!#$%&'*+-.^_`|~]+$")
	if err != nil {
		return 0, false, fmt.Errorf("regexp is invalid")
	}

	endlineIndex := strings.Index(string(data), crfl)
	if endlineIndex == -1 {
		return 0, false, nil
	}

	if string(data[0:len(crfl)]) == crfl {
		return len(crfl), true, nil
	}

	trimmedData := strings.Trim(string(data[:endlineIndex]), " ")
	firstSemi := strings.Index(trimmedData, ":")
	key := strings.ToLower(trimmedData[:firstSemi])
	value := strings.Trim(trimmedData[firstSemi+1:], " ")

	if !validCharacters.MatchString(key) {
		return 0, false, fmt.Errorf("header key contains invalid characters: %s", key)
	}

	if strings.Contains(key, " ") {
		return 0, false, fmt.Errorf("malformed header, invalid whitespace, data: %s, key: %s, value: %s", trimmedData, key, value)
	}

	if h[key] != "" {
		h[key] = h[key] + ", " + value
	} else {
		h[key] = value
	}

	consumedData := len(data[:endlineIndex]) + len(crfl)
	return consumedData, false, nil
}
