serve_local_dev:
	clear \
	&& go run cmd/byrd-pro-api/main.go -local -production=false

watch_serve_local:
	clear \
	&& spy go run cmd/byrd-pro-api/main.go -local -production=false

deployment_dev:
	docker build --rm -f "Dockerfile" -t byrdapp/byrd-pro-api:dev . \
	&& docker push byrdapp/byrd-pro-api

build_docker_tag:
	echo "building pro api with tag: ${tag}" \
	&& docker build --rm -f "Dockerfile" -t byrdapp/byrd-pro-api:${tag} . \
	&& docker push byrdapp/byrd-pro-api:${tag}

eval_dev_manager:
	eval $(docker-machine env manager1)
