# proxkeep
Proxy servers availability tracker. Public servers have a short lifetime span. Keeping up to date functional servers might be important part of business. Proxkeep uses GoLang lightweight goroutines to manage multiple, concurrent checks at low server memory / cpu resources.

```bash
proxkeep run
```
Runs proxy servers check. Progress stats are available via http on port :8000