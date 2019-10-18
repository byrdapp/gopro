serve_local_dev:
	clear \
	&& go run cmd/gopro/gopro.go -local -production=false

watch_serve_local:
	clear \
	&& spy go run cmd/gopro/gopro.go -local -production=false

deployment_dev:
	docker build --rm -f "Dockerfile" -t byrdapp/gopro:latest . \
	&& docker push byrdapp/gopro

build_docker_tag:
	echo "building pro api with tag: ${tag}" \
	&& docker build --rm -f "Dockerfile" -t byrdapp/gopro:${tag} . \
	&& docker push byrdapp/gopro:${tag}