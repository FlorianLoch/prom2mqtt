.PHONY: image build build-raspi run run-container clean push-image test

bin_dir := ./bin
bin := ./bin/prom2mqtt
raspi_bin := ./bin/prom2mqtt-arm32v7
image_built := ./bin/image_built
image_pushed := ./bin/image_pushed
go_files := $(shell find . -name "*.go")
image_tag := prom2mqtt

build: $(bin)


build-raspi: $(raspi_bin)


run: $(bin)
	$(bin)


image: $(image_built)


run-container: $(image_built)
	docker run -v $(shell pwd)/prom2mqtt.config.yaml:/app/prom2mqtt.config.yaml $(image_tag)


clean:
	rm -rf $(bin_dir)


push-image: $(image_pushed)


test:
	go test ./...


$(bin_dir):
	mkdir -p $(bin_dir)


$(image_pushed): $(image_built)
	docker save $(image_tag) | pv | ssh raspi docker load
	touch $(image_pushed)


$(bin): $(go_files) | $(bin_dir)
	CGO_ENABLED=1 go build -o $(bin)


$(raspi_bin): $(go_files) | $(bin_dir)
	# '-ldflags "-w -s"' is used to shrink the binary image by stripping debug information
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-w -s" -o $(raspi_bin) .


$(image_built): $(go_files) Dockerfile $(raspi_bin)
	docker build --platform linux/arm/v7 --tag $(image_tag) .
	touch $(image_built)