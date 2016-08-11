package dbOperations

import (
	"errors"
	"log"

	"animedom.com/common"
	"animedom.com/projectModels"

	// RethinkDB driver for Golang
	"strings"

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
	common.CheckErrorAndPanic(err)

	res, err := r.Table("animes").Count().Run(dbSession)
	if err != nil {
		log.Fatal(err)
	}

	var count float64
	err = res.One(&count)
	common.CheckErrorAndPanic(err)

	err = res.Close()
	common.CheckErrorAndPanic(err)

	log.Println("Loaded database with", count, "animes")
	//DBInsertHash()
}

//func DBInsertHash() {
//	res, err := r.Table("animes").Run(dbSession)
//	common.CheckErrorAndPanic(err)
//
//	var animes []projectModels.StructureAnime
//	err = res.All(&animes)
//	common.CheckErrorAndPanic(err)
//
//	for _, anime := range animes {
//		anime.WikiHash = common.GetFNV("wiki" +
//			anime.MALID +
//			anime.Slug +
//			common.HashWiki)
//		for epIndex := 0; epIndex < len(anime.EpisodeList); epIndex++ {
//			for mirIndex := 0; mirIndex < len(anime.EpisodeList[epIndex].Mirrors); mirIndex++ {
//				anime.EpisodeList[epIndex].Mirrors[mirIndex].MirrorHash =
//					common.GetFNV(fmt.Sprintf("mirror%s%s%d%d%s",
//						anime.MALID,
//						anime.Slug,
//						epIndex,
//						mirIndex,
//						common.HashMirror))
//			}
//		}
//
//		err = DBModifyAnime(anime, anime.MALID)
//		common.CheckErrorAndPanic(err)
//	}
//}

// DBInsertNewBlog inserts a new blog post from the blogs folder into database
func DBInsertNewBlog(blog projectModels.StructureBlog) error {
	err := r.DB("animedom").Table("blogs").Insert(blog).Exec(dbSession)
	return err
}

// DBModifyBlog updates a given blog in the database
func DBModifyBlog(blog projectModels.StructureBlog) error {
	_, err := r.Table("blogs").Filter(map[string]interface{}{
		"Title": blog.Title,
	}).Update(blog).RunWrite(dbSession)
	return err
}

// DBFetchBlogByCol fetches blog by column name
func DBFetchBlogByCol(colName, blogName string) (projectModels.StructureBlog, error) {
	res, err := r.Table("blogs").Filter(map[string]interface{}{
		colName: blogName,
	}).Run(dbSession)
	if err != nil {
		return projectModels.StructureBlog{}, err
	}

	if res.IsNil() {
		return projectModels.StructureBlog{}, errors.New("Empty Result")
	}
	var blog projectModels.StructureBlog
	if err = res.One(&blog); err != nil {
		return projectModels.StructureBlog{}, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	return blog, nil
}

// DBFetchXBlogItems fetches x blogs after skipping y blogs
func DBFetchXBlogItems(skipN, limitN int) ([]projectModels.StructureBlog, int, error) {
	resCountQuery, err := r.Table("blogs").OrderBy(r.Desc("Date")).Count().Run(dbSession)
	if err != nil {
		return []projectModels.StructureBlog{}, 0, err
	}

	if resCountQuery.IsNil() {
		return []projectModels.StructureBlog{}, 0, errors.New("Empty Result")
	}

	var resCount int
	err = resCountQuery.One(&resCount)
	if err != nil {
		return []projectModels.StructureBlog{}, 0, err
	}

	err = resCountQuery.Close()
	common.CheckErrorAndPanic(err)

	resp, err := r.Table("blogs").OrderBy(r.Asc("Date")).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []projectModels.StructureBlog{}, 0, err
	}

	if resp.IsNil() {
		return []projectModels.StructureBlog{}, 0, errors.New("Empty Result")
	}

	var allBlogs []projectModels.StructureBlog
	if err = resp.All(&allBlogs); err != nil {
		return []projectModels.StructureBlog{}, 0, err
	}

	err = resp.Close()
	common.CheckErrorAndPanic(err)

	if len(allBlogs) == 0 {
		return []projectModels.StructureBlog{}, 0, errors.New("Empty Result")
	}
	return allBlogs, resCount, nil
}

