package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/subosito/gotenv"
)

func TestInspirationIndexURL(t *testing.T) {
	gotenv.MustLoad(os.Getenv("GOPATH") + "/src/github.com/bukalapak/jenkins_jr/.env")
	os.Setenv("ENV", "test")

	url := os.Getenv("INSPIRATION_INDEX_URL")
	assert.Equal(t, url, InspirationIndexURL())

	os.Setenv("INSPIRATION_INDEX_URL", "")
	assert.Equal(t, InspirationIndexDefaultURL, InspirationIndexURL())
	os.Setenv("INSPIRATION_INDEX_URL", url)
}
