build_server:
	GO_ENABLED=0 GOOS=linux go build -o emojify-api

docker_server: build_server
	docker build -t nicholasjackson/emojify-api .

build_and_push: docker_server
	docker push nicholasjackson/emojify-api:latest

run_consul:
	consul agent -config-file ./consul.hcl -config-format hcl

run_test_functional:
	docker run --rm -p 9090:9090 -p 9091:9091 -e BIND_ADDRESS=0.0.0.0 nicholasjackson/emojify-cache:v0.3.4

goconvey:
	goconvey -excludedDirs=dist,images
