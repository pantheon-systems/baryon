APP=baryon

include scripts/make/common.mk
include scripts/make/common-go.mk
include scripts/make/common-docker.mk

release: _deps-release ## run a release (usually from CI)
release: TAG=$(shell autotag -n)
release: VERSION=$(TAG) push-circle
#release:
#	@echo "Building release for $(TAG)"
#	autotag
#	GOOS=linux go build -o baryon-linux
#	GOOS=darwin go build -o baryon-darwin
#	GOOS=windows go build
#	github-release release -u pantheon-systems -r baryon -t $(TAG) --draft
#	github-release upload -u pantheon-systems -r baryon -n Linux -f baryon-linux -t $(TAG)
#	github-release upload -u pantheon-systems -r baryon -n OSX -f baryon-darwin -t $(TAG)
#	github-release upload -u pantheon-systems -r baryon -n Windows -f baryon.exe -t $(TAG)

_deps-release: # install tools needed for release
	go get github.com/aktau/github-release
	go get -u github.com/pantheon-systems/autotag
