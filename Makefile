build_server:
	GO_ENABLED=0 GOOS=linux go build -o emojify-api

docker_server: build_server
	docker build -t docker.io/nicholasjackson/emojify-api .

run_machinebox:
	docker run -p 127.0.0.1:8080:8080 -e "MB_KEY=${MB_KEY}" machinebox/facebox                                                                        
run_machinebox_connect:
	consul connect proxy \
        -service machinebox \
        -service-addr 127.0.0.1:8080 \
        -listen ':8443'

run_consul:
	consul agent -config-file ./consul.hcl -config-format hcl
