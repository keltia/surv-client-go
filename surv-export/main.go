// fa-export.go
//
//

package main

import (
	"../config"
	"../surv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"log"
	"os/signal"
	"time"
)

var (
	RcFile = filepath.Join(os.Getenv("HOME"), ".surveillance", "config.yml")

	// All possible feeds
	Feeds = map[string]string{
		"AsterixJSON": "feed_json",
		"AsterixXML": "feed_xml",
		"AsterixJSONgzipped": "feed_jsongz",
	}

	RunningFeeds = map[string]string{}

	SurvClient	*surv.Client

	fOutputFH	*os.File
)

// Subscribe to wanted topics
func doSubscribe(feeds map[string]string) {
	time.Sleep(1 * time.Second)
	// Go go go
	for name, target := range feeds {
		unsubFn, err := SurvClient.Subscribe(name, target)
		if err != nil {
			log.Printf("Error subscribing to %n: %v", name, err)
		}
		if fVerbose {
			log.Printf("Subscribing to /%s for %s\n", target, name)
			log.Printf("  unsub is %s\n", unsubFn)
		}
		topic := surv.Topic{Started: true, UnsubAddr: unsubFn}
		SurvClient.Topics[name] = topic
	}
}

// Handle shutdown operations
func doShutdown() {
	// do last actions and wait for all write operations to end
	for name, topic := range (SurvClient.Topics) {
		if topic.Started {
			err := SurvClient.Unsubscribe(name)
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

// fOutput callback
func file_output(buf []byte) {
	if nb, err := fOutputFH.Write(buf); err != nil {
		log.Fatalf("Error writing %d bytes: %v", nb, err)
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
		panic("Error loading "+RcFile)
	}
	if fVerbose {
		fmt.Printf("Config is %s://%s:%s/%s\n", c.Proto, c.Site, c.Port, c.Endpoint)
		fmt.Println(c.Dests)
		fmt.Println(c.Default, c.Dests[c.Default])
	}

	// Actually instanciate the client part
	SurvClient, err = surv.NewClient(c)

	// Open output file
	if (fOutput != "") {
		if (fVerbose) {
			log.Printf("Output file is %s\n", fOutput)
		}

		if fOutputFH, err = os.Create(fOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s\n", fOutput)
			panic(err)
		}

		SurvClient.AddHandler(file_output)
	}

	// Look for feed names on CLI
	for _, tn := range flag.Args() {
		if Feeds[tn] != "" {
			if fVerbose {
				log.Println("Configuring", Feeds[tn], "for", tn)
			}
			RunningFeeds[tn] = Feeds[tn]
			SurvClient.AddFeed(tn)
		}
	}

	// Start server for callback
	log.Println("Starting server for ", keys(RunningFeeds), "...")
	go doSubscribe(RunningFeeds)
	surv.ServerStart(SurvClient, RunningFeeds)
}
