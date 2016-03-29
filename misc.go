// misc.go

/*
This file implements various housekeeping functions.
 */
package main

import (
	"log"
	"github.com/keltia/wsn-go/wsn"
)

// Handle shutdown operations
func doShutdown(client *wsn.Client) {
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
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
