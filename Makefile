APP=baryon
PROJECT := $$GOOGLE_PROJECT
SERVICE := "baryon"

include scripts/make/common.mk
include scripts/make/common-kube.mk
include scripts/make/common-go.mk

all: deps test build

update-secrets: ## Update the kube secrets
	kubectl replace -f deploy/gce/baryon-ssl.yml
	kubectl replace -f deploy/gce/baryon-secrets.yml

release: _deps-release ## run a release (usually from CI)
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

_deps-release: # install tools needed for release
	go get github.com/aktau/github-release
	go get -u github.com/pantheon-systems/autotag

update_deployment: ## update the kube Deployer
	@sed -e "s#__IMAGE__#$(IMAGE)#" \
	    -e "s/__BRANCH__/$(KUBE_ENV)/" \
	    scripts/deployment.template \
	    | kubectl apply --namespace=$(KUBE_ENV) -f -
