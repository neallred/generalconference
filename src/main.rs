// extern crate tokio;
// use std::collections::HashMap;
// use futures::{stream, Future, Stream};

use tokio::fs;
use tokio::stream;
use tokio::prelude::*;
use scraper::{Html, Selector};
// use reqwest;
// use tokio::io;
// use tokio::net::TcpStream;
// use tokio::prelude::*;
//
// async fn main() -> Result<(), Box<dyn std::error::Error>> {
// async fn main() -> Result<(), dyn std::error::Error> {
// Result<(), Box<dyn std::error::Error + 'static>>
// const DOMAIN: &str  = "https://www.churchofjesuschrist.org";
// const INDEX_LINK: &str = concat!(DOMAIN, "/general-conference/conferences");
// const INDEX_LINK = DOMAIN String::from(format!("{}/general-conference/conferences?lang={}", DOMAIN, LANGUAGE));

const PARALLEL_REQUESTS: usize = 2;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let domain: &str = "https://www.churchofjesuschrist.org";
    let language: &str = "eng";
    let index_link = String::from(format!("{}/general-conference/conferences?lang={}", domain, language));
    let sel_index_container = Selector::parse("section .section-wrapper").unwrap();
    let sel_index_link = Selector::parse("a:not([href=\"\"])").unwrap();
    let sel_conference_sessions = Selector::parse("section .section-wrapper").unwrap();
    let sel_conference_session = Selector::parse(".tile-wrapper").unwrap();
    let sel_session_title = Selector::parse("span.section__header__title").unwrap();
    let http_client = reqwest::Client::new();


    let bob = fs::write("foo.txt", b"Hello world!\n").await?;

    let res = http_client
        .get(&index_link)
        .send()
        .await?;

    let body = res.text().await?;
    let document = Html::parse_document(&body);

    let index_container = document.select(&sel_index_container).next().unwrap();

    let index_links: scraper::element_ref::Select = index_container.select(&sel_index_link);
    let conferences_stream = stream::iter(index_links);
    let conferences_stream_2 = conferences_stream 
        .map(move |link| {
            let link_href = link.value().attr("href");
            // Kind of the right direction
            // tokio::future::ok(link_href) 

            if let Some(slug) = link_href {
                let second_request = http_client
                    .get(&format!("{}/{}", domain, slug))
                    .send();
                    second_request
            } else {
                http_client
                    .get("haskell.org")
                    .send()
            }
        })
        .buffer_unordered(PARALLEL_REQUESTS);

    tokio::spawn(conferences_stream_2);
    // tokio::run(conferences_stream_2);
//        .into_future()

    //     let client = Client::new();

    // let urls = vec!["https://api.ipify.org", "https://api.ipify.org"];

    // let bodies = stream::iter_ok(urls)
    //     .map(move |url| {
    //         http_client
    //             .get(url)
    //             .send()
    //             .and_then(|res| res.into_body().concat2().from_err())
    //     })
    //     .buffer_unordered(PARALLEL_REQUESTS);

    // let work = bodies
    //     .for_each(|b| {
    //         println!("Got {} bytes", b.len());
    //         Ok(())
    //     })
    //     .map_err(|e| panic!("Error while processing: {}", e));

    // tokio::run(work);

    // TODO:
    // right here when we get nearly 100 separate conferences is a great time to start
    // parallelizing data fetches.
    // for link in index_container.select(&sel_index_link) {

    //     let link_text = link.text().collect::<Vec<_>>();
    //     let link_href = link.value().attr("href");
    //     println!("session_link:\n{:?}", link_text );

    //     if let Some(slug) = link_href {
    //         let session_response = http_client
    //             .get(&format!("{}/{}", domain, slug))
    //             .send()
    //             .await?;

    //         let session_page_text = session_response.text().await?;
    //         let session_page = Html::parse_document(&session_page_text);

    //         let session_tiles = session_page
    //             .select(&sel_conference_sessions).next().unwrap()
    //             .select(&sel_conference_session);

    //         for tile in session_tiles {

    //             let session_title = tile
    //                 .select(&sel_session_title)
    //                 .next()
    //                 .unwrap()
    //                 .text()
    //                 .collect::<Vec<_>>();

    //             println!("session_title: {:?}", session_title);

    //             // sessions_soup = soup.select("section .section-wrapper")[0].select(".tile-wrapper")
    //         }
    //     }
    // }

    // let resp: HashMap<String, String> = reqwest::get("https://httpbin.org/ip")
    // .await?
    // .json()
    // .await?;
    // let bob = fs::write("foo.txt", b"Hello world!");
    // tokio::spawn(bob )
    Ok(())
}

// const PARALLEL_REQUESTS: usize = 2;
// helpful resources
// How can I perform parallel asynchronous HTTP GET requests with reqwest?
// https://stackoverflow.com/questions/51044467/how-can-i-perform-parallel-asynchronous-http-get-requests-with-reqwest
//
// Join futures with limited concurrency
//
// How to merge iterator of streams?
//
// How do I synchronously return a value calculated in an asynchronous Future in stable Rust?
//
