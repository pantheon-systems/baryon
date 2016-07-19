# Docker common things
#
# INPUT VARIABLES
# 	- QUAY_USER: The quay.io user to use (usually set in CI)
# 	- QUAY_PASSWD: The quay passwd to use  (usually set in CI)
# 	- IMAGE: the docker image to use. will be computed if it doesn't exist.
# 	- REGISTRY: The docker registry to use. defaults to quay.
#
# EXPORT VARIABLES
# 	- BUILD_NUM: The build number for this build. will use 'dev' if not building
# 	             on circleCI, will use CIRCLE_BUILD_NUM otherwise.
# 	- IMAGE: The image to use for the build. 
# 	- REGISTRY: The registry to use for the build.
#
#-------------------------------------------------------------------------------
ifeq ($(CIRCLE_BUILD_NUM),)
  BUILD_NUM := dev
else
  BUILD_NUM := $(CIRCLE_BUILD_NUM)
  QUAY := docker login -p "$$QUAY_PASSWD" -u "$$QUAY_USER" -e "unused@unused" quay.io
endif

# These can be overridden
IMAGE ?= $(REGISTRY)/$(APP):$(BUILD_NUM)
REGISTRY ?= quay.io/getpantheon

# if there is a docker file then set the docker variable so things can trigger off it
ifneq ("$(wildcard Dockerfile))","")
  # file is there
  DOCKER:=true
endif

# determinse the docker tag to build
build-docker::
ifndef  DOCKER
	$(error "Docker task called, but no DOCKER variable set. Eitehr Dockerfile is missing or you didn't include common.")
endif
build-docker:: build-linux ## build the docker container 
	docker build -t $(IMAGE) .

push:: ## push the container to the registry
	docker push $(IMAGE)

setup-quay:: ## setup docker login for quay.io
ifndef QUAY_PASSWD
	$(error "Need to set QUAY_PASSWD environment variable")
endif
ifndef QUAY_USER
	$(error "Need to set QUAY_USER environment variable")
endif
	@$(QUAY)

# we call make here to ensure new states are detected
push-circle:: ## build and push the container from circle
	make build-docker
push-circle:: setup-quay
	make push

.PHONY:: setup-quay build-docker push push-circle
