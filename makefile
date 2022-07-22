NAME=Clash.Meta_For_Magisk
BINDIR=bin
VERSION=$(shell git rev-parse --abbrev-ref HEAD)-$(shell git rev-parse --short HEAD)
BUILDTIME=$(shell date -u)
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags '-X "Clash.Meta_For_Magisk/constant.Version=$(VERSION)" \
		-X "Clash.Meta_For_Magisk/constant.BuildTime=$(BUILDTIME)" \
		-w -s -buildid='

all: android-arm64 linux-amd64v3

android-arm64:
	GOARCH=arm64 GOOS=android $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-amd64v3:
	GOARCH=amd64 GOOS=linux GOAMD64=v3 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

vet:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm $(BINDIR)/*