# Common kube things. This is the simplest set of common kube tasks
#
# INPUT VARIABLES
# 	- APP: should be defined in your topmost Makefile
#
# EXPORT VARIABLES
# 	- KUBE_NAMESPACE: represents the kube namespace that has been detected based on 
# 	           branch build and circle existence.
#-------------------------------------------------------------------------------

# If we have no circle branch, use development kube env
# If we are on master branch, use production kube env
# If we are NOT on master use testing kube env
ifeq ($(CIRCLE_BRANCH),) # Dev
  KUBE_NAMESPACE := development
else  ifeq ($(CIRCLE_BRANCH), master) # prod
  KUBE_NAMESPACE := production
else # testing
  KUBE_NAMESPACE := testing
endif

# debatable weather this should be in common or not, but I see it needed enough in dev.
# TODO(jesse): possibly guard this to prevent accidentally nuking production.
force-pod-restart:: ## nuke the pod
	kubectl --namespace=$(KUBE_NAMESPACE) get  pod -l"app=$(APP)" --no-headers | awk '{print $$1}' | xargs kubectl delete pod

# extend or define circle deps to install gcloud
deps-circle:: 
	@bash scripts/make/sh/install-gcloud.sh
