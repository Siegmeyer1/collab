build-server:
	cd cmd/server && go build .

clean-server:
	cd cmd/server && rm ./server

run-server:
	cd cmd/server && ./server

server: clean-server build-server run-server

