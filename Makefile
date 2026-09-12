BINDIR ?= $(HOME)/.local/bin
GO ?= go
VHS ?= vhs

.PHONY: build test install dist release demo clean
build:
	$(GO) build -trimpath -o hitmaker .
test:
	$(GO) test -race ./...
	$(GO) vet ./...
install: build
	mkdir -p "$(BINDIR)"
	install -m755 hitmaker "$(BINDIR)/hitmaker"
dist:
	mkdir -p dist
	@set -e; for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		os=$${target%/*}; arch=$${target#*/}; ext=; \
		if [ "$$os" = windows ]; then ext=.exe; fi; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build -trimpath -o "dist/hitmaker-$$os-$$arch$$ext" .; \
	done
release:
	GO="$(GO)" python3 scripts/release.py "$(VERSION)"
demo: build
	$(VHS) docs/demo.tape
	test -s docs/screenshot.png && test -s docs/demo.gif && test -s docs/demo.mp4
clean:
	rm -f hitmaker
