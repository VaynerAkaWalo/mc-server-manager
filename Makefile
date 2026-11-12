build:
	@go build -o bin/app ./cmd

web:
	@npm --prefix web ci
	@npm --prefix web run build
	@rm -rf static
	@cp -r web/dist static

run: build
	@./bin/app

.PHONY: build web run
