package main

import (
	"fmt"
	"log"
	"os"

	"github.com/neallred/generalconference/internal/download"
	"github.com/neallred/generalconference/internal/scout"
)

const index_container_selector = "section .section-wrapper"
const conferencesDir = "conferences"

func logStep(x string) {
	fmt.Println(x)
}

func quitOnErr(err error, msgs ...string) {
	if err != nil {
		for _, msg := range msgs {
			log.Println(msg)
		}
		log.Fatal(err)
	}
}

func makeDownloadTarget(stepLabel string) {
	logStep(stepLabel)

	homedir, err := os.UserHomeDir()
	quitOnErr(err, "Failed to find home directory")
	downloadTarget := fmt.Sprintf("%s/%s", homedir, conferencesDir)

	err = os.MkdirAll(downloadTarget, os.ModePerm)
	quitOnErr(err, fmt.Sprintf("Failed to make download directory: \"%s\"", downloadTarget))
	fmt.Printf("Made %s\n", downloadTarget)
}

func main() {
	makeDownloadTarget("Making download folder")
	conferences := scout.Gather("Gathering conference links")
	chDownloadConferences := make(chan struct{}, 1)

	homedir, err := os.UserHomeDir()
	quitOnErr(err, "Failed to find home directory")
	downloadTarget := fmt.Sprintf("%s/%s", homedir, conferencesDir)
	for _, conf := range conferences {
		conf := conf
		go download.GetConference(conf, downloadTarget, chDownloadConferences)
		<-chDownloadConferences
	}

	fmt.Println("Conferences downloaded.")
}
