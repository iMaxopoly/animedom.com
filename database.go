package main

import (
	"log"
	"errors"
	"animedom.com/templates"

	// RethinkDB driver for Golang
	r "github.com/dancannon/gorethink"
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

func dbFetchXOngoingAnime(skipN int, limitN int) ([]templates.StructureAnime, error) {
	res, err := r.Table("animes").Filter(map[string]interface{}{
		"Status":"Ongoing",
	}).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, err
	}
	defer res.Close()
	if res.IsNil() {
		return []templates.StructureAnime{}, errors.New("Empty Result")
	}

	var animelist []templates.StructureAnime
	if err = res.All(&animelist); err != nil {
		return []templates.StructureAnime{}, err
	}

	return animelist, nil
}

func dbFetchXTopRatingAnime(skipN int, limitN int) ([]templates.StructureAnime, error) {
	res, err := r.Table("animes").OrderBy(r.Desc("Score")).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, err
	}
	defer res.Close()
	if res.IsNil() {
		return []templates.StructureAnime{}, errors.New("Empty Result")
	}

	var animelist []templates.StructureAnime
	if err = res.All(&animelist); err != nil {
		return []templates.StructureAnime{}, err
	}

	return animelist, nil
}

func dbFetchXPopularAnime(value int) ([]templates.StructureAnime, error) {
	res, err := r.Table("cache_popular_animes").OrderBy(r.Asc("id")).Limit(value).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, err
	}
	if res.IsNil() {
		return []templates.StructureAnime{}, errors.New("Empty Result")
	}

	var popularAnimeIDs []map[string]interface{}
	if err = res.All(&popularAnimeIDs); err != nil {
		return []templates.StructureAnime{}, err
	}
	res.Close()

	var animelist []templates.StructureAnime
	for _, v := range popularAnimeIDs {
		res, err := r.Table("animes").Filter(map[string]interface{}{
			"id":v["malID"].(string),
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

	return animelist, nil
}

func dbFetchXRecentAnime(skipN, limitN int) ([]templates.StructureAnime, error) {
	res, err := r.Table("cache_recent_animes").OrderBy(r.Asc("id")).Field("malID").Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, err
	}
	if res.IsNil() {
		return []templates.StructureAnime{}, errors.New("Empty Result")
	}
	var malIDs []string
	if err = res.All(&malIDs); err != nil {
		return []templates.StructureAnime{}, err
	}

	var animelist []templates.StructureAnime

	for i := 0; i < len(malIDs); i++ {
		res, err := r.Table("animes").Filter(map[string]interface{}{
			"id" : malIDs[i],
		}).Run(dbSession)
		if err != nil {
			return []templates.StructureAnime{}, err
		}
		if res.IsNil() {
			return []templates.StructureAnime{}, errors.New("Empty Result")
		}
		var tempAnimelist templates.StructureAnime
		if err = res.One(&tempAnimelist); err != nil {
			return []templates.StructureAnime{}, err
		}
		animelist = append(animelist, tempAnimelist)
	}

	return animelist, nil
}

// Monitor Popular Animes
func dbCheckExistsAnimesByID(id string) (error) {
	res, err := r.Table("animes").Filter(map[string]interface{}{
		"id":id,
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
		"id" : index,
		"malID":      malID,
	}).RunWrite(dbSession)
	if err != nil {
		log.Println(err)
		return
	}
}

func dbPushRecentAnimes(index int, malID string) {
	_, err := r.Table("recent_animes").Insert(map[string]interface{}{
		"id" : index,
		"malID":      malID,
	}).RunWrite(dbSession)
	if err != nil {
		log.Println(err)
		return
	}
}

func dbTruncateTable(tableName string) (error) {
	_, err := r.Table(tableName).Delete().RunWrite(dbSession)
	if err != nil {
		return err
	}
	return nil
}

// Monitor Recent Animes

func dbCheckExistsAnimesByName(animeName string) (error) {
	res, err := r.Table("animes").Filter(map[string]interface{}{
		"AnimeShowName":animeName,
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

func dbInsertNewAnime(anime templates.StructureAnime) (error) {
	err := r.DB("animedom").Table("animes").Insert(anime).Exec(dbSession)
	if err != nil {
		return err
	}
	return nil
}

func dbFetchAnimeByName(animeName string) (templates.StructureAnime, error) {
	res, err := r.Table("animes").Filter(map[string]interface{}{
		"AnimeShowName":animeName,
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

func dbUpdateEpisodelist(episodes []templates.StructureEpisode, animeName string) (error) {
	_, err := r.Table("animes").Filter(map[string]interface{}{
		"AnimeShowName":animeName,
	}).Update(map[string]interface{}{
		"EpisodeList":episodes,
	}).RunWrite(dbSession)
	if err != nil {
		return err
	}
	return nil
}

func dbCopyTableTo(from, to string) (error) {
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