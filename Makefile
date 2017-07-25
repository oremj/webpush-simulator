
.PHONY: build
build:
	docker build -t oremj/webpush-simulator:latest .



.PHONY: release
release: build
	docker push oremj/webpush-simulator:latest
