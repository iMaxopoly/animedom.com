package gogoanimeAssistant

import (
	"fmt"
	"log"
	"math"
	"sync"
)

func RefreshAnimes() {
	defer func() {
		// Handle error
		if err := recover(); err != nil {
			log.Println(err)
		}
		// Job done notify
		log.Println("gogoanimeAssistant.RefreshAnimes Job done")
	}()
	log.Println("gogoanimeAssistant.RefreshAnimes Job Starting...")

	// Goroutines manager
	var wg sync.WaitGroup

	//mp4Order = make(chan string, 1)
	//	go downloadMp4uploadVideo(mp4Order)

	animelist := scrapeAnimelistPage()
	log.Print("Total animes found(with dupes too):", len(animelist))

	fmt.Println("Enter the dragon")
	fmt.Println(animelist, len(animelist))
	// Set num. of workers
	nWorkers := 15

	wg.Add(nWorkers)

	// Distribute workload to workers
	section := int(math.Floor(float64(len(animelist)) / float64(nWorkers)))
	startingPoint := 0
	remainder := len(animelist) - (section * nWorkers)
	for x := 0; x < nWorkers-1; x++ {
		go func(strt, en int) {
			defer wg.Done()
			mainscrape(job{whatKind: "RefreshAnimes", start: strt, end: en, allAnimes: animelist})
		}(func() (int, int) {
			start := startingPoint
			startingPoint = startingPoint + section
			end := startingPoint
			return start, end
		}())
	}
	go func(strt, en int) {
		defer wg.Done()
		mainscrape(job{whatKind: "RefreshAnimes", start: strt, end: en, allAnimes: animelist})
	}(func() (int, int) {
		start := startingPoint
		startingPoint = startingPoint + remainder
		end := startingPoint
		return start, end
	}())

	wg.Wait()
	//close(mp4Order)
}
