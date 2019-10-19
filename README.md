# General Conference
This project downloads all available talks (1971 onwards) given in the General Conference of The Church of Jesus Christ from [churchofjesuschrist.org][the_church]. They are saved in plain text. The project may eventually include some text analysis tools, probably using [nltk][nltk]. It may even eventually expose some of the text analysis tools or search functionality via a web interface.


## Setup
1. Ensure [pip][pip] is installed.
1. Install [pyenv][pyenv], maybe using [pyenv-installer][pyenv-installer].
1. Set up a virtual environment in the repo's root with `pyenv virtualenv 3.8.0 general_conference`. Make it activate for the repo by default with `pyenv local general_conference`. 
1. Install [poetry][poetry] with `pip install poetry`.
1. Install the project dependencies with `poetry install`


## Tests
After setup, run tests with `pytest` or `python -m pytest`


## Downloading talks
After setup, run `python general_conference/main.py`. Download takes a while, depending greatly on network speed. The talks total about 50 megabytes once saved (as of 2019). Downloaded talks are located in `<project root>/conferences`. They are bucketed into `<year and month>` and `<session name>`.

[nltk]: https://www.nltk.org/
[pip]: https://pip.pypa.io/en/stable/installing/
[poetry]: https://poetry.eustace.io/
[pyenv-installer]: https://github.com/pyenv/pyenv-installer
[pyenv]: https://github.com/pyenv/pyenv 
[the_church]: https://churchofjesuschrist.org
