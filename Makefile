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

_deps-release: # install tools needed for release, conditionally
ifneq ("$(wildcard Dockerfile))","")
	go get github.com/aktau/github-release
endif
ifneq ("$(wildcard autotag))","")
	curl -L https://github.com/pantheon-systems/autotag/releases/download/v0.0.4/autotag.linux.x86_64 -o ~/bin/autotag
endif
