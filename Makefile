EXE = ct-metric-reporting-daemon
EXTRA_FILES = config/config.prod.yml LICENSE.txt README.md libexec/device-name-to-data_source_map.rb libexec/results-to-memcache.rb libexec/phoenix-cache-locking.rb
TARFILE = $(EXE).tgz

.PHONY: all
all: $(TARFILE)

.PHONY: $(EXE)
$(EXE):
	go build  -o $(EXE) ./cmd/reporting/

$(TARFILE): $(EXE) $(EXTRA_FILES)
	tar czf $(TARFILE) $(EXE) $(EXTRA_FILES) \
		--transform "s/config.prod.yml/config.yml/" \
		--transform "s/^/$(EXE)\//"

.PHONY: clean
clean:
	rm -f $(EXE) $(TARFILE)
