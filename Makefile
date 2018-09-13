.PHONY: all test dep compile build push checkenv deploy kubefile setup migrate

IMAGE = registry.bukalapak.io/wiskarindra/jenkins_jr/$(svc)
DIRS  = $(shell cd deploy && ls -d */ | grep -v "_output")
FILE ?= deployment
ODIR := deploy/_output

export VERSION            ?= $(shell git show -q --format=%h)
export VAR_SERVICES       ?= $(DIRS:/=)
export VAR_KUBE_NAMESPACE ?= default
export VAR_CONSUL_PREFIX  ?= jenkins_jr

all: compile build push deploy

test:
	go test -v $$(go list ./... | grep -E "pkg")

dep:
	dep ensure -v -vendor-only

$(ODIR):
	@mkdir -p $(ODIR)

compile: $(ODIR)
	@$(foreach svc, $(VAR_SERVICES), \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(ODIR)/$(svc)/$(svc) app/$(svc)/main.go;)

build:
	@$(foreach svc, $(VAR_SERVICES), \
		docker build -t $(IMAGE):$(VERSION) -f ./deploy/$(svc)/Dockerfile .;)

push:
	@$(foreach svc, $(VAR_SERVICES), \
		docker push $(IMAGE):$(VERSION);)

checkenv:
ifndef ENV
	$(error ENV must be set.)
endif

deploy: checkenv $(ODIR)
	@$(foreach svc, $(VAR_SERVICES), \
		echo deploying "$(svc)" to environment "$(ENV)" && \
		! kubelize genfile --overwrite -c ./ -s $(svc) -e $(ENV) deploy/$(svc)/$(FILE).yml $(ODIR)/$(svc)/ || \
		kubectl replace -f $(ODIR)/$(svc)/$(FILE).yml || kubectl create -f $(ODIR)/$(svc)/$(FILE).yml ;)

# only generate files from services
kubefile: checkenv $(ODIR)
	$(foreach svc, $(VAR_SERVICES), \
		$(foreach f, $(shell ls deploy/$(svc)/*.yml), \
			kubelize genfile --overwrite -c ./ -s $(svc) -e $(ENV) $(f) $(ODIR)/$(svc)/;))

setup:
	docker run --rm -it --network host -v $(PWD)/db:/app/db -v $(PWD)/.env:/app/.env registry.bukalapak.io/sre/migration:0.0.1 db:create
	docker run --rm -it --network host -v $(PWD)/db:/app/db -v $(PWD)/.env:/app/.env registry.bukalapak.io/sre/migration:0.0.1 db:migrate

migrate:
	docker run --rm -it --network host -v $(PWD)/db:/app/db -v $(PWD)/.env:/app/.env registry.bukalapak.io/sre/migration:0.0.1 db:migrate

migrate-status:
	docker run --rm -it --network host -v $(PWD)/db:/app/db -v $(PWD)/.env:/app/.env registry.bukalapak.io/sre/migration:0.0.1 db:migrate:status

migrate-up:
	docker run --rm -it --network host -v $(PWD)/db:/app/db -v $(PWD)/.env:/app/.env registry.bukalapak.io/sre/migration:0.0.1 db:migrate:up VERSION=$(VERSION)

migrate-down:
	docker run --rm -it --network host -v $(PWD)/db:/app/db -v $(PWD)/.env:/app/.env registry.bukalapak.io/sre/migration:0.0.1 db:migrate:down VERSION=$(VERSION)

migrate-test:
	go build -o test-migration scripts/test-migration/main.go && ./test-migration
