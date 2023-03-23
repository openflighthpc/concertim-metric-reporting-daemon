BIN_NAME = metric-reporting-daemon
EXTRA_FILES = config/config.prod.yml LICENSE.txt README.md libexec/device-name-to-data_source_map.rb
TARFILE = $(BIN_NAME).tgz

build:
	go build -o $(BIN_NAME)

package: build $(EXTRA_FILES)
	tar czf $(TARFILE) $(BIN_NAME) $(EXTRA_FILES) \
		--transform "s/config.prod.yml/config.yml/" \
		--transform "s/^/$(BIN_NAME)\//"

clean:
	rm -f $(BIN_NAME) $(TARFILE)
