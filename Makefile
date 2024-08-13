build:
	@go build -o bin/api

run:build
	@./bin/api

test:
	@go test ./...

upload_github:
	@git add ./
	@git commit -m "Gallery"
	@git push -u origin main