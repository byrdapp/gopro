serve_local:
	go run *.go

serve_docker_local:
	docker-compose up --build

deployment_dev:
	docker build --rm -f "Dockerfile" -t byrdapp/basic-server:latest . \
	&& docker push byrdapp/basic-server