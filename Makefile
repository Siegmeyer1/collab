build-server:
	cd cmd/server && go build .

clean-server:
	cd cmd/server && rm ./server

run-server:
	cd cmd/server && ./server -p 1235

server: clean-server build-server run-server

env:
	docker-compose up --detach

run: env run-server

run-multihost:
	cd cmd/server && (trap 'kill 0' SIGINT; ./server -p 1234 & ./server -p 1235 & ./server -p 1236 & ./server -p 1237)
