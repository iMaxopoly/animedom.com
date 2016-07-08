package myanimelistAssistant

import (
	"strings"

	"animedom.com/dbOperations"

	"github.com/PuerkitoBio/goquery"
)

/*
 MonitorPopularAnimes() is the function where we automatically sync the popular animes from myanimelist.net sortd in
 order of popularity.
*/
func MonitorPopularAnimes() {
	doc, err := goquery.NewDocument("http://myanimelist.net/anime/season")
	if err != nil {
		panic(err)
	}
	animelist := doc.Find(".link-image")

	// Truncate Table
	err = dbOperations.DBTruncateTable("popular_animes")
	if err != nil {
		panic(err)
	}

	// Using count to index(for sorting) the IDs in animedom.popular_animes.
	count := 0
	for node := range animelist.Nodes {
		res, exists := animelist.Eq(node).Attr("href")
		if !exists {
			continue
		}

		// Parsing the URL to get MAL ID.
		res = strings.Replace(res, "http://myanimelist.net/anime/", "", 1)
		for i, v := range res {
			if v == '/' {
				res = res[0:i]
				break
			}
		}

		// If it exists in our database we check if it exists in animedom.animes table.
		if err := dbOperations.DBCheckExistsAnimesByID(res); err != nil {
			if err.Error() != "Empty Result" {
				// There was some unexpected error
				panic(err)
			}
			continue
		}
		// Insert it to animedom.popular_animes.
		dbOperations.DBPushPopularAnimes(count, res)
		count++
	}
	err = dbOperations.DBCopyTableTo("popular_animes", "cache_popular_animes")
	if err != nil {
		panic(err)
	}
}
