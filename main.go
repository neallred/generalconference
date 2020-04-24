package main

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
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const language = "eng"
const domain = "https://www.churchofjesuschrist.org"
const index_container_selector = "section .section-wrapper"
const conferencesDir = "conferences"

var indexLink = fmt.Sprintf("%s/general-conference/conferences?lang=%s", domain, language)

func slugToUrl(slug string) string {
	return fmt.Sprintf("%s%s", domain, slug)
}

type conference struct {
	link     string
	title    string
	sessions []session
	err      error
}

type session struct {
	title string
	talks []talk
	err   error
}

type talk struct {
	title  string
	author string
	link   string
	err    error
}

type result struct {
	link string
	res  http.Response
	err  error
}

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

func gatherConferences(stepLabel string) []conference {
	logStep(stepLabel)

	var conferences []conference

	resp, err := http.Get(indexLink)
	quitOnErr(err, "Unable to load General Conferences index page")
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	quitOnErr(err, "Unable parse General Conferences index page")

	quitOnErr(err, "Failed to find conference links on index page")
	doc.Find("a[href].year-line__link").Each(func(_ int, a *goquery.Selection) {
		href, _ := a.Attr("href")
		link := slugToUrl(href)
		title := strings.TrimSpace(a.Text())
		if strings.Contains(href, "/general-conference/") {
			var sessions []session
			conf := conference{link, title, sessions, nil}
			conferences = append(conferences, conf)
		}
	})

	fmt.Printf("Found %d conferences \n", len(conferences))
	return conferences
}

func summarizeTalk(markup *html.Node) talk {
	var talk_link string
	for _, attr := range markup.Attr {
		if attr.Key == "href" {
			talk_link = slugToUrl(attr.Val)
		}
	}

	author, author_err := htmlquery.Query(markup, "//div[contains(@class, 'lumen-tile__content')]/text()")
	talk_title, title_err := htmlquery.Query(markup, "//div[contains(@class, 'lumen-tile__title')]/div/text()")

	var effective_author string
	if author_err != nil || author == nil {
		fmt.Println("unknown author:", author_err, author, markup)
		effective_author = "unknown"
	} else {
		effective_author = author.Data
	}

	var effective_title string
	if title_err != nil || talk_title == nil {
		// Hackery because a handful of talks have a slightly different html structure
		talk_title, title_err = htmlquery.Query(markup, "//div[contains(@class, 'lumen-tile__title')]/text()")
		if title_err != nil || talk_title == nil {
			fmt.Println("unknown title:", title_err, talk_title, talk_link)
			effective_title = "unknown"
		}
	} else {
		effective_title = talk_title.Data
	}

	return talk{effective_title, effective_author, talk_link, author_err}
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

func addConferenceTalks(conf conference, out chan<- *conference) {
	doc, err := htmlquery.LoadURL(conf.link)
	if err != nil {
		log.Println("err on conference page")
		log.Println(err)
	}

	sessions_container, err := htmlquery.Query(doc, "//div[contains(@class,'section-wrapper')]")
	quitOnErr(err, "Unable to find element expected to wrap all sessions")
	session_markups, err := htmlquery.QueryAll(sessions_container, "//div[contains(@class,'section tile-wrapper')]")
	quitOnErr(err, "Unable to find session div")

	for _, session_markup := range session_markups {
		title, err := htmlquery.Query(session_markup, "//span[contains(@class,'section__header__title')]/text()")
		talks_markup, err := htmlquery.QueryAll(session_markup, "//a[contains(@class,'lumen-tile__link')]")

		var talks []talk
		for _, markup := range talks_markup {
			talks = append(talks, summarizeTalk(markup))
		}
		conf.sessions = append(conf.sessions, session{title.Data, talks, nil})

		if err != nil {
			log.Println("session_title err")
			log.Println(err)
		}

	}

	out <- &conf
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
	fmt.Println(filePath)
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

func downloadConference(conf *conference, dlTarget string) {
	chDownloadTalks := make(chan struct{}, 5)
	talkCount := 0
	for _, sess := range conf.sessions {
		talkCount += len(sess.talks)
	}

	// bootstrap year and month folders
	confLink := conf.link
	yr, mth, ok := getConferenceFolders(confLink)
	if !ok {
		fmt.Println("failed to parse conference link for year and month")
	}
	confPath := fmt.Sprintf("%s/%s/%s", dlTarget, yr, mth)
	err := os.MkdirAll(confPath, os.ModePerm)
	quitOnErr(err, fmt.Sprintf("Failed to make conference directory: \"%s\"", confPath))

	for _, session := range conf.sessions {
		sessPath := fmt.Sprintf("%s/%s", confPath, toFileName(session.title))
		err := os.MkdirAll(sessPath, os.ModePerm)
		quitOnErr(err, fmt.Sprintf("Failed to make session directory: \"%s\"", sessPath))
		for _, talk := range session.talks {
			talkPath := fmt.Sprintf("%s/%s.txt", sessPath, toFileName(talk.title))
			go downloadTalk(talk.link, talkPath, chDownloadTalks)

		}
	}

	for i := 0; i < talkCount; i++ {
		<-chDownloadTalks
	}
}

// TODO:
// Consolidate html library usage
// need a good html element text grabber
// need to parallelize file downloading
// need a function to standardize talk names, remove smart quotes and such

func main() {
	makeDownloadTarget("Making download folder")
	conferences := gatherConferences("Gathering conference links")

	chScoutConferences := make(chan *conference, 100)
	// chDownloadConferences := make(chan *conference, 20)

	defer func() {
		close(chScoutConferences)
	}()

	for _, conf := range conferences {
		conf := conf
		go addConferenceTalks(conf, chScoutConferences)
	}

	conferencesCounter := 0

	homedir, err := os.UserHomeDir()
	quitOnErr(err, "Failed to find home directory")
	downloadTarget := fmt.Sprintf("%s/%s", homedir, conferencesDir)
	for conf := range chScoutConferences {
		numTalks := 0
		for _, session := range conf.sessions {
			numTalks += len(session.talks)
		}
		// fmt.Printf("%s: %d sessions, %d talks\n", conf.title, len(conf.sessions), numTalks)
		downloadConference(conf, downloadTarget)
		conferencesCounter++
		// chDownloadConferences := make(chan *conference, 20)

		allConferencesRead := len(conferences) == conferencesCounter
		if allConferencesRead {
			break
		}
	}

	fmt.Println("Conferences downloaded.")
}
