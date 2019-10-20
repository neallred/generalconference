import itertools
import datetime
now = datetime.datetime.now()

presidents = {
    "Joseph Smith": {
        "birth": (1805, 12, 23),
        "ordained": (1832, 1, 25), # (1830, 4, 6) was first elder
        "end": (1844, 6, 27),
    },
    "Brigham Young": {
        "birth": (1801, 6, 1),
        "ordained": (1847, 12, 27),
        "end": (1877, 8, 29),
    },
    "John Taylor": {
        "birth": (1808, 11, 1),
        "ordained": (1880, 10, 10),
        "end": (1887, 7, 25),
    },
    "Wilford Woodruff": {
        "birth": (1807, 3, 1),
        "ordained": (1889, 4, 7),
        "end": (1898, 9, 2),
    },
    "Lorenzo Snow": {
        "birth": (1814, 4, 3),
        "ordained": (1898, 9, 13),
        "end": (1901, 10, 10),
    },
    "Joseph F. Smith": {
        "birth": (1838, 11, 13),
        "ordained": (1901, 10, 17),
        "end": (1918, 11, 19),
    },
    "Heber J. Grant": {
        "birth": (1856, 11, 22),
        "ordained": (1918, 11, 23),
        "end": (1945, 5, 14),
    },
    "George Albert Smith": {
        "birth": (1870, 4, 4),
        "ordained": (1945, 5, 21),
        "end": (1951, 4, 4),
    },
    "David O. McKay": {
        "birth": (1873, 9, 8),
        "ordained": (1951, 4, 9),
        "end": (1970, 1, 18),
    },
    "Joseph Fielding Smith": {
        "birth": (1876, 7, 19),
        "ordained": (1970, 1, 23),
        "end": (1972, 7, 2),
    },
    "Harold B. Lee": {
        "birth": (1899, 3, 28),
        "ordained": (1972, 7, 7),
        "end": (1973, 12, 26),
    },
    "Spencer W. Kimball": {
        "birth": (1895, 3, 28),
        "ordained": (1973, 12, 30),
        "end": (1985, 11, 5),
    },
    "Ezra Taft Benson": {
        "birth": (1899, 8, 4),
        "ordained": (1985, 11, 10),
        "end": (1994, 5, 30),
    },
    "Howard W. Hunter": {
        "birth": (1907, 11, 14),
        "ordained": (1994, 6, 5),
        "end": (1995, 3, 3),
    },
    "Gordon B. Hinckley": {
        "birth": (1910, 6, 23),
        "ordained": (1995, 3, 12),
        "end": (2008, 1, 27),
    },
    "Thomas S. Monson": {
        "birth": (1927, 8, 21),
        "ordained": (2008, 2, 3),
        "end": (2018, 1, 2),
    },
    "Russell M. Nelson": {
        "birth": (1924, 9, 9),
        "ordained": (2018, 1, 14),
        "end": None,
    },
}

conference_year_tuples = list(itertools.product(range(1971,2020), [4,10]))

def get_president_by_conference(year_month):
    padded_month = f"{year_month[1]}".rjust(2, "0")
    year_month_string = f"{year_month[0]}{padded_month}"
    for k, v in presidents.items():
        ordained = v["ordained"][:2]
        end = v["end"][:2] if v["end"] else (now.year, now.month)
        if ordained <= year_month and end >= year_month:
            return (k, year_month_string)
    return ("N/A", year_month_string)

conferences_by_president = [get_president_by_conference(year_month) for year_month in conference_year_tuples]


