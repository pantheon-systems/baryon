build:: ## build project for current arch
	go build

build-linux:: _fetch-cert ## build project for linux
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w"

# if there is a docker file then set the docker variable so things can trigger off it
build-docker:: ## build the docker container if the dockerfile exists
ifeq ("$(wildcard $(CURDIR)/Dockerfile)","")
	$(error "Docker task called, but no DOCKER variable set. Eitehr Dockerfile is missing or you didn't include common.")
else
	docker build -t $(IMAGE) .
endif

build-circle:: build-linux ## build project for linux. If you need docker you will have to invoke that with an extension

deps:: _gvt-install ## install dependencies for project assumes you have go binary installed
	find  ./vendor/* -maxdepth 0 -type d -exec rm -rf "{}" \;
	gvt rebuild

test:: ## run go tests
	go test -race -v $$(go list ./... | grep -v /vendor/)

test-circle:: test test-coveralls ## invoke test tasks for CI

deps-circle:: ## install Golang and pull dependencies in CI
	bash scripts/make/sh/install-go.sh
deps-circle:: deps

deps-coverage::
	go get github.com/pierrre/gotestcover
	go get github.com/mattn/goveralls

deps-status:: ## check status of deps with gostatus
ifeq (, $(shell which gostatus))
	go get -u github.com/shurcooL/gostatus
endif
	go list -f '{{join .Deps "\n"}}' . | gostatus -stdin -v

test-coveralls:: deps-coverage ## run coverage and report to coveralls
	gotestcover -v -race  -coverprofile=coverage.out $$(go list ./... | grep -v /vendor/)
	goveralls -repotoken $$COVERALLS_TOKEN -service=circleci -coverprofile=coverage.out

test-coverage-html:: deps-coverage ## output html coverage file
	go tool cover -html=coverage.html

_gvt-install::
	go get -u github.com/FiloSottile/gvt

_fetch-cert::
ifdef $(FETCH_CA_CERT)
	curl  https://curl.haxx.se/ca/cacert.pem -o ca-certificates.crt
endif
