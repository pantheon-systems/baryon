APP=baryon
PROJECT := $$GOOGLE_PROJECT

include scripts/common.mk
include scripts/common-kube.mk
include scripts/common-go.mk

all: deps test build

deps::
	@echo "deps!"

update_secrets: ## Update the kube secrets
	kubectl replace -f deploy/gce/baryon-ssl.yml
	kubectl replace -f deploy/gce/baryon-secrets.yml

release: release_deps ## run a release (usually from CI)
release: TAG=$(shell autotag -n)
release:
	 @echo "Building release for $(TAG)"
	 autotag
	 GOOS=linux go build -o baryon-linux
	 GOOS=darwin go build -o baryon-darwin
	 GOOS=windows go build
	 github-release release -u pantheon-systems -r baryon -t $(TAG) --draft
	 github-release upload -u pantheon-systems -r baryon -n Linux -f baryon-linux -t $(TAG)
	 github-release upload -u pantheon-systems -r baryon -n OSX -f baryon-darwin -t $(TAG)
	 github-release upload -u pantheon-systems -r baryon -n Windows -f baryon.exe -t $(TAG)
