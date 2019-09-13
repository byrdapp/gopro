serve_local:
	clear \
	&& go run *.go -local -host=""

serve_local_watch:
	clear \
	&& spy go run *.go -local -host=""

serve_docker_local:
	docker-compose up --build

deployment_dev:
	docker build --rm -f "Dockerfile" -t byrdapp/gopro:latest . \
	&& docker push byrdapp/gopro
