// cli.go
//
// Everything related to command-line flag handling
//
// Copyright 2015 © by Ollivier Robert for the EEC
//

package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"path/filepath"
)

var (
	timeMods = map[string]int64{
		"mn": 60,
		"h":  3600,
		"d":  24 *3600,
	}

	// cli
	fVerbose	bool
	fOutput		string
	fTimeout   int64
	fsTimeout  string
	fDest      string
)

// my usage string
const (
	cliUsage	= `%s version %s
Usage: %s [-o FILE] [-i N(s|mn|h|d)] [-v] [-d dest] feed

`
)

// Usage string override.
var Usage = func() {
		myName := filepath.Base(os.Args[0])
        fmt.Fprintf(os.Stderr, cliUsage, myName, SurvVersion, myName)
        flag.PrintDefaults()
}

// called by flag.Parse()
func init() {
	// cli
	flag.BoolVar(&fVerbose, "v", false, "Set verbose flag.")
	flag.StringVar(&fDest, "d", "", "Set default destination")
	flag.StringVar(&fsTimeout, "i", "60s", "Stop after N s/mn/h/days")
	flag.StringVar(&fOutput, "o", "", "Specify output FILE.")
}

// Check for specific modifiers, returns seconds
//
//XXX could use time.ParseDuration except it does not support days.
func checkTimeout(value string) (timeInSec int64) {
	mod := int64(1)
	re := regexp.MustCompile(`(?P<time>\d+)(?P<mod>(s|mn|h|d)*)`)
	match := re.FindStringSubmatch(value)
	if match == nil {
		return 0
	}
	// Get the base time
	time, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0
	}

	// Look for meaningful modifier
	if match[2] != "" {
		mod = timeMods[match[2]]
		if mod == 0 {
			mod = 1
		}
	}

	// At the worst, mod == 1.
	timeInSec = time * mod
	return
}
