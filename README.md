# General Conference
This project downloads all available talks (1971 onwards) given in the General Conference of The Church of Jesus Christ from [churchofjesuschrist.org][the_church]. They are saved in plain text. The project may eventually include some text analysis tools, probably using [nltk][nltk]. It may even eventually expose some of the text analysis tools or search functionality via a web interface.


## Running
1. Ensure [go][go] is installed, perhaps via [Go Version Manager][gvm].
1. Run `go run main.go` to download the talks to `$HOME/conferences`. They are bucketed into `<year and month>` and `<session name>`. Download takes a while, depending greatly on network speed. The talks total about 50 megabytes once saved (as of 2019). 

## Tests
TBD

## Downloading talks

[go]: https://golang.org/
[gvm]: https://github.com/moovweb/gvm
[nltk]: https://www.nltk.org/
[the_church]: https://churchofjesuschrist.org
