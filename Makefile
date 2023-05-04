BIN_NAME = ct-metric-reporting-daemon
EXTRA_FILES = config/config.prod.yml LICENSE.txt README.md libexec/device-name-to-data_source_map.rb
TARFILE = $(BIN_NAME).tgz

.PHONY: all
all: $(TARFILE)

.PHONY: $(BIN_NAME)
$(BIN_NAME):
	go build  -o $(BIN_NAME) ./cmd/reporting/

$(TARFILE): $(BIN_NAME) $(EXTRA_FILES)
	tar czf $(TARFILE) $(BIN_NAME) $(EXTRA_FILES) \
		--transform "s/config.prod.yml/config.yml/" \
		--transform "s/^/$(BIN_NAME)\//"

.PHONY: clean
clean:
	rm -f $(BIN_NAME) $(TARFILE)
