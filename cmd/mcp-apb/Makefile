DOCKERORG = feedhenry
USER=$(shell id -u)
PWS=$(shell pwd)
build_and_push: apb_build docker_push

.PHONY: apb_build
apb_build:
	docker run --rm -u $(USER) -v $(PWD):/mnt:z feedhenry/apb prepare
	docker build -t $(DOCKERORG)/mcp-apb .

.PHONY: docker_push
docker_push:
	docker push $(DOCKERORG)/mcp-apb
