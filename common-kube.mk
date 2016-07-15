REGISTRY ?= quay.io/getpantheon

# determinse the docker tag to build
ifeq ($(CIRCLE_BUILD_NUM),)
  BUILD_NUM := dev
else
  BUILD_NUM := $(CIRCLE_BUILD_NUM)
  QUAY := docker login -p "$$QUAY_PASSWD" -u "$$QUAY_USER" -e "unused@unused" quay.io
endif

# If we have no circle branch, use development kube env
# If we are on master branch, use production kube env
# If we are NOT on master use testing kube env
ifeq ($(CIRCLE_BRANCH),) # Dev
  KUBE_ENV := development
else  ifeq ($(CIRCLE_BRANCH), master) # prod
  KUBE_ENV := production
else # testing
  KUBE_ENV := testing
endif

# These can be overridden
IMAGE ?= $(REGISTRY)/$(APP):$(BUILD_NUM)
SERVCE ?= $(APP)

# debatable weather this should be in common or not, but I see it needed enough in dev.
# TODO(jesse): possibly guard this to prevent accidentally nuking production.
force-pod-restart:: ## nuke the pod
	kubectl --namespace=$(KUBE_ENV) get  pod -l"app=$(APP)" --no-headers | awk '{print $$1}' | xargs kubectl delete pod

push:: ## push the container to the registry
	docker push $(IMAGE)

setup-quay:: ## setup docker login for quay.io
ifndef $(QUAY_PASSWD) 
	$(error "Need to set QUAY_PASSWD environment variable")
endif
ifndef $(QUAY_USER)
	$(error "Need to set QUAY_USER environment variable")
endif
	$(QUAY)

# we call make here to ensure new states are detected
push-circle:: ## build and push the container from circle
	make build-docker
psuh-circle:: setup-quay
	make push

# extend or define circle deps to install gcloud
deps-circle:: 
	@bash scripts/make/sh/install-gcloud.sh
