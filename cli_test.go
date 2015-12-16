package main

import (
	"testing"
)

func TestCheckTimeout(t *testing.T) {
	var testValues = map[string]int64{
		"42": 42,
		"20mn": 20*60,
		"1h": 3600,
		"1d": 24*3600,
	}

	for str, val := range testValues {
		if chk := checkTimeout(str); chk != val {
			t.Errorf("Error: wrong value parsed: %d - %d", chk, val)
		}
	}
}
