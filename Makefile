test:
	go test ./... -v

build-docker:
	docker build . -t rlimit

run-docker:
	docker run -p 4000:4000 rlimit 