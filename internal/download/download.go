package download

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/neallred/generalconference/internal/scout"
)

func quitOnErr(err error, msgs ...string) {
	if err != nil {
		for _, msg := range msgs {
			log.Println(msg)
		}
		log.Fatal(err)
	}
}

var nonFileChars = regexp.MustCompile(`[^A-Za-z0-9 ]`)
var spaceChars = regexp.MustCompile(" +")
var dashes = regexp.MustCompile("-+")

func toFileName(str string) string {
	toSpaces := nonFileChars.ReplaceAll([]byte(str), []byte(" "))
	toDashes := spaceChars.ReplaceAll(toSpaces, []byte("-"))
	trimmedDashes := dashes.ReplaceAll(toDashes, []byte("-"))
	return strings.ToLower(strings.Trim(strings.TrimSpace(string(trimmedDashes)), "-"))
}

// yr, mth, ok
func getConferenceFolders(confLink string) (string, string, bool) {
	parsed, err := url.Parse(confLink)
	if err != nil {
		return "", "", false
	}
	path := strings.Split(parsed.Path, "/")
	if len(path) < 2 {
		return "", "", false
	}
	dates := path[len(path)-2:]

	yr := dates[0]
	mth := dates[1]
	ok := true

	return yr, mth, ok
}

func downloadTalk(link, filePath string, out chan<- struct{}) {
	// fmt.Println("dl", link)
	resp, err := http.Get(link)
	defer resp.Body.Close()
	quitOnErr(err, "Unable to load talk "+link)
	goqueryDoc, err := goquery.NewDocumentFromReader(resp.Body)
	quitOnErr(err, "Unable to load document "+link)
	article := goqueryDoc.Find("article")
	articleText := article.Contents().Text()
	err = ioutil.WriteFile(filePath, []byte(articleText), 0644)
	quitOnErr(err, "unable to write talk")
	out <- struct{}{}
}

func GetConference(conf scout.Conference, dlTarget string, out chan<- int) {
	talkCount := 0
	for _, sess := range conf.Sessions {
		talkCount += len(sess.Talks)
	}
	chDownloadTalks := make(chan struct{}, 100)

	// bootstrap year and month folders
	confLink := conf.Link
	yr, mth, ok := getConferenceFolders(confLink)
	if !ok {
		fmt.Println("failed to parse conference link for year and month")
	}
	confPath := fmt.Sprintf("%s/%s/%s", dlTarget, yr, mth)
	err := os.MkdirAll(confPath, os.ModePerm)
	quitOnErr(err, fmt.Sprintf("Failed to make conference directory: \"%s\"", confPath))

	for _, session := range conf.Sessions {
		session := session
		sessPath := fmt.Sprintf("%s/%s", confPath, toFileName(session.Title))
		err := os.MkdirAll(sessPath, os.ModePerm)
		quitOnErr(err, fmt.Sprintf("Failed to make session directory: \"%s\"", sessPath))
		for _, talk := range session.Talks {
			talk := talk
			talkPath := fmt.Sprintf("%s/%s.txt", sessPath, toFileName(talk.Title))
			go downloadTalk(talk.Link, talkPath, chDownloadTalks)

		}
		<-chDownloadTalks
	}

	out <- talkCount
}
