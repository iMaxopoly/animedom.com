package gogoanimeAssistant

import (
	"log"
	"math"
	"sync"
	"os"
)

// Goroutines manager
var wg sync.WaitGroup

func writeToFileGo(work <-chan fileWrite) {
	for {
		select {
		case msg := <-work:
			file, err := os.OpenFile(msg.filename, os.O_APPEND | os.O_WRONLY, 0600)
			if err != nil{
				panic(err)
			}
			_, err = file.WriteString(msg.message + "\n")
			if err != nil{
				panic(err)
			}
			file.Close()
		}
	}
}

func UpdateAnimes() {
	animelist := scrapeAnimelistPage()
	log.Print("Total animes found(with dupes too):", len(animelist))

	// Set num. of workers
	nWorkers := 15

	wg.Add(nWorkers)

	fileWriterChan := make(chan fileWrite, 1)

	// Start goroutines waiting to write to files
	go writeToFileGo(fileWriterChan)

	// Distribute workload to workers
	section := int(math.Floor(float64(len(animelist)) / float64(nWorkers)))
	starting_point := 0
	remainder := len(animelist) - (section * nWorkers)
	for x := 0; x < nWorkers - 1; x++ {
		go startScrapin(starting_point, starting_point + section, animelist, fileWriterChan)
		starting_point += section
	}
	go startScrapin(starting_point, starting_point + remainder, animelist, fileWriterChan)

	wg.Wait()
}