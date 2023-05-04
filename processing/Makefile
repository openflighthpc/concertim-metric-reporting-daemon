BIN_NAME = ct-metric-processing-daemon
EXTRA_FILES = config/config.prod.yml LICENSE.txt README.md libexec/dsm-to-memcache-key.rb libexec/results-to-memcache.rb libexec/phoenix-cache-locking.rb
TARFILE = $(BIN_NAME).tgz

.PHONY: all
all: $(TARFILE)

.PHONY: $(BIN_NAME)
$(BIN_NAME):
	go build  -o $(BIN_NAME) ../cmd/processing/

$(TARFILE): $(BIN_NAME) $(EXTRA_FILES)
	tar czf $(TARFILE) $(BIN_NAME) $(EXTRA_FILES) \
		--transform "s/config.prod.yml/config.yml/" \
		--transform "s/^/$(BIN_NAME)\//"

clean:
	rm -f $(BIN_NAME) $(TARFILE)
