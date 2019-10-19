import os
import json
import pprint
import re

pp = pprint.PrettyPrinter(indent=4)

import requests
from bs4 import BeautifulSoup

own_path = os.path.dirname(os.path.abspath(__file__))

LANGUAGE="eng"
DOMAIN = "https://www.churchofjesuschrist.org"
INDEX_LINK = f"{DOMAIN}/general-conference/conferences?lang={LANGUAGE}"
INDEX_CONTAINER_SELECTOR = "section .section-wrapper"
INDEX_LINK_SELECTOR = "a:not([href=\"\"])"
HTML_PARSER = 'html.parser'

DOCUMENT_DIRECTORY = "../conferences"

def get_page_soup(link):
    index_page = requests.get(link)
    index_page.encoding = 'utf-8'
    index_document = index_page.text
    soup = BeautifulSoup(index_document, HTML_PARSER)
    return soup

def get_href(x):
    return x.get("href")

def denest(l):
    return [item for sublist in l for item in sublist]

def get_conference_links():
    soup = get_page_soup(INDEX_LINK)
    container = soup.select(f"{INDEX_CONTAINER_SELECTOR}")[0].select(f"{INDEX_LINK_SELECTOR}")
    session_links = map(get_href, container)
    return session_links

# end of url is of form /1971/04?lang=eng
YEAR_LENGTH = 7
LANGUAGE_SLUG_LENGTH = 9
def get_conference_dir(conference_link):
    begin = (0 - YEAR_LENGTH - LANGUAGE_SLUG_LENGTH)
    end = (0-LANGUAGE_SLUG_LENGTH)
    return conference_link[begin:end].replace('/', '_')

def ensure_conference_dirs(conference_links):
    for folder in map(get_conference_dir, conference_links):
        os.makedirs(f"{own_path}/{DOCUMENT_DIRECTORY}/{folder}", exist_ok=True)

def get_conference(conference_link, conference_dir):
    soup = get_page_soup(f"{DOMAIN}{conference_link}")

    print(conference_dir)

    sessions_soup = soup.select("section .section-wrapper")[0].select(".tile-wrapper")
    conference_talks = [get_session_talks(session_soup, conference_dir) for session_soup in sessions_soup]

    return denest(conference_talks)


def get_talk(talk_soup):
    return {
        "title": talk_soup.find("div", class_="lumen-tile__title").get_text().strip(),
        "link": talk_soup.find("a").get("href"),
        "author": talk_soup.find("div", class_="lumen-tile__content").get_text(),
    };

def get_session_talks(session_soup, conference_dir):
    session_title = session_soup.find("span", class_="section__header__title").get_text()

    os.makedirs(f"{own_path}/{DOCUMENT_DIRECTORY}/{conference_dir}/{session_title}", exist_ok=True)

    talks_soup = session_soup.select('.lumen-tile')
    talks = list(map(get_talk, talks_soup))
    for talk in talks:
        talk["session_title"] = session_title
        talk["conference_dir"] = conference_dir

    return talks

def write_talks_list(talks_list):
    os.makedirs(f"{own_path}/../tmp", exist_ok=True)
    with open(f"{own_path}/../tmp/talks-list.json", "w+") as f:
        f.write(json.dumps(talks_list))


# These special characters appear in talk titles:
# ["!", "&", "(", ")", "*", ",", "-", ".", ":", ";", "?", "[", "]", " ", "—", "’", "“", "”", "…" ]
# The thing that looks like a space is "\x0a", a non breaking space
allowed_filename_characters = re.compile("[^A-Za-z0-9 ,-]")
def write_talk_text(talk, failed_talks):
    try:
        link = f"{DOMAIN}{talk['link']}"
        sanitized_file_name = re.sub(allowed_filename_characters, "", talk["title"])
        file_name = f"{own_path}/{DOCUMENT_DIRECTORY}/{talk['conference_dir']}/{talk['session_title']}/{sanitized_file_name}.txt"
        talk_text = get_page_soup(link).find('article').get_text(separator="\n\n")

        print(f"writing {file_name}")
        with open(file_name, "w+") as f:
            f.write(talk_text)
    except Exception as e:
        failed_talks.append(talk)
        print(f"failed to save {file_name} because: {e}")


def main():
    conference_links = get_conference_links()
    conf_links = list(conference_links)
    conference_links_and_dirs = [(x, get_conference_dir(x)) for x in conf_links]
    ensure_conference_dirs(conf_links)
    conference_talks = denest([get_conference(link, dir_) for (link, dir_) in conference_links_and_dirs])
    write_talks_list(conference_talks)
    failed_talks = []
    # TODO: have a "retry with backoff" helper
    # TODO: Parallelize the fetching of talks
    for talk in conference_talks:
        write_talk_text(talk, failed_talks)

    print("failed to fetch these talks:")
    pp.pprint(failed_talks)

    double_failed_talks = []
    print("trying refetch now:")
    for talk in failed_talks:
        write_talk_text(talk, double_failed_talks)

    print("failed to refetch these talks:")
    pp.pprint(double_failed_talks)

if __name__ == '__main__':
    main()


# example_talk = {
#     'author': 'Joseph Fielding Smith',
#     'conference_dir': '1971_10',
#     'link': '/study/general-conference/1971/10/let-the-spirit-of-oneness-prevail?lang=eng',
#     'session_title': 'Sunday Afternoon Session',
#     'title': 'Let the Spirit of Oneness Prevail'
# }

# There are 60 video presentations as for 1971-2019.
# These URLs are of the form
# "/general-conference/1972/04/media/1789534104001?lang=eng"
# There is no text transcript for them.
# Let them fail and retry rather than
# trying to build brittle logic for filtering them out.
