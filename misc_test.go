package main

import (
	"testing"
	"sort"
)

func TestKeys(t *testing.T) {
	var dict = map[string]string{
		"foo": "junk",
		"bar": "don't care",
		"baz": "go to hell",
	}
	var myk = []string{"foo", "bar", "baz"}

	k := keys(dict)
	sort.Strings(myk)
	for k, v := range k {
		if v != myk[k] {
			t.Errorf("Error: wrong keys: %v - %v", v, myk[k])
		}
	}
}
