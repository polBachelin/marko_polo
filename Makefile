BINARY = marko
PREFIX = /usr/local/bin

.PHONY: build install uninstall clean

build:
	go build -ldflags "-s -w" -o $(BINARY)

install: build
	cp $(BINARY) $(PREFIX)/$(BINARY)

uninstall:
	rm -f $(PREFIX)/$(BINARY)

clean:
	rm -f $(BINARY)
