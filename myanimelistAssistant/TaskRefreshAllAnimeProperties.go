package myanimelistAssistant

//import (
//	"fmt"
//	"log"
//	"math"
//	"strconv"
//	"strings"
//	"sync"
//
//	"animedom.com/dbOperations"
//	"animedom.com/projectModels"
//)
//
//// Goroutines manager
//var wg sync.WaitGroup
//
//func RefreshAllProperties() {
//	defer func() {
//		// Handle error
//		if err := recover(); err != nil {
//			log.Println(err)
//		}
//
//		// Job done notify
//		log.Println("myanimelistAssistant.RefreshAllProperties Job done")
//	}()
//
//	// Get all animes
//	animes, err := dbOperations.DBGetAllAnime()
//	if err != nil {
//		panic(err)
//	}
//
//	// Set num. of workers
//	nWorkers := 5
//
//	wg.Add(nWorkers)
//
//	// Distribute workload to workers
//	section := int(math.Floor(float64(len(animes)) / float64(nWorkers)))
//	startingPoint := 0
//	remainder := len(animes) - (section * nWorkers)
//	for x := 0; x < nWorkers-1; x++ {
//		go rf(startingPoint, startingPoint+section, animes)
//		startingPoint += section
//	}
//	go rf(startingPoint, startingPoint+remainder, animes)
//
//	wg.Wait()
//}
//
//func rf(startIndex, endIndex int, animes []projectModels.StructureAnime) {
//	defer func() {
//		// Handle error
//		if err := recover(); err != nil {
//			log.Println(err)
//		}
//		wg.Done()
//	}()
//
//	for index := startIndex; index < endIndex; index++ {
//		malAnimeData, err := MalAPIFetchData(animes[index].MALTitle)
//		if err != nil {
//			if err.Error() == "EOF" {
//				continue
//			}
//			panic(err)
//		}
//
//		// Find anime id in the myanimelist xml response
//		malIDIndex := FindIDinMALXML(malAnimeData, animes[index].MALTitle)
//		if malIDIndex == -1 {
//			log.Println("Mal ID not found", animes[index].MALTitle)
//			continue
//		}
//
//		animes[index].MALEnglish = malAnimeData.Entries[malIDIndex].EnglishName
//		animes[index].Genre = MalFetchGenre(animes[index].MALID)
//		animes[index].MALDescription = CleanMALDescription(malAnimeData.Entries[malIDIndex].Synopsis)
//		animes[index].Score = func() float64 {
//			var i float64
//			i, err = strconv.ParseFloat(malAnimeData.Entries[malIDIndex].Score, 64)
//			if err == nil {
//				return i
//			}
//			return 0.0
//		}()
//		animes[index].Status = malAnimeData.Entries[malIDIndex].Status
//		animes[index].Type = malAnimeData.Entries[malIDIndex].Type
//		animes[index].SynonymNames = func() []string {
//			s := strings.Split(malAnimeData.Entries[malIDIndex].SynonymNames, ";")
//			if len(s) < 1 {
//				return []string{}
//			}
//			var val []string
//			for _, v := range s {
//				val = append(val, strings.TrimSpace(v))
//			}
//			return val
//		}()
//		animes[index].Year = malAnimeData.Entries[malIDIndex].StartDate
//		animes[index].Image = strings.Replace(malAnimeData.Entries[malIDIndex].Image, ".jpg", "l.jpg", 1)
//		animes[index].Trailer = MalFetchTrailer(animes[index].MALID)
//
//		malEpisodeNames := MalFetchEpisodelist(malAnimeData.Entries[malIDIndex].ID)
//		if len(malEpisodeNames) > 0 {
//			fmt.Println(malEpisodeNames)
//		}
//		for i, episode := range animes[index].EpisodeList {
//
//			episode.Name = func() string {
//				if i >= len(malEpisodeNames) || i < 0 {
//					return ""
//				}
//				return malEpisodeNames[i]
//			}()
//		}
//
//		// Save to db
//		err = dbOperations.DBModifyAnime(animes[index], animes[index].MALID)
//		if err != nil {
//			panic(err)
//		}
//	}
//}
