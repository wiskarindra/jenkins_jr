COVERALLS_TOKEN = MTVaL1kvbEhnbkJBMC9VdVFlRUFncUhWQTZKQjUyQ1c1bEp6VkV3SQ==
OPENCOV_URL     = https://opencov.bukalapak.io
BRANCH          = $(shell git rev-parse --abbrev-ref HEAD)

test:
	go test ./...

dep:
	dep ensure

cover:
	go test -coverprofile=coverage.out && go tool cover -html=coverage.out

goverall:
	go get github.com/mattn/goveralls
	COVERALLS_TOKEN=$(COVERALLS_TOKEN) goveralls -endpoint=$(OPENCOV_URL) -v