build_server:
	GO_ENABLED=0 GOOS=linux go build -o emojify-api

docker_server: build_server
	docker build -t nicholasjackson/emojify-api .

build_and_push: docker_server
	docker push nicholasjackson/emojify-api:latest

run_machinebox_connect:
	consul connect proxy \
        -service machinebox \
        -service-addr 127.0.0.1:8080 \
        -listen ':8443'

run_consul:
	consul agent -config-file ./consul.hcl -config-format hcl

run_test_functional:
	docker run --rm -p 9090:9090 -p 9091:9091 -e BIND_ADDRESS=0.0.0.0 nicholasjackson/emojify-cache:v0.3.4

run_cache:
	docker run --rm -p 9001:9090 -e BIND_ADDRESS=0.0.0.0 nicholasjackson/emojify-cache:v0.4.3

run_facedetect:
	docker run --rm -p 9002:9090 -e BIND_ADDRESS=0.0.0.0 nicholasjackson/emojify-facedetection:v0.1.2

run_api:
	go run server.go -authn-disable=true -allow-origin=* -facebox-address=localhost:9002 -cache-address=localhost:9001