// DBCountSize counts the size of a given table
func DBCountSize(tablename string) (int, error) {
	res, err := r.Table(tablename).Count().Run(dbSession)
	if err != nil {
		return 0, err
	}

	var count int
	err = res.One(&count)
	if err != nil {
		return 0, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	return count, nil
}

// DBFetchXOngoingAnime fetches x currently-airing animes after skipping y animes
func DBFetchXOngoingAnime(skipN int, limitN int) ([]projectModels.StructureAnime, int, error) {
	resCountQuery, err := r.Table("animes").Filter(map[string]interface{}{
		"Status": "Currently Airing",
	}).Count().Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if resCountQuery.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int
	err = resCountQuery.One(&resCount)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = resCountQuery.Close()
	common.CheckErrorAndPanic(err)

	res, err := r.Table("animes").Filter(map[string]interface{}{
		"Status": "Currently Airing",
	}).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if res.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var animelist []projectModels.StructureAnime
	if err = res.All(&animelist); err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}
	return animelist, resCount, nil
}

// DBFetchXTopRatingAnime fetches x top-rated animes after skipping y animes
func DBFetchXTopRatingAnime(skipN int, limitN int) ([]projectModels.StructureAnime, int, error) {

	resCountQuery, err := r.Table("animes").OrderBy(r.Desc("Score")).Count().Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}
	if resCountQuery.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int

	err = resCountQuery.One(&resCount)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = resCountQuery.Close()
	common.CheckErrorAndPanic(err)

	res, err := r.Table("animes").OrderBy(r.Desc("Score")).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	// Not working: https://github.com/dancannon/gorethink/issues/326
	if res.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var animelist []projectModels.StructureAnime
	if err = res.All(&animelist); err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}
	return animelist, resCount, nil
}

// DBFetchXSeasonalAnime fetches x seasonal animes
func DBFetchXSeasonalAnime(value int) ([]projectModels.StructureAnime, error) {

	resp, err := r.Table("cache_seasonal_animes").OrderBy(r.Asc("id")).Limit(value).Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, err
	}

	if resp.IsNil() {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}

	var seasonalAnimeIDs []map[string]interface{}
	if err = resp.All(&seasonalAnimeIDs); err != nil {
		return []projectModels.StructureAnime{}, err
	}

	err = resp.Close()
	common.CheckErrorAndPanic(err)

	var animelist []projectModels.StructureAnime
	for _, v := range seasonalAnimeIDs {
		res, err := r.Table("animes").Filter(map[string]interface{}{
			"id": v["malID"].(string),
		}).Run(dbSession)
		if err != nil {
			return []projectModels.StructureAnime{}, err
		}
		if res.IsNil() {
			return []projectModels.StructureAnime{}, errors.New("Empty Result")
		}
		var anime projectModels.StructureAnime
		if err = res.One(&anime); err != nil {
			return []projectModels.StructureAnime{}, err
		}
		err = res.Close()
		common.CheckErrorAndPanic(err)

		animelist = append(animelist, anime)
	}

	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}
	return animelist, nil
}

// DBFetchXPopularAnime fetches x popular animes
func DBFetchXPopularAnime(skipN, limitN int) ([]projectModels.StructureAnime, int, error) {

	resCountQuery, err := r.Table("cache_popular_animes").Count().Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if resCountQuery.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int
	err = resCountQuery.One(&resCount)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = resCountQuery.Close()
	common.CheckErrorAndPanic(err)

	resp, err := r.Table("cache_popular_animes").OrderBy("id").Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if resp.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var popularAnimeIDs []map[string]interface{}
	if err = resp.All(&popularAnimeIDs); err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = resp.Close()
	common.CheckErrorAndPanic(err)

	var animelist []projectModels.StructureAnime
	for _, v := range popularAnimeIDs {
		res, err := r.Table("animes").Filter(map[string]interface{}{
			"id": v["malID"].(string),
		}).Run(dbSession)
		if err != nil {
			return []projectModels.StructureAnime{}, 0, err
		}
		if res.IsNil() {
			return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
		}
		var anime projectModels.StructureAnime
		if err = res.One(&anime); err != nil {
			return []projectModels.StructureAnime{}, 0, err
		}
		err = res.Close()
		common.CheckErrorAndPanic(err)

		animelist = append(animelist, anime)
	}
	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}
	return animelist, resCount, nil
}

