package main

import "github.com/godsic/tidalapi"

var (
	qualityMap = map[string]string{
		tidalapi.Quality[tidalapi.LOSSLESS]: "🆩",
		tidalapi.Quality[tidalapi.MASTER]:   "🆨",
		tidalapi.Quality[tidalapi.HIGH]:     tidalapi.Quality[tidalapi.HIGH],
		tidalapi.Quality[tidalapi.LOW]:      "💩"}
)

func year(date string) string {
	if len(date) == 0 {
		return ""
	}
	return date[0:4]
}
