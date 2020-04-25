package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3"
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
	talkCount := 0
	for _, conf := range conferences {
		for _, sess := range conf.Sessions {
			talkCount += len(sess.Talks)
		}
	}
	chDownloadConferences := make(chan int, 1)

	homedir, err := os.UserHomeDir()
	quitOnErr(err, "Failed to find home directory")
	downloadTarget := fmt.Sprintf("%s/%s", homedir, conferencesDir)
	fmt.Printf("Downloading talks:\n")
	bar := pb.StartNew(talkCount).SetWidth(80)
	for _, conf := range conferences {
		conf := conf
		go download.GetConference(conf, downloadTarget, chDownloadConferences)
		bar.Add(<-chDownloadConferences)
	}
	bar.Finish()

	fmt.Println("Conferences downloaded.")
}
