package models

type IndustryIdentifiers struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

type ReadingModes struct {
	Text  bool `json:"text"`
	Image bool `json:"image"`
}

type PanelizationSummary struct {
	ContainsEpubBubbles  bool `json:"containsEpubBubbles"`
	ContainsImageBubbles bool `json:"containsImageBubbles"`
}

type ImageLinks struct {
	SmallThumbnail string `json:"smallThumbnail"`
	Thumbnail      string `json:"thumbnail"`
}

type VolumeInfo struct {
	Title               string                `json:"title"`
	Subtitle            string                `json:"subtitle"`
	Authors             []string              `json:"authors"`
	Publisher           string                `json:"publisher"`
	PublishedDate       string                `json:"publishedDate"`
	Description         string                `json:"description"`
	IndustryIdentifiers []IndustryIdentifiers `json:"industryIdentifiers"`
	ReadingModes        ReadingModes          `json:"readingModes"`
	PageCount           int                   `json:"pageCount"`
	PrintType           string                `json:"printType"`
	Categories          []string              `json:"categories"`
	AverageRating       float64               `json:"averageRating"`
	RatingsCount        int                   `json:"ratingsCount"`
	MaturityRating      string                `json:"maturityRating"`
	AllowAnonLogging    bool                  `json:"allowAnonLogging"`
	ContentVersion      string                `json:"contentVersion"`
	PanelizationSummary PanelizationSummary   `json:"panelizationSummary"`
	ImageLinks          ImageLinks            `json:"imageLinks"`
	Language            string                `json:"language"`
	PreviewLink         string                `json:"previewLink"`
	InfoLink            string                `json:"infoLink"`
	CanonicalVolumeLink string                `json:"canonicalVolumeLink"`
}

type SaleInfo struct {
	Country     string `json:"country"`
	Saleability string `json:"saleability"`
	IsEbook     bool   `json:"isEbook"`
}

type AccessInfo struct {
	Country                string `json:"country"`
	Viewability            string `json:"viewability"`
	Embeddable             bool   `json:"embeddable"`
	PublicDomain           bool   `json:"publicDomain"`
	TextToSpeechPermission string `json:"textToSpeechPermission"`
	Epub                   struct {
		IsAvailable bool `json:"isAvailable"`
	} `json:"epub"`
	Pdf struct {
		IsAvailable bool `json:"isAvailable"`
	} `json:"pdf"`
	WebReaderLink       string `json:"webReaderLink"`
	AccessViewStatus    string `json:"accessViewStatus"`
	QuoteSharingAllowed bool   `json:"quoteSharingAllowed"`
}

type SearchInfo struct {
	TextSnippet string `json:"textSnippet"`
}

type BookInfo struct {
	Kind       string     `json:"kind"`
	ID         string     `json:"id"`
	Etag       string     `json:"etag"`
	SelfLink   string     `json:"selfLink"`
	VolumeInfo VolumeInfo `json:"volumeInfo"`
	SaleInfo   SaleInfo   `json:"saleInfo"`
	AccessInfo AccessInfo `json:"accessInfo"`
	SearchInfo SearchInfo `json:"searchInfo"`
}

type Result struct {
	Kind       string     `json:"kind"`
	TotalItems int        `json:"totalItems"`
	Items      []BookInfo `json:"items"`
}