BINARY = goromdb

ifeq ($(shell uname -s),Darwin)
MD5 = md5 -r
else
MD5 = md5sum
endif

DB_FILES = \
	sample-data.json sample-bdb.db sample-memcachedb-bdb.db sample-boltdb.db \
	sample-ns-data.json sample-ns-boltdb.db
DB_DIR = _data/store
DB_PATHS = $(addprefix $(DB_DIR)/,$(DB_FILES))
MD5_PATHS = $(foreach path,$(DB_PATHS),$(path).md5)

all:
	$(MAKE) -j 4 dep $(DB_DIR)
	$(MAKE) -j 4 $(DB_PATHS) $(MD5_PATHS) $(BINARY)

dep:
	which dep || go get -u -v github.com/golang/dep/cmd/dep
	dep ensure -v

test:
	go vet ./...
	go test ./...

lint:
	(golint ./... | grep -v '^vendor/') || true

$(DB_DIR):
	mkdir -p $@

$(DB_DIR)/%.md5: $(DB_DIR)/%
	$(MD5) $< > $@

$(DB_DIR)/sample-data.json: _data/sample-data.json
	cp $< $@

$(DB_DIR)/sample-bdb.db: _data/sample-data.json
	go run ./cmd/sample-data/bdb/bdb.go -input-from $< -output-to $@

$(DB_DIR)/sample-memcachedb-bdb.db: _data/sample-data.json
	go run ./cmd/sample-data/memcachedb-bdb/memcachedb-bdb.go -input-from $< -output-to $@

$(DB_DIR)/sample-boltdb.db: _data/sample-data.json
	go run ./cmd/sample-data/boltdb/boltdb.go -input-from $< -output-to $@

$(DB_DIR)/sample-ns-data.json: _data/sample-ns-data.json
	cp $< $@

$(DB_DIR)/sample-ns-boltdb.db: _data/sample-ns-data.json
	go run ./cmd/sample-data/nsboltdb/nsboltdb.go -input-from $< -output-to $@

bench:
	go test -bench .

$(BINARY):
	go build

clean:
	go clean -testcache
	rm -rf $(BINARY) $(DB_PATHS) $(MD5_PATHS)

realclean: clean
	rm -rf vendor

.PHONY: dep test bench clean realclean
