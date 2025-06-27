# Search Logger
A deduplicated search logging function built in Go, designed to capture the **final search term** a user enters during progressive searches (e.g. `B`, `Bu`, `Bus`, … → `Business`), and store only the **most complete** form of the query.

---

The solution is in `service/search_log_service.go`, in function LogSearch().

The strategy involves utilizing a debounce mechanism to ensure that only the final search term is logged after a user has stopped typing for a specified period. This prevents logging intermediate search terms and reduces noise in the logs.
It uses a cache to store the last search term for each user, as well as time of that search. 

To run tests,
```bash
make test
```

# Config
## LOG_SEARCH_DEBOUNCE_DELAY_SECONDS
The debounce delay in seconds for the search logger. This is the time period during which if a user types a new character, the previous search term will be discarded and the new one will be logged after the delay.

# AI
I did not use AI for the general solution, but I did use it for writing tests.
