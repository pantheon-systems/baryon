# common make tasks and variables that should be imported into all projects
# 
#-------------------------------------------------------------------------------
help: ## print list of tasks and descriptions
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?##"}; { split($$0,a,":"); printf "\033[36m%-30s\033[0m %s \n", a[2], $$2}'
.DEFAULT_GOAL := help

update-makefiles: ## update the make subtree, assumes the subtree is in scripts/make
	git subtree pull --prefix scripts/make common_makefiles master --squash	

.PHONY: all