// DBFetchFeaturedAnimes fetches handpicked featured animes
func DBFetchFeaturedAnimes() ([]projectModels.StructureAnime, error) {
	/*
			r.db("animedom").table("animes").filter(function(doc){
		  return doc("id").eq("199").
		    or(doc("id").eq("31240")).
		    or(doc("id").eq("249")).
		    or(doc("id").eq("4722")).
		    or(doc("id").eq("205")).
		    or(doc("id").eq("431")).
		    or(doc("id").eq("534")).
		    or(doc("id").eq("20"))
		})
	*/
	resp, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("id").Eq("199").
			Or(anime.Field("id").Eq("31240")).
			Or(anime.Field("id").Eq("249")).
			Or(anime.Field("id").Eq("4722")).
			Or(anime.Field("id").Eq("205")).
			Or(anime.Field("id").Eq("431")).
			Or(anime.Field("id").Eq("534")).
			Or(anime.Field("id").Eq("20"))
	}).Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, err
	}

	if resp.IsNil() {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}

	var featuredAnimes []projectModels.StructureAnime
	if err = resp.All(&featuredAnimes); err != nil {
		return []projectModels.StructureAnime{}, err
	}

	err = resp.Close()
	common.CheckErrorAndPanic(err)

	if len(featuredAnimes) == 0 {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}
	return featuredAnimes, nil
}

// DBFetchXRecentEpisodes fetches x recently updated episodes after skipping y anime episodes
func DBFetchXRecentEpisodes(skipN, limitN int) ([]projectModels.StructureAnime, []projectModels.StructureRecentAnimesDbHelper, int, error) {
	resCountQuery, err := r.Table("cache_recent_animes").OrderBy(r.Asc("id")).Field("malID").Count().Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, err
	}

	if resCountQuery.IsNil() {
		return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, errors.New("Empty Result")
	}

	var resCount int

	err = resCountQuery.One(&resCount)
	if err != nil {
		return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, err
	}

	err = resCountQuery.Close()
	common.CheckErrorAndPanic(err)

	resp, err := r.Table("cache_recent_animes").OrderBy(r.Asc("id")).Pluck("malID", "episodeIndex").Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, err
	}

	if resp.IsNil() {
		return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, errors.New("Empty Result")
	}
	recentAnimes := []projectModels.StructureRecentAnimesDbHelper{}
	if err = resp.All(&recentAnimes); err != nil {
		return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, err
	}

	err = resp.Close()
	common.CheckErrorAndPanic(err)

	var animelist []projectModels.StructureAnime

	for i := 0; i < len(recentAnimes); i++ {
		res, err := r.Table("animes").Filter(map[string]interface{}{
			"id": recentAnimes[i].MALID,
		}).Run(dbSession)
		if err != nil {
			return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, err
		}

		if res.IsNil() {
			return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, errors.New("Empty Result")
		}

		var tempAnimelist projectModels.StructureAnime
		if err = res.One(&tempAnimelist); err != nil {
			return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, err
		}

		err = res.Close()
		common.CheckErrorAndPanic(err)

		animelist = append(animelist, tempAnimelist)
	}
	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, []projectModels.StructureRecentAnimesDbHelper{}, 0, errors.New("Empty Result")
	}

	return animelist, recentAnimes, resCount, nil
}

// DBFetchXMovies fetches x anime movies after skipping y animes
func DBFetchXMovies(skipN int, limitN int) ([]projectModels.StructureAnime, int, error) {
	resCountQuery, err := r.Table("animes").OrderBy(r.Desc("Score")).Filter(map[string]interface{}{
		"Type": "Movie",
	}).Count().Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if resCountQuery.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int

	err = resCountQuery.One(&resCount)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = resCountQuery.Close()
	common.CheckErrorAndPanic(err)

	res, err := r.Table("animes").OrderBy(r.Desc("Score")).Filter(map[string]interface{}{
		"Type": "Movie",
	}).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	// Not working: https://github.com/dancannon/gorethink/issues/326
	if res.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var animelist []projectModels.StructureAnime
	if err = res.All(&animelist); err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}
	return animelist, resCount, nil
}

