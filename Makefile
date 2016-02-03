GO15VENDOREXPERIMENT=1
APP=baryon

all: deps test build

list:
	@make -rqp | awk -F':' '/^[a-zA-Z0-9][^$$#\/\t=]*:([^=]|$$)/ {split($$1,A,/ /);for(i in A)print A[i]}' | sort | uniq

deps: gvt_install
	gvt rebuild

test:
	go test

cov:
	go get github.com/pierrre/gotestcover
	go get github.com/mattn/goveralls
	gotestcover -coverprofile=coverage.out $$(go list ./... | grep -v /vendor/)

coveralls: cov
	goveralls -repotoken $$COVERALLS_TOKEN -service=circleci -coverprofile=coverage.out

cov_html: cov
	go tool cover -html=coverage.out

build:
	go build

build_linux:
	GOOS=linux go build

deploy:
	kubectl rolling-update $(APP)  --poll-interval="500ms" --image=gcr.io/$(GOOGLE_PROJECT)/$(APP)

force_pod_restart:
	kubectl get  pod -l"app=baryon" --no-headers | awk '{print $$1}' | xargs kubectl delete pod

cert:
	curl https://raw.githubusercontent.com/bagder/ca-bundle/master/ca-bundle.crt -o ca-certificates.crt

circle_deps:
	bash deploy/gce/gcloud-setup.sh
	bash deploy/install-go.sh

update_secrets:
	kubectl replace -f deploy/gce/baryon-ssl.yml
	kubectl replace -f deploy/gce/baryon-secrets.yml

update_rc:
	kubectl get rc/baryon && kubectl replace -f deploy/gce/baryon-rc.yml ||  kubectl create -f deploy/gce/baryon-rc.yml

release_deps:
	go get github.com/aktau/github-release
	go get -u github.com/pantheon-systems/autotag

release: release_deps
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


refresh_deps: gvt_install
	bash deploy/refresh.sh

gvt_install:
	go get -u github.com/FiloSottile/gvt

# gce task builds the docker container, and pushes it to gce container repo.
# no reason this can't be quay.io
#  We pull down root certs, cause the docker image doesn't have them.
gce: docker
	gcloud docker push gcr.io/$(GOOGLE_PROJECT)/$(APP)

docker: cert build_linux
	docker build -t gcr.io/$(GOOGLE_PROJECT)/$(APP)  .

.PHONY: all
