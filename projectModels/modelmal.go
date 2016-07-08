package dbModels

import "encoding/xml"

/* myanimelist.net models */
type StructureMALApiAnime struct {
	Anime   xml.Name               `xml:"anime"`
	Entries []StructureMALApiEntry `xml:"entry"`
}

type StructureMALApiEntry struct {
	ID            string `xml:"id"`
	Title         string `xml:"title"`
	EnglishName   string `xml:"english"`
	SynonymNames  string `xml:"synonyms"` // Semi-colon+space seperated
	EpisodesCount string `xml:"episodes"`
	Score         string `xml:"score"`
	Type          string `xml:"type"`
	Status        string `xml:"status"`
	StartDate     string `xml:"start_date"`
	EndDate       string `xml:"end_date"`
	Synopsis      string `xml:"synopsis"`
	Image         string `xml:"image"`
}