// DBCreateSearchJSON fetches and prepares the search result json
func DBCreateSearchJSON(term string, limit int, baseUrl string) ([]projectModels.StructureSearchJSON, error) {
	//r.db("animedom")
	//.table("animes")
	//.filter(function(doc){
	//	return doc('MALTitle').downcase().match("attack")
	//	.or(doc('MALEnglish').downcase().match("attack"))
	//}).pluck("MALTitle", "id", "Slug").orderBy("MALTitle")
	res, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("MALTitle").Downcase().Match(strings.ToLower(term)).
			Or(anime.Field("MALEnglish").Downcase().Match(strings.ToLower(term)))
	}).OrderBy("MALTitle").Pluck("MALTitle", "MALEnglish", "id", "Slug", "Genre", "Type", "WikiFNVHash").Limit(limit).Run(dbSession)
	if err != nil {
		return []projectModels.StructureSearchJSON{}, err
	}

	if res.IsNil() {
		return []projectModels.StructureSearchJSON{}, errors.New("Empty Result")
	}

	var mappedData []projectModels.StructureAnime

	err = res.All(&mappedData)
	if err != nil {
		return []projectModels.StructureSearchJSON{}, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	var readyData []projectModels.StructureSearchJSON

	imgCDNUrl := baseUrl
	if imgCDNUrl != "http://localhost:1993" {
		if imgCDNUrl[0:5] == "https" {
			imgCDNUrl = imgCDNUrl[0:5] + "://cdn." + imgCDNUrl[8:]
		} else {
			imgCDNUrl = imgCDNUrl[0:4] + "://cdn." + imgCDNUrl[7:]
		}
	}

	for i := 0; i < len(mappedData); i++ {
		readyDataInstance := projectModels.StructureSearchJSON{}

		// Assign Name
		if mappedData[i].MALEnglish != "" {
			readyDataInstance.AnimeName = mappedData[i].MALEnglish
		} else {
			readyDataInstance.AnimeName = mappedData[i].MALTitle
		}

		// Assign URL
		readyDataInstance.AnimeUrl = baseUrl + "/wiki/" +
			mappedData[i].MALID + "/" + mappedData[i].WikiHash

		// Assign Genres
		for index := 0; index < len(mappedData[i].Genre); index++ {
			if index == len(mappedData[i].Genre)-1 {
				readyDataInstance.AnimeGenre += mappedData[i].Genre[index]
			} else {
				readyDataInstance.AnimeGenre += mappedData[i].Genre[index] + ", "
			}
		}

		// Assign Thumbnail
		readyDataInstance.AnimeThumb = imgCDNUrl + "/assets/img/smallestanime/" + mappedData[i].MALID + ".jpg"

		// Assign Type
		readyDataInstance.AnimeType = mappedData[i].Type

		// IsLast?
		readyDataInstance.IsLast = false

		readyData = append(readyData, readyDataInstance)
	}

	readyData = append(readyData, projectModels.StructureSearchJSON{IsLast: true, SearchTerm: term})

	return readyData, nil
}

// DBXFetchAnimesByGenre fetches x anime by genre after skipping y animes
func DBXFetchAnimesByGenre(queries [][]byte, skipN int, limitN int) ([]projectModels.StructureAnime, int, error) {
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
		return []projectModels.StructureAnime{}, 0, err
	}

	if resCountQuery.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int

	err = resCountQuery.One(&resCount)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = resCountQuery.Close()
	common.CheckErrorAndPanic(err)

	res, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("Genre").Contains(genres...)
	}).Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if res.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var animelist []projectModels.StructureAnime

	err = res.All(&animelist)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	return animelist, resCount, nil
}

func DBXFetchAnimesByGenreSample(score, resWanted int, moodsContain, moodsAvoid []string) ([]projectModels.StructureAnime, error) {
	//  r.db("animedom").table('animes').filter(function (anime){
	//      return anime.getField("Genre").contains('Action', 'Fantasy', 'Shoujo');
	//  }).getField('id')
	//

	genresInc := make([]interface{}, len(moodsContain))
	for i, v := range moodsContain {
		genresInc[i] = v
	}

	genresAvoid := make([]interface{}, len(moodsAvoid))
	for i, v := range moodsAvoid {
		genresAvoid[i] = v
	}

	res, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("Genre").Contains(genresInc...).
			And(anime.Field("Genre").Contains(genresAvoid).Not()).
			And(anime.Field("Score").Ge(score))
	}).Sample(resWanted).Run(dbSession)
	common.CheckErrorAndPanic(err)

	if res.IsNil() {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}

	var animelist []projectModels.StructureAnime

	err = res.All(&animelist)
	if err != nil {
		return []projectModels.StructureAnime{}, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}

	return animelist, nil
}

