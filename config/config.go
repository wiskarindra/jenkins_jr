package config

import "os"

const (
	DatabaseDatetimeFormat     = "2006-01-02 15:04:05"
	ResponseDatetimeFormat     = "2006-01-02T15:04:05Z"
	DateRangeSearchFormat      = "2006-01-02T15:04:05Z07:00"
	InspirationIndexDefaultURL = "https://www.bukalapak.com/inspirasi"
	InfluencerIndexDefaultURL  = "https://www.bukalapak.com/i"
	InspirationHomepageTitle   = "Inspirasi"
	IndexImageURLStyle         = "s-1080-1350"
	HomepageImageURLStyle      = "s-240-300"
)

func InspirationIndexURL() string {
	url := os.Getenv("INSPIRATION_INDEX_URL")
	if url == "" {
		url = InspirationIndexDefaultURL
	}
	return url
}

func InfluencerIndexURL() string {
	url := os.Getenv("INFLUENCER_INDEX_URL")
	if url == "" {
		url = InfluencerIndexDefaultURL
	}
	return url
}
