package myanimelistAssistant

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"animedom.com/common"
	"animedom.com/projectModels"

	"github.com/PuerkitoBio/goquery"
	"github.com/disintegration/imaging"
	"github.com/kennygrant/sanitize"
)

var malNameFixList []string
var gogoanimeNameList []string

func init() {
	{
		log.Println("Loading malfix.txt")
		f, err := os.Open("malfix.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Create a new Scanner for the file.
		scanner := bufio.NewScanner(f)
		// Loop over all lines in the file.
		for scanner.Scan() {
			line := scanner.Text()
			malNameFixList = append(malNameFixList, line)
		}
	}
	log.Println("Loading gogoanimeraw.txt")
	f, err := os.Open("gogoanimeraw.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Create a new Scanner for the file.
	scanner := bufio.NewScanner(f)
	// Loop over all lines in the file.
	for scanner.Scan() {
		line := scanner.Text()
		gogoanimeNameList = append(gogoanimeNameList, line)
	}
	//generateImages()
}

//
//func generateImages() {
//	animes, err := dbOperations.DBGetAllAnime()
//	common.CheckErrorAndPanic(err)
//
//	total := len(animes)
//	for i, anime := range animes {
//		fmt.Println(i, "/", total, "images fetched.")
//		err = MalFetchImage(anime.Image, anime.MALID)
//		common.CheckErrorAndPanic(err)
//	}
//	fmt.Println("Image fetching done.")
//}

func CorrectNameForMALLookup(val string) string {
	foundIndex := -1
	for i, v := range gogoanimeNameList {
		if v == val {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return ""
	}

	return malNameFixList[foundIndex]
}

func MalAPIFetchData(animeName string) (projectModels.StructureMALApiAnime, error) {
	client := &http.Client{}

	animeName = strings.Replace(animeName, " ", "+", -1)

	req, err := http.NewRequest("GET", fmt.Sprintf("http://myanimelist.net/api/anime/search.json?q=%s", animeName), nil)
	req.SetBasicAuth(malUsername, malPassword)
	resp, err := client.Do(req)
	if err != nil {
		return projectModels.StructureMALApiAnime{}, err
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return projectModels.StructureMALApiAnime{}, err
	}

	xobj := projectModels.StructureMALApiAnime{}
	err = xml.Unmarshal(bodyText, &xobj)
	if err != nil {
		return projectModels.StructureMALApiAnime{}, err
	}
	return xobj, nil
}

func FindIDinMALXML(malAnimeData projectModels.StructureMALApiAnime, fixedAnimeName string) int {
	checkSynonyms := func(goganimeName, synonyms string) bool {
		synonymList := strings.Split(synonyms, ";")
		for i := 0; i < len(synonymList); i++ {
			if goganimeName == strings.TrimSpace(synonymList[i]) {
				return true
			}
		}
		return false
	}

	for index, malEntry := range malAnimeData.Entries {
		if strings.ToLower(fixedAnimeName) == strings.ToLower(malEntry.Title) ||
			strings.ToLower(fixedAnimeName) == strings.ToLower(malEntry.EnglishName) ||
			checkSynonyms(strings.ToLower(fixedAnimeName), strings.ToLower(malEntry.SynonymNames)) {
			return index
		}
	}
	return -1
}

func MalFetchEpisodelist(malID string) []string {
	doc, err := goquery.NewDocument(fmt.Sprintf("http://myanimelist.net/anime/%s", malID))
	if err != nil {
		panic(err)
	}

	var episodesURL string
	horizontalDoc := doc.Find("#horiznav_nav").Find("ul").Find("a")
	for node := range horizontalDoc.Nodes {
		if horizontalDoc.Eq(node).Text() == "Episodes" {
			link, exists := horizontalDoc.Eq(node).Attr("href")
			if !exists || link == "" {
				panic(errors.New("Episode link doesnt exist in MAL " + malID))
			}
			episodesURL = link
		}
	}

	if episodesURL == "" {
		return []string{}
	}

	epiDoc, err := goquery.NewDocument(episodesURL)
	if err != nil {
		panic(err)
	}

	pagesDoc := epiDoc.Find(".pagination.ac").Find("a")

	pages := []string{}
	for node := range pagesDoc.Nodes {
		link, exists := pagesDoc.Eq(node).Attr("href")
		if !exists {
			panic(errors.New("Pagination detected by link not found"))
		}
		pages = append(pages, link)
	}

	var result []string

	for _, page := range pages {
		pDoc, err := goquery.NewDocument(page)
		if err != nil {
			panic(err)
		}
		episodeList := pDoc.Find(".fl-l.fw-b")

		for node := range episodeList.Nodes {
			episode := episodeList.Eq(node)
			result = append(result, episode.Text())
		}
	}

	return result
}

func MalFetchTrailer(malID string) string {
	doc, err := goquery.NewDocument(fmt.Sprintf("http://myanimelist.net/anime/%s/", malID))
	if err != nil {
		panic(err)
	}

	trailer, exists := doc.Find(".iframe.js-fancybox-video.video-unit.promotion").Attr("href")

	if !exists {
		return "nil"
	}

	if strings.Contains(trailer, "void") || strings.Contains(trailer, "nil") {
		return "nil"
	}

	trailer = strings.Replace(trailer, "http://www.youtube.com/embed/", "", 1)
	for i, v := range trailer {
		if v == '?' {
			trailer = trailer[0:i]
			break
		}
	}

	return trailer
}

func MalFetchImage(path, name string) error {
	response, err := http.Get(path)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	imgBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if len(imgBytes) < 3000 {
		// Possibly damaged
		doc, err := goquery.NewDocument(fmt.Sprintf("http://myanimelist.net/anime/%s", name))
		if err != nil {
			return err
		}

		imgLink, exists := doc.Find(".ac").Attr("src")
		if !exists {
			return errors.New(fmt.Sprintf("Image doesn't exist: NAME[%s] - IMGLINK[%s] - PATH[%s]", name, imgLink, path))
		}

		resp, err := http.Get(imgLink)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		imgBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}

	imgImage, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return err
	}

	newImageSmall := imaging.Resize(imgImage, 210, 250, imaging.Lanczos)
	newImageSmallest := imaging.Resize(imgImage, 120, 160, imaging.Lanczos)

	var savLocSmallest string
	var savLocSmall string
	var savLoc string
	if !common.Production {
		savLocSmallest = "assets/img/smallestanime/%s.jpg"
		savLocSmall = "assets/img/smallanime/%s.jpg"
		savLoc = "assets/img/anime/%s.jpg"
	} else {
		savLocSmallest = "expose/assets/img/smallestanime/%s.jpg"
		savLocSmall = "expose/assets/img/smallanime/%s.jpg"
		savLoc = "expose/assets/img/anime/%s.jpg"
	}

	out, err := os.Create(fmt.Sprintf(savLocSmall, name))
	if err != nil {
		return err
	}

	err = jpeg.Encode(out, newImageSmall, nil)
	if err != nil {
		return err
	}

	out, err = os.Create(fmt.Sprintf(savLocSmallest, name))
	if err != nil {
		return err
	}

	err = jpeg.Encode(out, newImageSmallest, nil)
	if err != nil {
		return err
	}

	// open a file for writing
	err = ioutil.WriteFile(fmt.Sprintf(savLoc, name), imgBytes, 0644)
	return err
}

func CleanMALDescription(s string) string {
	s = strings.Replace(s, "[i]", "", -1)
	s = strings.Replace(s, "[/i]", "", -1)
	return sanitize.HTML(s)
}

func MalFetchGenre(malID string) []string {
	doc, err := goquery.NewDocument(fmt.Sprintf("http://myanimelist.net/anime/%s", malID))
	if err != nil {
		panic(err)
	}

	sideDoc := doc.Find(".borderClass").Find("span")

	var genres []string

	for node := range sideDoc.Nodes {
		singleDoc := sideDoc.Eq(node)
		if singleDoc.Text() == "Genres:" {
			genreDoc := singleDoc.NextAllFiltered("a")
			for subnode := range genreDoc.Nodes {
				genre, err := genreDoc.Eq(subnode).Html()
				if err != nil {
					panic(err)
				}
				genres = append(genres, genre)
			}
		}
	}

	return genres
}
