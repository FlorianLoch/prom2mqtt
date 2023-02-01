# prom2mqtt

`prom2mqtt` is a small daemon service doing what its same implies: it scrapes a defined target exporting metrics in Prometheus's text format and pushes the values, accordingly to the configuration, to an MQTT broker.

My personal use case for this is a Go application reading measurements from a weather station connected via USB, that is exporting Prometheus metrics already (as it gets scraped by an actual Prometheus server).
Instead of integrating MQTT support into this service it, data gets bridged via `prom2mqtt`.

Possible future additions could be:
- Support for Prometheus other export formats
- Support for multiple scraping targets
- Support for *Basic Authentication* when scraping


## Preparations for building the image

Currently, I could not manage to build the container image for the Raspberry 3 (`arm32v7`) on my macBook (`aarch64`) using `podman` and `podman machine`.

Reason for this seems to be the lack of the right `qemu` packages in the `podman machines`'s CoreOS image.

Actually, adding some layers should do the trick but didn't work for me:

```bash
podman machine ssh
# Then, once logged into the machine, simply install packages for all platforms
sudo rpm-ostree install qemu-user-static
sudo rpm-ostree install qemu-user-binfmt
```

Also see https://github.com/containers/podman/issues/17267#issuecomment-1409779092.


Fallback currently is to use `Docker for macOS` which just works perfectly fine.
The built image can then be pushed and deployed to the Raspberry.

