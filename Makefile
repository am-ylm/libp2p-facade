tests:
	go clean -testcache
	go test -v -timeout 30s ./...

docker-tests:
	docker build -t libp2p-facade-tests -f Dockerfile.tests .
	docker run -it --rm libp2p-facade-tests
