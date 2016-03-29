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
	"github.com/keltia/wsn-go/config"
	"github.com/keltia/wsn-go/wsn"
)

var (
	// RcFile is the tag to find the proper file now. $HOME/.<tag>/config.toml
	RcFile = "surveillance"

	// Feeds lists all possible feeds
	Feeds = map[string]string{
		"AsterixJSON": "feed_json",
		"AsterixXML": "feed_xml",
		"AsterixJSONgzipped": "feed_jsongz",
	}

	runningFeeds = map[string]string{}

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

		doShutdown(client)

	    os.Exit(0)
	}()

	flag.Usage = Usage
	flag.Parse()

	if fVerbose {
		fmt.Printf("%s version %s\n\n", filepath.Base(os.Args[0]), SURV_VERSION)
	}

	if len(flag.Args()) == 0 {
		fmt.Fprint(os.Stderr, "You must specify at least one feed!\n")
		fmt.Fprintln(os.Stderr, "List of possible feeds:")
		for f := range Feeds {
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
		runningFeeds[tn] = Feeds[tn]
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
	log.Println("Starting server for ", keys(runningFeeds), "...")

	// Start subscriptions asynchronously w/ a delay
	go doSubscribe(runningFeeds)

	// Get the ball rolling
	client.ServerStart()
}
