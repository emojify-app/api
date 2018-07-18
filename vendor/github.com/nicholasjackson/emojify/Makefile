build:
	GO_ENABLED=0 GOOS=linux go build -o emojifyServer ./server/.

docker: build
	docker build -t docker.io/nicholasjackson/emojify .
