
default: build

build:
	#CGO_ENABLED=1 GOOS=linux go build -ldflags "-s -w" -o read2succeed
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=/usr/local/bin/x86_64-linux-musl-cc \
				go build -v -a -ldflags '-linkmode external -extldflags "-static"' -o read2succeed

docker:
	docker build --no-cache -t r2s:latest -f Dockerfile .

run-container:
	docker run --rm \
		-e BIND_ADDRESS="0.0.0.0:8080" \
		-e SESSION_KEY="4666fe1e1f3418b8caa6b3335ca0952f8986f03edb571b1ba4a54518d0f8b75d64" \
		-e CSRF_KEY="x8b8830b3b023fbbdaf52dd80c9bzzx11" \
		-e DB_PATH="/app/var/db.db" \
		-e TZ="America/New_York" \
		-p 8080:8080 \
		-v ${PWD}/var/:/app/var/:rw \
		--name r2s-docker r2s:latest



