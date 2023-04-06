BIN_NAME = ct-metric-processing-daemon
EXTRA_FILES = config/config.prod.yml LICENSE.txt README.md libexec/dsm-to-memcache-key.rb libexec/results-to-memcache.rb
TARFILE = $(BIN_NAME).tgz

.PHONY: all
all: package

build:
	go build -o $(BIN_NAME)

package: build $(EXTRA_FILES)
	tar czf $(TARFILE) $(BIN_NAME) $(EXTRA_FILES) \
		--transform "s/config.prod.yml/config.yml/" \
		--transform "s/^/$(BIN_NAME)\//"

clean:
	rm -f $(BIN_NAME) $(TARFILE)
