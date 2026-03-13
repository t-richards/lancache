all:
	export GOFLAGS="-trimpath -mod=readonly -modcacherw"; \
	export CGO_ENABLED=0; \
	export GOAMD64=v3; \
	go build -o bin/lancache
	sudo setcap 'cap_net_bind_service=+ep' bin/lancache

.PHONY: clean
clean:
	rm -f bin/lancache
