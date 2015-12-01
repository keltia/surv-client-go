// data.go

/*
  Decodes the different payloads.
 */
package main

import (
	"encoding/xml"
	"fmt"
	"log"
)

type PLJson struct {
	XMLName xml.Name `xml:"Cat62SurveillanceJSON"`
	PlainText struct {
		Data []byte `xml:",innerxml"`
	}
}

type PLXml struct {
	XMLName xml.Name `xml:"Cat62Surveillance"`
	Data []byte `xml:",innerxml"`
}

// fOutput JSON callback
func fileOutputJSON(buf []byte) {
	notify := &PLJson{}

	err := xml.Unmarshal(buf, notify)
	if err != nil {
		real := fmt.Sprintf("Error reading payload: %v/%v", buf, err)
		log.Println(real)
	} else {
		if fVerbose {
			log.Printf("payload size is %d\n", len(notify.PlainText.Data))
		}
		if nb, err := fOutputFH.Write(notify.PlainText.Data); err != nil {
			log.Fatalf("Error writing %d bytes: %v", nb, err)
		}
	}
}

// fOutput XML callback
func fileOutputXML(buf []byte) {
	notify := &PLXml{}

	err := xml.Unmarshal(buf, notify)
	if err != nil {
		real := fmt.Sprintf("Error reading payload: %v/%v", buf, err)
		log.Println(real)
	} else {
		if fVerbose {
			log.Printf("payload size is %d\n", len(notify.Data))
		}
		if nb, err := fOutputFH.Write(notify.Data); err != nil {
			log.Fatalf("Error writing %d bytes: %v", nb, err)
		}
	}
}

