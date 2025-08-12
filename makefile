OUT := agentur-logo-importer.exe
PKG := .
# This is not run initially.
#PFX_PWD:=$(shell eval d:/dev/tools/keepass.exe -key sectigoKey.pfx)
PKG_LIST := main


VERSION := $(shell git describe --always --long --dirty --tag)
#`date '+%FT%T%z'
VERSION_BUILD := $(shell date '+%FT%T%z')
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
GO_FILES_NOVERSION := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v version.go)
export GOOS=windows

LOCAL_PKG_DIR := $(shell eval pwd)

.PHONY: print_pwd
print_pwd:
	@echo $(PFX_PWD)
	

all: run

version.go: get_version.sh ${GO_FILES_NOVERSION}
	## Build version.go
	bash ./get_version.sh
	go generate

## test in amd64 does not allow resource i386, create empty object file
empty.ii:
	echo "/* empty obj file */"> empty.ii
	
resource_amd64.syso: empty.ii
	gcc -c empty.ii
	mv empty.o resource_amd64.syso

resource_i386.syso: Bokehlicia-Captiva-Software-upload.ico versioninfo.json version.go
	## build  resource_i386.syso
	go generate
	mv resource.syso resource_i386.syso


server: resource_i386.syso resource_amd64.syso
	go build -v -o ${OUT} -ldflags='-X main.version="${VERSION}" -X main.build="${VERSION_BUILD}"' ${PKG}
	# cleaning i386 resource otherwise tests will fail ...
	rm resource_i386.syso

test:
	@go test -v -short ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

static: vet lint
	go build -i -v -o ${OUT}-v${VERSION} -tags netgo -ldflags="-extldflags \"-static\" -w -s -X main.version=${VERSION}" ${PKG}

run: server
	./${OUT}

dist: server 
	@echo Packaging Binaries...
	@rm -f __debug*.exe
	@mkdir -p ./dist
	@cp -R *.exe ./dist
	@rm -f ./dist/app.ini
	@cp *.ico ./dist
	@cp -R keys ./dist
	@cp *.ini ./dist
	@sh ./deploy.sh

clean:
	-@rm ${OUT} ${OUT}-v*

.PHONY: run server static vet lint