func DBXFetchClassicsSample(score, resWanted int, olderThan string) ([]projectModels.StructureAnime, error) {
	//classics
	//r.db("animedom").table("animes").filter(function(doc){
	//return doc('Year').ne('0000-00-00').and(doc('Year').lt("2005-05-15")).and(doc('Score').ge(7))
	//}).count()
	res, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("Year").Ne("0000-00-00").
			And(anime.Field("Year").Lt(olderThan)).
			And(anime.Field("Score").Ge(score))
	}).Sample(resWanted).Run(dbSession)
	common.CheckErrorAndPanic(err)

	if res.IsNil() {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}

	var animelist []projectModels.StructureAnime

	err = res.All(&animelist)
	if err != nil {
		return []projectModels.StructureAnime{}, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}

	return animelist, nil
}

// DBCheckExistsAnimesByID checks if anime exists by MALID
func DBCheckExistsAnimesByID(id string) error {
	res, err := r.Table("animes").Filter(map[string]interface{}{
		"id": id,
	}).Run(dbSession)
	if err != nil {
		return err
	}

	if res.IsNil() {
		return errors.New("Empty Result")
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	return nil
}

// DBPushPopularAnimes adds a new popular anime to popular_anime table
func DBPushPopularAnimes(index int, malID string) {
	_, err := r.Table("popular_animes").Insert(map[string]interface{}{
		"id":    index,
		"malID": malID,
	}).RunWrite(dbSession)
	if err != nil {
		log.Println(err)
		return
	}
}

// DBPushSeasonalAnimes adds a new popular anime to seasonal_anime table
func DBPushSeasonalAnimes(index int, malID string) {
	_, err := r.Table("seasonal_animes").Insert(map[string]interface{}{
		"id":    index,
		"malID": malID,
	}).RunWrite(dbSession)
	if err != nil {
		log.Println(err)
		return
	}
}

// DBTruncateTable truncates a given table
func DBTruncateTable(tableName string) error {
	_, err := r.Table(tableName).Delete().RunWrite(dbSession)
	return err
}

// DBInsertNewAnime adds a new anime to database
func DBInsertNewAnime(anime projectModels.StructureAnime) error {
	err := r.DB("animedom").Table("animes").Insert(anime).Exec(dbSession)
	return err
}

// DBFetchAnimeByCol fetches a single anime by name by given column name
func DBFetchAnimeByCol(colName, objectName string) (projectModels.StructureAnime, error) {
	res, err := r.Table("animes").Filter(map[string]interface{}{
		colName: objectName,
	}).Run(dbSession)
	if err != nil {
		return projectModels.StructureAnime{}, err
	}

	if res.IsNil() {
		return projectModels.StructureAnime{}, errors.New("Empty Result")
	}

	var anime projectModels.StructureAnime
	if err = res.One(&anime); err != nil {
		panic(err)
		return projectModels.StructureAnime{}, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	return anime, nil
}

// DBGetAllAnime fetches all database animes
func DBGetAllAnime() ([]projectModels.StructureAnime, error) {
	res, err := r.Table("animes").OrderBy("MALTitle").Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, err
	}

	if res.IsNil() {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}

	var animelist []projectModels.StructureAnime
	if err = res.All(&animelist); err != nil {
		return []projectModels.StructureAnime{}, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, errors.New("Empty Result")
	}
	return animelist, nil
}

// DBCopyTableTo copies from one table to another
func DBCopyTableTo(from, to string) error {
	_, err := r.Table(to).Delete().RunWrite(dbSession)
	if err != nil {
		return err
	}
	_, err = r.Table(to).Insert(r.Table(from)).RunWrite(dbSession)
	return err
}

// DBPushRecentAnimes adds a new recent anime to recent_anime table
func DBPushRecentAnimes(index int, malID string, episodeIndex int) {
	_, err := r.Table("recent_animes").Insert(map[string]interface{}{
		"id":           index,
		"malID":        malID,
		"episodeIndex": episodeIndex,
	}).RunWrite(dbSession)
	if err != nil {
		log.Println(err)
		return
	}
}

// DBUpdateEpisodelist updates the episode listing
func DBUpdateEpisodelist(episodes []projectModels.StructureEpisode, malID string) error {
	_, err := r.Table("animes").Filter(map[string]interface{}{
		"id": malID,
	}).Update(map[string]interface{}{
		"EpisodeList": episodes,
	}).RunWrite(dbSession)
	return err
}

// DBModifyAnime updates a given anime by id
func DBModifyAnime(anime projectModels.StructureAnime, malID string) error {
	_, err := r.Table("animes").Filter(map[string]interface{}{
		"id": malID,
	}).Update(anime).RunWrite(dbSession)
	return err
}

// DBGetRandomAnimeByScore Gets a random anime greater than score provided
func DBGetRandomAnimeByScore(score int) projectModels.StructureAnime {
	//r.db("animedom").table("animes").filter(function(doc){
	//	return doc('Score').ge(6)
	//}).sample(1)

	res, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("Score").Ge(score)
	}).Sample(1).Run(dbSession)
	common.CheckErrorAndPanic(err)

	var anime projectModels.StructureAnime
	err = res.One(&anime)
	common.CheckErrorAndPanic(err)

	return anime
}

