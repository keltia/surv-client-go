// surv-export.go
//
// Export data from the WS-N endpoint giving out ADS-B data in various formats.
//
// @author Ollivier Robert <ollivier.robert@eurocontrol.int>

package main

import (
	"path/filepath"
	"flag"
	"fmt"
	"os"
	"log"
	"os/signal"
	"time"
	"strconv"
	"regexp"
	"github.com/keltia/wsn-go/config"
	"github.com/keltia/wsn-go/wsn"
)

var (
	// We use a tag to find the proper file now. $HOME/.<tag>/config.toml
	RcFile = "surveillance"

	// All possible feeds
	Feeds = map[string]string{
		"AsterixJSON": "feed_json",
		"AsterixXML": "feed_xml",
		"AsterixJSONgzipped": "feed_jsongz",
	}

	timeMods = map[string]int64{
		"mn": 60,
		"h":  3600,
		"d":  24 *3600,
	}

	RunningFeeds = map[string]string{}

	client    *wsn.Client

	fOutputFH	*os.File
)

// Subscribe to wanted topics
func doSubscribe(feeds map[string]string) {
	time.Sleep(1 * time.Second)
	// Go go go
	for name, target := range feeds {
		unsubFn, err := client.Subscribe(name, target)
		if err != nil {
			log.Printf("Error subscribing to %n: %v", name, err)
		}
		if fVerbose {
			log.Printf("Subscribing to /%s for %s\n", target, name)
			log.Printf("  unsub is %s\n", unsubFn)
		}
		topic := wsn.Topic{Started: true, UnsubAddr: unsubFn}
		client.Topics[name] = &topic
	}
}

// Handle shutdown operations
func doShutdown() {
	// do last actions and wait for all write operations to end
	for name, topic := range (client.Topics) {
		if topic.Started {
			err := client.Unsubscribe(name)
			if err != nil {
				log.Printf("Error unsubscribing to %n: %v", name, err)
			}
			if fVerbose {
				log.Println("Unsubscribing from", name)
				log.Printf("Topic: %s Bytes: %d Pkts: %d", name, topic.Bytes, topic.Pkts)
			}
		}
	}
}

// return list of keys of map m
func keys(m map[string]string) []string {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}

// Check for specific modifiers, returns seconds
//
//XXX could use time.ParseDuration except it does not support days.
func checkTimeout(value string) int64 {
	mod := int64(1)
	re := regexp.MustCompile(`(?P<time>\d+)(?P<mod>(s|mn|h|d)*)`)
	match := re.FindStringSubmatch(value)
	if match == nil {
		return 0
	} else {
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
		return time * mod
	}
}

// Main program
func main() {
	// Handle SIGINT
	go func() {
	    sigint := make(chan os.Signal, 3)
	    signal.Notify(sigint, os.Interrupt)
	    <-sigint
	    log.Println("Program killed !")

		doShutdown()

	    os.Exit(0)
	}()

	flag.Parse()

	if fVerbose {
		fmt.Printf("%s version %s\n", filepath.Base(os.Args[0]), SURV_VERSION)
	}

	if len(flag.Args()) == 0 {
		fmt.Fprint(os.Stderr, "You must specify at least one feed!\n")
		fmt.Fprintln(os.Stderr, "List of possible feeds:")
		for f, _ := range Feeds {
			fmt.Fprintf(os.Stderr, "  %s\n", f)
		}
		os.Exit(1)
	}

	// Load configuration
	c, err := config.LoadConfig(RcFile)
	if err != nil {
		log.Fatalf("Error loading %s: %v", RcFile, err)
	}
	if fVerbose {
		fmt.Printf("Config is %s://%s:%d/%s\n", c.Proto, c.Site, c.Port, c.Endpoint)
		fmt.Println(c.Dests)
		fmt.Println(c.Default, c.Dests[c.Default])
	}

	// Actually instanciate the client part
	if client, err = wsn.NewClient(c); err != nil {
		log.Fatalf("Error connecting to %s: %v", client.Target)
	}

	if fVerbose {
		client.Verbose = true
	}

	// Handle other destinations.
	if fDest != "" {
		client.Config.Default = fDest
	}

	// Look for the feed name on CLI
	if len(flag.Args()) > 1 {
		fmt.Errorf("Error: only one feed is allowed")
		os.Exit(1)
	}

	tn := flag.Arg(0)
	if Feeds[tn] != "" {
		if fVerbose {
			log.Println("Configuring", Feeds[tn], "for", tn)
		}
		RunningFeeds[tn] = Feeds[tn]
		client.AddFeed(tn)
	}

	// Open output file
	if (fOutput != "") {
		if (fVerbose) {
			log.Printf("Output file is %s\n", fOutput)
		}

		if fOutput != "-" {
			if fOutputFH, err = os.Create(fOutput); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating %s\n", fOutput)
				panic(err)
			}
		} else {
			// stdout
			fOutputFH = os.Stdout
		}

		// select output callback
		switch tn {
		case "AsterixXML":
			client.AddHandler(fileOutputXML)
		case "AsterixJSON":
			client.AddHandler(fileOutputJSON)
		case "AsterixJSONgzipped":
			client.AddHandler(fileOutputJSON)
		}
	}

	// Check if we did specify a timeout with -i
	if fsTimeout != "" {
		fTimeout = checkTimeout(fsTimeout)

		if fVerbose {
			log.Printf("Run for %ds\n", fTimeout)
		}
		client.SetTimer(fTimeout)
	}

	// Start server for callback
	log.Println("Starting server for ", keys(RunningFeeds), "...")
	go doSubscribe(RunningFeeds)
	client.ServerStart(RunningFeeds)
}
