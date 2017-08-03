
VERSION ?= v006

.PHONY: build
build:
	docker build -t oremj/webpush-simulator:$(VERSION) .



.PHONY: release
release: build
	docker push oremj/webpush-simulator:$(VERSION)
