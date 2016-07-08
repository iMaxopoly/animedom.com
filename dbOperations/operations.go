package main

import (
	"animedom.com/templates"
	"errors"
	"log"

	// RethinkDB driver for Golang
	r "github.com/dancannon/gorethink"
	"net/url"
)

var dbSession *r.Session

func init() {
	/* Open connection to rethinkdb */
	var err error
	dbSession, err = r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "animedom",
	})
	if err != nil {
		log.Fatal(err)
	}
	res, err := r.Table("animes").Count().Run(dbSession)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Close()

	var count float64
	err = res.One(&count)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Loaded database with", count, "animes")
}

func dbFetchXOngoingAnime(skipN int, limitN int) ([]templates.StructureAnime, int, error) {
	resCountQuery, err := r.Table("animes").Filter(map[string]interface{}{
		"Status": "Ongoing",
	}).Count().Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}
	defer resCountQuery.Close()
	if resCountQuery.IsNil() {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int

	err = resCountQuery.One(&resCount)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}

	res, err := r.Table("animes").Filter(map[string]interface{}{
		"Status": "Ongoing",
	}).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}
	defer res.Close()

	if res.IsNil() {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var animelist []templates.StructureAnime
	if err := res.All(&animelist); err != nil {
		return []templates.StructureAnime{}, 0, err
	}

	if len(animelist) == 0 {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}
	return animelist, resCount, nil
}

func dbFetchXTopRatingAnime(skipN int, limitN int) ([]templates.StructureAnime, int, error) {
	resCountQuery, err := r.Table("animes").OrderBy(r.Desc("Score")).Count().Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}
	defer resCountQuery.Close()
	if resCountQuery.IsNil() {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int

	err = resCountQuery.One(&resCount)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}

	res, err := r.Table("animes").OrderBy(r.Desc("Score")).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}
	defer res.Close()

	// Not working: https://github.com/dancannon/gorethink/issues/326
	if res.IsNil() {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var animelist []templates.StructureAnime
	if err := res.All(&animelist); err != nil {
		return []templates.StructureAnime{}, 0, err
	}

	if len(animelist) == 0 {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}
	return animelist, resCount, nil
}

func dbFetchXPopularAnime(value int) ([]templates.StructureAnime, error) {

	resp, err := r.Table("cache_popular_animes").OrderBy(r.Asc("id")).Limit(value).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, err
	}
	defer resp.Close()
	if resp.IsNil() {
		return []templates.StructureAnime{}, errors.New("Empty Result")
	}

	var popularAnimeIDs []map[string]interface{}
	if err = resp.All(&popularAnimeIDs); err != nil {
		return []templates.StructureAnime{}, err
	}

	var animelist []templates.StructureAnime
	for _, v := range popularAnimeIDs {
		res, err := r.Table("animes").Filter(map[string]interface{}{
			"id": v["malID"].(string),
		}).Run(dbSession)
		if err != nil {
			return []templates.StructureAnime{}, err
		}
		if res.IsNil() {
			return []templates.StructureAnime{}, errors.New("Empty Result")
		}
		var anime templates.StructureAnime
		if err = res.One(&anime); err != nil {
			return []templates.StructureAnime{}, err
		}
		res.Close()
		animelist = append(animelist, anime)
	}

	if len(animelist) == 0 {
		return []templates.StructureAnime{}, errors.New("Empty Result")
	}
	return animelist, nil
}

func dbFetchXRecentEpisodes(skipN, limitN int) ([]templates.StructureAnime, int, error) {
	resCountQuery, err := r.Table("cache_recent_animes").OrderBy(r.Asc("id")).Field("malID").Count().Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}
	defer resCountQuery.Close()
	if resCountQuery.IsNil() {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int

	err = resCountQuery.One(&resCount)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}

	resp, err := r.Table("cache_recent_animes").OrderBy(r.Asc("id")).Field("malID").Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}
	defer resp.Close()
	if resp.IsNil() {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}
	var malIDs []string
	if err = resp.All(&malIDs); err != nil {
		return []templates.StructureAnime{}, 0, err
	}

	var animelist []templates.StructureAnime

	for i := 0; i < len(malIDs); i++ {
		res, err := r.Table("animes").Filter(map[string]interface{}{
			"id": malIDs[i],
		}).Run(dbSession)
		if err != nil {
			return []templates.StructureAnime{}, 0, err
		}

		if res.IsNil() {
			return []templates.StructureAnime{}, 0, errors.New("Empty Result")
		}
		var tempAnimelist templates.StructureAnime
		if err = res.One(&tempAnimelist); err != nil {
			return []templates.StructureAnime{}, 0, err
		}
		res.Close()
		animelist = append(animelist, tempAnimelist)
	}
	if len(animelist) == 0 {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}
	return animelist, resCount, nil
}

