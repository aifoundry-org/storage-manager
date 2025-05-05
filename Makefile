VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/aifoundry-org/storage-manager/cmd.Version=$(VERSION) -X github.com/aifoundry-org/storage-manager/cmd.Commit=$(COMMIT) -X github.com/aifoundry-org/storage-manager/cmd.BuildDate=$(BUILD_DATE)"

all: build

BINDIR = bin
STORAGE_MANAGER = $(BINDIR)/storage-manager
IMAGE ?= aifoundryorg/storage-manager

build: $(STORAGE_MANAGER)

$(STORAGE_MANAGER): $(BINDIR)
	go build $(LDFLAGS) -o $@ .

$(BINDIR):
	@mkdir -p $@

image:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(IMAGE) .

