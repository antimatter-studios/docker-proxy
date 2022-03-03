.SILENT :
.PHONY : test test

build-webserver:
	docker build -t web test/requirements/web

build-ddt-proxy-test:
	docker build -t antimatter-studios/docker-dev-tools-proxy:latest .

test: build-webserver build-ddt-proxy-test
	test/pytest.sh

test: test-alpine
