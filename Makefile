all: build

BIN_DIR = bin
STORAGE_MANAGER = $(BIN_DIR)/storage-manager

build: $(STORAGE_MANAGER)

$(STORAGE_MANAGER): $(BINDIR)
	go build -o $@ .

$(BINDIR):
	@mkdir -p $@

