serve_local_dev:
	clear \
	&& go run cmd/gopro/main.go -local -production=false

serve_local_watch_dev:
	clear \
	&& spy go run *.go -local -host="" -production=false

deployment_dev:
	docker build --rm -f "Dockerfile" -t byrdapp/gopro:latest . \
	&& docker push byrdapp/gopro

build_docker_tag:
	echo "building pro api with tag: ${tag}" \
	&& docker build --rm -f "Dockerfile" -t byrdapp/gopro:${tag} . \
	&& docker push byrdapp/gopro:${tag}