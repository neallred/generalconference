package main

import (
	"fmt"
	// "io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/antchfx/htmlquery"
)

const language = "eng"
const domain = "https://www.churchofjesuschrist.org"
const index_container_selector = "section .section-wrapper"
const document_directory = "conferences"

var index_link = fmt.Sprintf("%s/general-conference/conferences?lang=%s", domain, language)

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

func quitOnErr(err error, extra_message string) {
	if err != nil {
		if extra_message != "" {
			log.Println(extra_message)
		}
		log.Fatal(err)
	}
}

func makeDownloadTarget(stepLabel string) {
	logStep(stepLabel)

	homedir, err := os.UserHomeDir()
	quitOnErr(err, "Failed to find home directory")
	download_target := fmt.Sprintf("%s/%s", homedir, document_directory)

	err = os.MkdirAll(download_target, os.ModePerm)
	quitOnErr(err, fmt.Sprintf("Failed to make download directory: \"%s\"", download_target))
	fmt.Printf("Made %s\n", download_target)
}

func trim(x string) string {
	return strings.Trim(x, "	\n")
}

func gatherConferences(stepLabel string) []conference {
	logStep(stepLabel)

	var conferences []conference

	doc, err := htmlquery.LoadURL(index_link)
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
			title := trim(a.FirstChild.Data)
			conf := conference{conference_link, title, sessions, nil}
			conferences = append(conferences, conf)
		}
	}

	fmt.Printf("Gathered %d conferences\n", len(conferences))

	return conferences
}

func main() {
	makeDownloadTarget("Making download folder")
	conferences := gatherConferences("Gathering conference links")

	sessionsChan := make(chan *conference)

	defer func() {
		close(sessionsChan)
	}()

	for i, conf := range conferences {

		// start a go routine with the index and url in a closure
		go func(i int, conf conference) {

			// this sends an empty struct into the semaphoreChan which
			// is basically saying add one to the limit, but when the
			// limit has been reached block until there is room
			// semaphoreChan <- struct{}{}

			// send the request and put the response in a result struct
			// along with the index so we can sort them later along with
			// any error that might have occoured

			doc, err := htmlquery.LoadURL(conf.link)
			if err != nil {
				log.Println("err on conference page")
				log.Println(err)
			}

			sessions_container, err := htmlquery.Query(doc, "//div[contains(@class,'section-wrapper')]")
			quitOnErr(err, "Unable to find element expected to wrap all sessions: //div[contains(@class,'section-wrapper')]")
			session_markups, err := htmlquery.QueryAll(sessions_container, "//div[contains(@class,'section tile-wrapper')]")
			quitOnErr(err, "Unable to find expected session div: //div[contains(@class,'section tile-wrapper')]")

			log.Println("session_html greppins:")
			for _, session_markup := range session_markups {
				title, err := htmlquery.Query(session_markup, "//span[contains(@class,'section__header__title')]/text()")
				talks_markup, err := htmlquery.QueryAll(session_markup, "//a[@href]")
				var talks []talk

				for _, talk_markup := range talks_markup {
					var talk_link string
					for _, attr := range talk_markup.Attr {
						if attr.Key == "href" {
							talk_link = slugToUrl(attr.Val)
						}
					}

					author, author_err := htmlquery.Query(talk_markup, "//div[contains(@class, 'lumen-tile__content')]/text()")
					talk_title, title_err := htmlquery.Query(talk_markup, "//div[contains(@class, 'lumen-tile__title')]/div/text()")

					var effective_author string
					if author_err != nil || author == nil {
						effective_author = "unknown"
					} else {
						effective_author = author.Data
					}

					var effective_title string
					if title_err != nil || talk_title == nil {
						effective_title = "unknown"
					} else {
						effective_title = talk_title.Data
					}

					talk := talk{effective_title, effective_author, talk_link, err}
					talks = append(talks, talk)
				}
				conf.sessions = append(conf.sessions, session{title.Data, talks, nil})
				fmt.Println(conf)

				// session_title = title.Data
				log.Println("session_title")
				log.Println(title.Data)
				if err != nil {
					log.Println("session_title err")
					log.Println(err)
				}

			}
			// log.Println(conf.link)

			// now we can send the result struct through the sessionsChan
			sessionsChan <- &conf

			// once we're done it's we read from the semaphoreChan which
			// has the effect of removing one from the limit and allowing
			// another goroutine to start
			// <-semaphoreChan

		}(i, conf)

	}

	// start listening for any results over the resultsChan
	// once we get a result append it to the result slice
	var results []conference
	for {
		result := <-sessionsChan
		results = append(results, *result)
		fmt.Println("result received")
		fmt.Println(*result)

		// if we've reached the expected amount of urls then stop
		// if len(results) == len(urls) {
		// 	break
		// }
	}

	// fmt.Println(conferences)
}
