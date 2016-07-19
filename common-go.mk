# Common  Go Tasks
#
# INPUT VARIABLES
# 	- COVERALLS_TOKEN: Token to use when pushing coverage to coveralls.
#
# 	- FETCH_CA_CERT: The presence of this variable will add a  Pull root ca certs
# 	                 to  ca-certificats.crt before build.
#
#-------------------------------------------------------------------------------
build:: ## build project for current arch
	go build

build-linux:: _fetch-cert ## build project for linux
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w"

build-circle:: build-linux ## build project for linux. If you need docker you will have to invoke that with an extension

deps:: _gvt-install ## install dependencies for project assumes you have go binary installed
	find  ./vendor/* -maxdepth 0 -type d -exec rm -rf "{}" \;
	gvt rebuild

test:: ## run go tests
	go test -race -v $$(go list ./... | grep -v /vendor/)

test-circle:: test test-coveralls ## invoke test tasks for CI

deps-circle:: ## install Golang and pull dependencies in CI
	bash scripts/make/sh/install-go.sh

deps-coverage::
ifeq (, $(shell which gotestcover))
	go get github.com/pierrre/gotestcover
endif
ifeq (, $(shell which goveralls))
	go get github.com/mattn/goveralls
endif

deps-status:: ## check status of deps with gostatus
ifeq (, $(shell which gostatus))
	go get -u github.com/shurcooL/gostatus
endif
	go list -f '{{join .Deps "\n"}}' . | gostatus -stdin -v

test-coveralls:: deps-coverage ## run coverage and report to coveralls
ifdef COVERALLS_TOKEN
	gotestcover -v -race  -coverprofile=coverage.out $$(go list ./... | grep -v /vendor/)
	goveralls -repotoken $$COVERALLS_TOKEN -service=circleci -coverprofile=coverage.out
else
	$(error "You asked to use Coveralls, but neglected to set the COVERALLS_TOKEN environment variable")
endif

test-coverage-html:: deps-coverage ## output html coverage file
	go tool cover -html=coverage.html

_gvt-install::
ifeq (, $(shell which gvt))
	go get -u github.com/FiloSottile/gvt
endif

_fetch-cert::
ifdef FETCH_CA_CERT
	curl  https://curl.haxx.se/ca/cacert.pem -o ca-certificates.crt
endif
