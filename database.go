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
	res, err := r.Table("animes").Filter(map[string]string{
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
	if res.All(&animelist); err != nil {
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
	if res.All(&animelist); err != nil {
		return []templates.StructureAnime{}, err
	}

	return animelist, nil
}

func dbFetchXPopularAnime(value int) ([]templates.StructureAnime, error) {
	res, err := r.Table("popular_animes").OrderBy(r.Asc("index")).Limit(value).Run(dbSession)
	if err != nil {
		return []templates.StructureAnime{}, err
	}
	if res.IsNil() {
		return []templates.StructureAnime{}, errors.New("Empty Result")
	}

	var popularAnimeIDs []map[string]interface{}
	if res.All(&popularAnimeIDs); err != nil {
		return []templates.StructureAnime{}, err
	}
	res.Close()

	var animelist []templates.StructureAnime
	for _, v := range popularAnimeIDs {
		res, err := r.Table("animes").Filter(map[string]string{
			"id":v["malID"].(string),
		}).Run(dbSession)
		if err != nil {
			return []templates.StructureAnime{}, err
		}
		if res.IsNil() {
			return []templates.StructureAnime{}, errors.New("Empty Result")
		}
		var anime templates.StructureAnime
		if res.One(&anime); err != nil {
			return []templates.StructureAnime{}, err
		}
		res.Close()
		animelist = append(animelist, anime)
	}

	return animelist, nil
}


// Monitor Popular Animes
func dbCheckExistsAnimesByID(id string) (error) {
	res, err := r.Table("animes").Filter(map[string]string{
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

func dbPushPopularAnimes(index int, malID string) {
	_, err := r.Table("popular_animes").Insert(map[string]interface{}{
		"index" : index,
		"malID":      malID,
	}).RunWrite(dbSession)
	if err != nil {
		log.Println(err)
		return
	}
}

func dbDropTable(tableName string) (error) {
	_, err := r.TableDrop(tableName).Run(dbSession)
	if err != nil {
		return err
	}
	return nil
}

func dbCreateTable(tableName string) (error) {
	_, err := r.TableCreate(tableName).RunWrite(dbSession)
	if err != nil {
		return err
	}
	return nil
}

// Monitor Recent Animes

func dbCheckExistsAnimesByName(animeName string)(error){
	res, err := r.Table("animes").Filter(map[string]string{
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

func dbInsertNewAnime(anime templates.StructureAnime)(error){
	err := r.DB("animedom").Table("animes").Insert(anime).Exec(dbSession)
	if err != nil {
		return err
	}
	return nil
}

func dbGetEpisodeCount(animeName string)(int, error){
	res, err := r.Table("animes").Filter(map[string]string{
		"AnimeShowName":animeName,
	}).Field("EpisodeList").Run(dbSession)
	if err != nil {
		return 0, err
	}
	defer res.Close()
	if res.IsNil() {
		return 0, errors.New("Empty Result")
	}
	var episodes structureEpisode
	if res.One(&episodes); err != nil {
		return 0, err
	}
	return len(episodes), nil
}