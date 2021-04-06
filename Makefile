# Config
# Change the default config with `make cnf="local.env" build`
cnf ?= .env

include $(cnf)
export $(shell sed 's/=.*//' $(cnf))

.PHONY: help

# https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# Docker Tasks
# Build the container
build: ## Build the container
	docker build -t $(APP_NAME) -f Dockerfile .

tag-latest: ## Generate container {version} tag
	@echo 'Tag container with latest'
	docker tag $(APP_NAME) $(AWS_CONTAINER_REPO)/$(APP_NAME):latest

publish-latest: tag-latest ## Publish the latest tagged container to ECR
	@echo 'Publish latest $(APP_NAME) to $(AWS_CONTAINER_REPO)'
	docker push $(AWS_CONTAINER_REPO)/$(APP_NAME):latest
