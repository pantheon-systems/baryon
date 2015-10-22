GO15VENDOREXPERIMENT=1
GOOGLE_PROJECT=pantheon-internal
APP=baryon

all: deps test build

deps: gvt_install
deps:
	gvt rebuild

test:
	go test

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

release:
	 GOOS=linux go build -o baryon-x86_64
	 GOOS=darwin go build -o baryon-darwin
	 GOOS=windows go build

refresh_deps: gvt_install
refresh_deps:
	bash deploy/refresh.sh

gvt_install:
	go get -u github.com/FiloSottile/gvt

# gce task builds the docker container, and pushes it to gce container repo.
# no reason this can't be quay.io
#  We pull down root certs, cause the docker image doesn't have them.
gce: docker
gce:
	gcloud docker push gcr.io/$(GOOGLE_PROJECT)/$(APP)

docker: cert build_linux
docker:
	docker build -t gcr.io/$(GOOGLE_PROJECT)/$(APP)  .

.PHONY: all
