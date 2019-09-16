serve_local_dev:
	clear \
	&& go run *.go -local -host="" -production=false

serve_local_watch_dev:
	clear \
	&& spy go run *.go -local -host="" -production=false

deployment_dev:
	docker build --rm -f "Dockerfile" -t byrdapp/gopro:latest . \
	&& docker push byrdapp/gopro
