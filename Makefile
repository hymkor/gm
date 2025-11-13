ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=set
    DEL=del
    NUL=nul
    WHICH=where.exe
else
    SET=export
    DEL=rm
    NUL=/dev/null
    WHICH=which
endif

ifndef GO
    SUPPORTGO=go1.20.14
    GO:=$(shell $(WHICH) $(SUPPORTGO) 2>$(NUL) || echo go)
endif

NAME=$(notdir $(CURDIR))
VERSION=$(shell git describe --tags 2>$(NUL) || echo v0.0.0)
EXE=$(shell $(GO) env GOEXE)
GOOPT=-ldflags "-s -w -X main.version=$(VERSION)"
TARGET=$(NAME)$(EXE)

build : $(wildcard *.go) SKK-JISYO.L.bz2
	$(GO) fmt
	$(GO) build $(GOOPT)

SKK-JISYO.L.bz2:
	curl https://raw.githubusercontent.com/skk-dev/dict/master/SKK-JISYO.L | bzip2 > SKK-JISYO.L.bz2

_dist:
	$(SET) "CGO_ENABLED=0" && $(GO) build $(GOOPT)
	zip -9 $(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).zip $(NAME)$(EXE)

dist:
	$(SET) "GOOS=linux" && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=linux" && $(SET) "GOARCH=amd64" && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=amd64" && $(MAKE) _dist

clean:
	rm *.zip gm gm$(EXE)

manifest:
	make-scoop-manifest *-windows-*.zip > $(NAME).json

release:
	gh release create -d --notes "" -t $(VERSION) $(VERSION) $(wildcard $(NAME)-$(VERSION)-*.zip)

.PHONY: all test dist _dist clean manifest release
