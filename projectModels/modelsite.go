package projectModels

type StructurePage struct {
	BaseURL        string
	Selection      string
	Title          string
	PageNum        int
	Blogs          []StructureBlog
	Blog           StructureBlog
	GridAnimes     []StructureAnime
	TableAnimes    []StructureAnime
	AZAlphabet     string
	OngoingAnimes  []StructureAnime
	PopularAnimes  []StructureAnime
	FeaturedAnimes []StructureAnime
	RecentAnimes   []StructureAnime
	EpisodeIDs     []StructureRecentAnimesDbHelper
	Anime          StructureAnime
	SimilarAnimes  []StructureAnime
	EpisodeNum     int
	MirrorNum      int
	ResultCount    int
	ErrorMsg       string
}

type StructureNameAndLink struct {
	Name string
	Link string
}

// Main model
//type StructureAnime struct {
//	MALID          string             `json:"MAL ID" gorethink:"id"`
//	MALTitle       string             `json:"MAL Title" gorethink:"MALTitle"`
//	MALEnglish     string             `json:"MAL English Title" gorethink:"MALEnglish"`
//	Genre          []string           `json:"Genre" gorethink:"Genre"`
//	MALDescription string             `json:"MAL Description" gorethink:"MALDescription"`
//	Score          float64            `json:"Score" gorethink:"Score"`
//	Status         string             `json:"Status" gorethink:"Status"`
//	Type           string             `json:"Type" gorethink:"Type"`
//	SynonymNames   []string           `json:"Synonym Names" gorethink:"SynonymNames"` // Semi-colon+space seperated
//	Year           string             `json:"Year" gorethink:"Year"`
//	Image          string             `json:"Image Url" gorethink:"Image"`
//	Trailer        string             `json:"Trailer" gorethink:"Trailer"`
//	EpisodeList    []StructureEpisode `json:"Episode List" gorethink:"EpisodeList"`
//	Slug           string             `json:"slug" gorethink:"Slug"`
//	WikiHash       string             `json:"wikifnvhash" gorethink:"WikiFNVHash"`
//}
type StructureAnime struct {
	MALID                    string             `json:"MAL ID" gorethink:"id"`
	MALTitle                 string             `json:"MAL Title" gorethink:"MALTitle"`
	MALEnglish               string             `json:"MAL English Title" gorethink:"MALEnglish"`
	Genre                    []string           `json:"Genre" gorethink:"Genre"`
	MALDescription           string             `json:"MAL Description" gorethink:"MALDescription"`
	Score                    float64            `json:"Score" gorethink:"Score"`
	Status                   string             `json:"Status" gorethink:"Status"`
	Type                     string             `json:"Type" gorethink:"Type"`
	SynonymNames             []string           `json:"Synonym Names" gorethink:"SynonymNames"` // Semi-colon+space seperated
	Year                     string             `json:"Year" gorethink:"Year"`
	Image                    string             `json:"Image Url" gorethink:"Image"`
	Trailer                  string             `json:"Trailer" gorethink:"Trailer"`
	SubbedEpisodeList        []StructureEpisode `json:"Subbed Episode List" gorethink:"SubbedEpisodeList"`
	EnglishDubbedEpisodeList []StructureEpisode `json:"English Episode List" gorethink:"EngDubbedEpisodeList"`
	FrenchDubbedEpisodeList  []StructureEpisode `json:"French Episode List" gorethink:"FraDubbedEpisodeList"`
	SpanishDubbedEpisodeList []StructureEpisode `json:"Spanish Episode List" gorethink:"SpaDubbedEpisodeList"`
	GermanDubbedEpisodeList  []StructureEpisode `json:"German Episode List" gorethink:"GerDubbedEpisodeList"`
	RelatedAnimes            []string           `json:"Related Animes" gorethink:"RelatedAnimes"`
	Slug                     string             `json:"slug" gorethink:"Slug"`
	WikiHash                 string             `json:"wikifnvhash" gorethink:"WikiFNVHash"`
}

type StructureEpisode struct {
	Name      string            `json:"Episode Name" gorethink:"Name"`
	Mirrors   []StructureMirror `json:"Episode Mirrors" gorethink:"Mirrors"`
	EpisodeID string            `json:"Episode Number" gorethink:"EpisodeID"`
}

type StructureMirror struct {
	Name         string `json:"Mirror Name" gorethink:"Name"`
	EmbedCode    string `json:"Embed Code" gorethink:"EmbedCode"`
	Quality      string `json:"Quality" gorethink:"Quality"`
	MirrorOrigin string `json:"MirrorOrigin" gorethink:"MirrorOrigin"`
	MirrorHash   string `json:"mirrorfnvhash" gorethink:"MirrorFNVHash"`
}

// Search related
type StructureSearchJSON struct {
	AnimeUrl   string `json:"url"`
	AnimeName  string `json:"title"`
	AnimeGenre string `json:"genre"`
	AnimeType  string `json:"type"`
	AnimeThumb string `json:"img"`
	IsLast     bool   `json:"islast"`
	SearchTerm string `json:"searchterm"`
}

// Blog related
type StructureBlog struct {
	ChangePrint  string `gorethink:"ChangePrint"`
	Date         string `gorethink:"Date"`
	Title        string `gorethink:"Title"`
	Article      string `gorethink:"Article"`
	ArticleImage string `gorethink:"ArticleImage"`
	Keywords     string `gorethink:"Keywords"`
	Slug         string `gorethink:"Slug" ini:"-"`
}

// AZ list related
type StructureAZList struct {
	MALID string
	Url   string
	Name  string
	Type  string
	Genre string
	Score string
}
