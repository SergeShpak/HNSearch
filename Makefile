all: build-indexer build-server

build-indexer:
	cd ./indexer && $(MAKE)

build-server:
	cd ./server && $(MAKE)

clean:
	cd ./indexer && $(MAKE) clean
	cd ./server && $(MAKE) clean