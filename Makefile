REPORTING_EXE = ct-metric-reporting-daemon
EXTRA_FILES = config/config.prod.yml LICENSE.txt README.md libexec/device-name-to-data_source_map.rb
REPORTING_TARFILE = $(REPORTING_EXE).tgz

PROCESSING_EXE = ct-metric-processing-daemon
PROCESSING_EXTRA_FILES = processing/config/config.prod.yml LICENSE.txt README.md libexec/device-name-to-data_source_map.rb libexec/results-to-memcache.rb libexec/phoenix-cache-locking.rb
PROCESSING_TARFILE = $(PROCESSING_EXE).tgz

.PHONY: all
all: $(REPORTING_TARFILE) $(PROCESSING_TARFILE)

.PHONY: $(REPORTING_EXE)
$(REPORTING_EXE):
	go build  -o $(REPORTING_EXE) ./cmd/reporting/

.PHONY: $(PROCESSING_EXE)
$(PROCESSING_EXE):
	go build  -o $(PROCESSING_EXE) ./cmd/processing/


$(REPORTING_TARFILE): $(REPORTING_EXE) $(EXTRA_FILES)
	tar czf $(REPORTING_TARFILE) $(REPORTING_EXE) $(EXTRA_FILES) \
		--transform "s/config.prod.yml/config.yml/" \
		--transform "s/^/$(REPORTING_EXE)\//"

$(PROCESSING_TARFILE): $(PROCESSING_EXE) $(PROCESSING_EXTRA_FILES)
	tar czf $(PROCESSING_TARFILE) $(PROCESSING_EXE) $(PROCESSING_EXTRA_FILES) \
		--transform "s/config.prod.yml/config.yml/" \
		--transform "s/^processing\///" \
		--transform "s/^/$(PROCESSING_EXE)\//"

.PHONY: clean
clean:
	rm -f $(REPORTING_EXE) $(REPORTING_TARFILE)
	rm -f $(PROCESSING_EXE) $(PROCESSING_TARFILE)