func DBFetchAnimeAZList(skipN, limitN int, alphabet string) ([]projectModels.StructureAnime, int, error) {
	//r.db("animedom").table("animes").filter(function(doc){
	//	return doc('MALTitle').split("").nth(0).eq("M")
	//}).orderBy("MALTitle").limit(20)

	//r.db("animedom").table("animes").filter(function(doc){
	//	return doc('MALTitle').match("^[ABCDEFGHIJKLMNOPQRSTUVWXYZ]").not()
	//}).orderBy("MALTitle").limit(20)

	alphabets := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	resCountQuery := new(r.Cursor)
	var err error

	if !common.StringInSlice(alphabet, alphabets) {
		resCountQuery, err = r.Table("animes").Filter(func(anime r.Term) r.Term {
			return anime.Field("MALTitle").Match("^[ABCDEFGHIJKLMNOPQRSTUVWXYZ]").Not()
		}).Count().Run(dbSession)
	} else {
		resCountQuery, err = r.Table("animes").Filter(func(anime r.Term) r.Term {
			return anime.Field("MALTitle").Match("^" + alphabet)
		}).Count().Run(dbSession)
	}
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if resCountQuery.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int
	err = resCountQuery.One(&resCount)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = resCountQuery.Close()
	common.CheckErrorAndPanic(err)

	res := new(r.Cursor)
	if !common.StringInSlice(alphabet, alphabets) {
		res, err = r.Table("animes").Filter(func(anime r.Term) r.Term {
			return anime.Field("MALTitle").Match("^[ABCDEFGHIJKLMNOPQRSTUVWXYZ]").Not()
		}).OrderBy("MALTitle").Pluck("id", "MALTitle", "Slug", "Score", "Genre", "Type", "Status", "WikiFNVHash").Skip(skipN).Limit(limitN).Run(dbSession)
	} else {
		res, err = r.Table("animes").Filter(func(anime r.Term) r.Term {
			return anime.Field("MALTitle").Match("^" + alphabet)
		}).OrderBy("MALTitle").Pluck("id", "MALTitle", "Slug", "Score", "Genre", "Type", "Status", "WikiFNVHash").Skip(skipN).Limit(limitN).Run(dbSession)
	}
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if res.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var animelist []projectModels.StructureAnime
	if err = res.All(&animelist); err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	if len(animelist) == 0 {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}
	return animelist, resCount, nil
}

func DBAdvancedSearch(term string, skipN, limitN int) ([]projectModels.StructureAnime, int, error) {
	//r.db("animedom")
	//.table("animes")
	//.filter(function(doc){
	//	return doc('MALTitle').downcase().match("attack")
	//	.or(doc('MALEnglish').downcase().match("attack"))
	//}).pluck("MALTitle", "id", "Slug").orderBy("MALTitle")

	resCountQuery, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("MALTitle").Downcase().Match(strings.ToLower(term)).
			Or(anime.Field("MALEnglish").Downcase().Match(strings.ToLower(term)))
	}).Count().Run(dbSession)

	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if resCountQuery.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var resCount int
	err = resCountQuery.One(&resCount)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	res, err := r.Table("animes").Filter(func(anime r.Term) r.Term {
		return anime.Field("MALTitle").Downcase().Match(strings.ToLower(term)).
			Or(anime.Field("MALEnglish").Downcase().Match(strings.ToLower(term)))
	}).OrderBy("MALTitle").Pluck("MALTitle", "MALEnglish", "id", "Slug", "Genre", "Type", "MALDescription", "Score", "WikiFNVHash").Skip(skipN).Limit(limitN).Run(dbSession)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	if res.IsNil() {
		return []projectModels.StructureAnime{}, 0, errors.New("Empty Result")
	}

	var mappedData []projectModels.StructureAnime

	err = res.All(&mappedData)
	if err != nil {
		return []projectModels.StructureAnime{}, 0, err
	}

	err = res.Close()
	common.CheckErrorAndPanic(err)

	return mappedData, resCount, nil
}
