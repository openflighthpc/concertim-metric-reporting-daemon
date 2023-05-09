EXE = ct-metric-reporting-daemon
EXTRA_FILES = config/config.prod.yml LICENSE.txt README.md libexec/device-name-to-data_source_map.rb libexec/results-to-memcache.rb libexec/phoenix-cache-locking.rb
TARFILE = $(EXE).tgz

VERSION = $(shell git describe --tags --dirty --broken)

.PHONY: all
all: $(TARFILE)

.PHONY: $(EXE)
$(EXE):
	go build  -v -ldflags="-X 'main.version=$(VERSION)'" -o $(EXE) ./cmd/reporting/

$(TARFILE): $(EXE) $(EXTRA_FILES)
	tar czf $(TARFILE) $(EXE) $(EXTRA_FILES) \
		--transform "s/config.prod.yml/config.yml/" \
		--transform "s/^/$(EXE)\//"

.PHONY: package
package: $(TARFILE)

.PHONY: clean
clean:
	rm -f $(EXE) $(TARFILE)