func dbCreateSearchJSON(term string) ([]structureSearchJSON, error) {
	//r.db("animedom").table("animes").filter(function(doc){
	//	return doc('AnimeShowName').match("fate")
	//}).pluck("AnimeShowName", "id").orderBy("AnimeShowName")
	res, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("AnimeShowName").Match(term)
	}).Pluck("AnimeShowName", "id").OrderBy("AnimeShowName").Run(dbSession)
	if err != nil {
		return []structureSearchJSON{}, err
	}
	defer res.Close()

	if res.IsNil() {
		return []structureSearchJSON{}, errors.New("Empty Result")
	}

	var mappedData []structureSearchJSON

	err = res.All(&mappedData)
	if err != nil {
		return []structureSearchJSON{}, err
	}

	for i, v := range mappedData {
		mappedData[i].MALID = "http://animedom.com/details/" + url.QueryEscape(v.AnimeShowName)
	}

	return mappedData, nil
}

func dbXFetchAnimesByGenre(queries [][]byte, skipN int, limitN int) ([]templates.StructureAnime, int, error) {
	//  r.db("animedom").table('animes').filter(function (anime){
	//      return anime.getField("Genre").contains('Action', 'Fantasy', 'Shoujo');
	//  }).getField('id')
	//

	genres := make([]interface{}, len(queries))
	for i, v := range queries {
		genres[i] = string(v)
	}

	resCountQuery, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("Genre").Contains(genres...)
	}).Count().Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}
	defer resCountQuery.Close()

	if resCountQuery.IsNil() {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int

	err = resCountQuery.One(&resCount)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}

	res, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("Genre").Contains(genres...)
	}).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}
	defer res.Close()

	if res.IsNil() {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var animelist []templates.StructureAnime

	err = res.All(&animelist)
	if err != nil {
		return []templates.StructureAnime{}, 0, err
	}

	if len(animelist) == 0 {
		return []templates.StructureAnime{}, 0, errors.New("Empty Result")
	}

	return animelist, resCount, nil
}

// Monitor Popular Animes
func dbCheckExistsAnimesByID(id string) error {
	res, err := r.Table("animes").Filter(map[string]interface{}{
		"id": id,
	}).Run(dbSession)
	if err != nil {
		return err
	}
	defer res.Close()
	if res.IsNil() {
		return errors.New("Empty Result")
	}
	return nil
}

/* Explicitly defined recent and popular pushing functions separately to avoid confusion and mistake */
func dbPushPopularAnimes(index int, malID string) {
	_, err := r.Table("popular_animes").Insert(map[string]interface{}{
		"id":    index,
		"malID": malID,
	}).RunWrite(dbSession)
	if err != nil {
		log.Println(err)
		return
	}
}

func dbPushRecentAnimes(index int, malID string) {
	_, err := r.Table("recent_animes").Insert(map[string]interface{}{
		"id":    index,
		"malID": malID,
	}).RunWrite(dbSession)
	if err != nil {
		log.Println(err)
		return
	}
}

func dbTruncateTable(tableName string) error {
	_, err := r.Table(tableName).Delete().RunWrite(dbSession)
	if err != nil {
		return err
	}
	return nil
}

// Monitor Recent Animes

func dbCheckExistsAnimesByName(animeName string) error {
	res, err := r.Table("animes").Filter(map[string]interface{}{
		"AnimeShowName": animeName,
	}).Run(dbSession)
	if err != nil {
		return err
	}
	defer res.Close()
	if res.IsNil() {
		return errors.New("Empty Result")
	}
	return nil
}

func dbInsertNewAnime(anime templates.StructureAnime) error {
	err := r.DB("animedom").Table("animes").Insert(anime).Exec(dbSession)
	if err != nil {
		return err
	}
	return nil
}

func dbFetchAnimeByName(animeName string) (templates.StructureAnime, error) {
	res, err := r.Table("animes").Filter(map[string]interface{}{
		"AnimeShowName": animeName,
	}).Run(dbSession)
	if err != nil {
		return templates.StructureAnime{}, err
	}
	defer res.Close()
	if res.IsNil() {
		return templates.StructureAnime{}, errors.New("Empty Result")
	}
	var anime templates.StructureAnime
	if err = res.One(&anime); err != nil {
		return templates.StructureAnime{}, err
	}
	return anime, nil
}

func dbUpdateEpisodelist(episodes []templates.StructureEpisode, animeName string) error {
	_, err := r.Table("animes").Filter(map[string]interface{}{
		"AnimeShowName": animeName,
	}).Update(map[string]interface{}{
		"EpisodeList": episodes,
	}).RunWrite(dbSession)
	if err != nil {
		return err
	}
	return nil
}

func dbCopyTableTo(from, to string) error {
	_, err := r.Table(to).Delete().RunWrite(dbSession)
	if err != nil {
		return err
	}
	_, err = r.Table(to).Insert(r.Table(from)).RunWrite(dbSession)
	if err != nil {
		return err
	}
	return nil
}
