package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
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

func trim(x string) string {
	return strings.Trim(x, "	\n")
}

func gatherConferences(stepLabel string) []conference {
	logStep(stepLabel)

	var conferences []conference

	doc, err := htmlquery.LoadURL(indexLink)

	quitOnErr(err, "Unable to load General Conferences index page")
	conference_links, err := htmlquery.QueryAll(doc, "//a")
	quitOnErr(err, "Failed to find conference links on index page")

	for _, a := range conference_links {
		var href string
		var class string
		for _, attr := range a.Attr {
			if attr.Key == "class" {
				class = attr.Val
			}
			if attr.Key == "href" {
				href = attr.Val
			}
		}
		if class == "year-line__link" && href != "" {
			var sessions []session
			conference_link := slugToUrl(href)
			title := strings.TrimSpace(a.FirstChild.Data)
			conf := conference{conference_link, title, sessions, nil}
			conferences = append(conferences, conf)
		}
	}

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

func downloadConference(conf *conference, dlTarget string) {
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
		sessPath := fmt.Sprintf("%s/%s", confPath, session.title)
		err := os.MkdirAll(sessPath, os.ModePerm)
		quitOnErr(err, fmt.Sprintf("Failed to make session directory: \"%s\"", sessPath))
		for _, talk := range session.talks {
			talkPath := fmt.Sprintf("%s/%s", sessPath, talk.title)
			fmt.Println(talk.title)

			doc, err := htmlquery.LoadURL(talk.link)
			quitOnErr(err, "Unable to load talk", talk.link)
			goqueryDoc := goquery.NewDocumentFromNode(doc)
			article := goqueryDoc.Find("article")
			articleText := article.Text()
			ioutil.WriteFile(talkPath, []byte(articleText), 0644)
		}
	}
}

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
