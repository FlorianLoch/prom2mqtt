.PHONY: image build run run-container clean push-image test

bin_dir := ./bin
bin := ./bin/prom2mqtt
image_built := ./.make/image_built
image_pushed := ./.make/image_pushed
make_dir := ./.make
go_files := $(shell find . -name "*.go")
image_tag := prom2mqtt

build: $(bin)


run: $(bin)
	$(bin)


image: $(image_built)


run-container: $(image_built)
	docker run -v $(shell pwd)/prom2mqtt.config.yaml:/app/prom2mqtt.config.yaml $(image_tag)


clean:
	rm -rf $(bin_dir)
	rm -rf $(make_dir)


push-image: $(image_pushed)


test:
	go test ./...


$(bin_dir):
	mkdir -p $(bin_dir)


$(make_dir):
	mkdir -p $(make_dir)


$(image_pushed): $(image_built)
	podman save $(image_tag) | pv | ssh raspi docker load
	touch $(image_pushed)


$(bin): $(go_files) | $(bin_dir)
	go generate ./...
	CGO_ENABLED=1 go build -o $(bin)


$(image_built): $(go_files) Dockerfile $(raspi_bin) $(make_dir)
	podman build --platform linux/arm64 --tag $(image_tag) .
	touch $(image_built)
