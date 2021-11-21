INSTALLDIR := /home/roy/Tmp/quic-tools

build-client:
	go build -o ./bin/quic-client enac/client/client.go

build-server:
	go build -o ./bin/quic-server enac/server/server.go

build: build-client build-server

clean:
	rm -rf bin

install: clean build
	@echo "installing client in '${INSTALLDIR}/client' directory"
	cp bin/quic-client ${INSTALLDIR}/client
	@echo "installing server in '${INSTALLDIR}/server' directory"
	cp bin/quic-server ${INSTALLDIR}/server

.PHONY: build clean install
