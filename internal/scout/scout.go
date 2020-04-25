package scout

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb/v3"
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
	fmt.Printf("Locating talks for conferences:\n")
	bar := pb.New(lenConferences).SetWidth(80)
	bar.Start()

	defer close(chScoutConferences)

	for _, conf := range skeletonConferences {
		conf := conf
		go addConferenceTalks(conf, chScoutConferences)
	}

	var conferences []Conference
	for i := 0; i < lenConferences; i++ {
		conf := <-chScoutConferences
		conferences = append(conferences, conf)
		bar.Increment()
	}
	bar.Finish()

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

	return conferences
}

func summarizeTalk(markup *goquery.Selection) Talk {
	href, _ := markup.Attr("href")
	talk_link := slugToUrl(href)
	author := markup.Find("div.lumen-tile__content").Text()
	talk_title := markup.Find("div.lumen-tile__title > div").Text()

	if author == "" {
		fmt.Println("unknown author:", author, markup)
		author = "unknown"
	}

	if talk_title == "" {
		// Hackery because a handful of talks have a slightly different html structure
		talk_title = markup.Find("div.lumen-tile__title").Text()
		if talk_title == "" {
			fmt.Println("unknown title:", talk_title, talk_link)
			talk_title = "unknown"
		}
	}

	return Talk{talk_title, author, talk_link, nil}
}

func addConferenceTalks(conf Conference, out chan<- Conference) {
	resp, err := http.Get(conf.Link)
	quitOnErr(err, "err on conference page")
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	quitOnErr(err, "err on conference page")

	doc.Find("div.section-wrapper div.section.tile-wrapper").Each(func(_ int, session_markup *goquery.Selection) {
		title := session_markup.Find("span.section__header__title").Text()

		var talks []Talk
		session_markup.Find("a.lumen-tile__link").Each(func(_ int, talk *goquery.Selection) {
			talks = append(talks, summarizeTalk(talk))
		})
		conf.Sessions = append(conf.Sessions, Session{strings.TrimSpace(title), talks, nil})

		if err != nil {
			log.Println("session_title err")
			log.Println(err)
		}
	})

	out <- conf
}
