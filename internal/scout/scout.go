package scout

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type Conference struct {
	Link     string
	Title    string
	Sessions []Session
	err      error
}

type Session struct {
	Title string
	Talks []Talk
	err   error
}

type Talk struct {
	Title  string
	Author string
	Link   string
	err    error
}

const language = "eng"

const domain = "https://www.churchofjesuschrist.org"

var indexLink = fmt.Sprintf("%s/general-conference/conferences?lang=%s", domain, language)

func quitOnErr(err error, msgs ...string) {
	if err != nil {
		for _, msg := range msgs {
			log.Println(msg)
		}
		log.Fatal(err)
	}
}

func slugToUrl(slug string) string {
	return fmt.Sprintf("%s%s", domain, slug)
}

func logStep(x string) {
	fmt.Println(x)
}

func Gather(stepLabel string) []Conference {
	logStep(stepLabel)
	skeletonConferences := getConferences()
	lenConferences := len(skeletonConferences)

	chScoutConferences := make(chan Conference, lenConferences)

	defer func() {
		close(chScoutConferences)
	}()

	for _, conf := range skeletonConferences {
		conf := conf
		go addConferenceTalks(conf, chScoutConferences)
	}

	var conferences []Conference
	for i := 0; i < lenConferences; i++ {
		fmt.Println("scout conf", i)
		conf := <-chScoutConferences
		conferences = append(conferences, conf)
	}

	return conferences
}

func getConferences() []Conference {

	var conferences []Conference

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
			var sessions []Session
			conf := Conference{link, title, sessions, nil}
			conferences = append(conferences, conf)
		}
	})

	fmt.Printf("Found %d conferences \n", len(conferences))
	return conferences
}

func summarizeTalk(markup *html.Node) Talk {
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

	return Talk{effective_title, effective_author, talk_link, author_err}
}

func addConferenceTalks(conf Conference, out chan<- Conference) {
	doc, err := htmlquery.LoadURL(conf.Link)
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

		var talks []Talk
		for _, markup := range talks_markup {
			talks = append(talks, summarizeTalk(markup))
		}
		conf.Sessions = append(conf.Sessions, Session{title.Data, talks, nil})

		if err != nil {
			log.Println("session_title err")
			log.Println(err)
		}

	}

	out <- conf
}
