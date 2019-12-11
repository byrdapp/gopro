serve_local_dev:
	clear \
	&& go run cmd/gopro/main.go -local -production=false

serve_local_dev_no_db:
	clear \
	&& go run cmd/gopro/main.go -local -production=false -db_active=false

watch_serve_local:
	clear \
	&& spy go run cmd/gopro/main.go -local -production=false

deployment_dev:
	docker build --rm -f "Dockerfile" -t byrdapp/gopro:dev . \
	&& docker push byrdapp/gopro

build_docker_tag:
	echo "building pro api with tag: ${tag}" \
	&& docker build --rm -f "Dockerfile" -t byrdapp/gopro:${tag} . \
	&& docker push byrdapp/gopro:${tag}

eval_dev_manager:
	eval $(docker-machine env manager1)
