# Borrowed from https://gist.github.com/turtlemonvh/38bd3d73e61769767c35931d8c70ccb4

PACKAGE = irrenhaus-cli
BINARY = irrenhaus-cli
GOARCH = amd64

# Symlink into GOPATH
GITHUB_USERNAME=fuchsi
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${PACKAGE}
CURRENT_DIR=$(shell pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})

# Build the project
all: link clean linux darwin windows

link:
	BUILD_DIR=${BUILD_DIR}; \
	BUILD_DIR_LINK=${BUILD_DIR_LINK}; \
	CURRENT_DIR=${CURRENT_DIR}; \
	if [ "$${BUILD_DIR_LINK}" != "$${CURRENT_DIR}" ]; then \
	    echo "Fixing symlinks for build"; \
	    rm -f $${BUILD_DIR}; \
	    ln -s $${CURRENT_DIR} $${BUILD_DIR}; \
	fi

linux:
	cd ${BUILD_DIR}; \
	GOOS=linux GOARCH=${GOARCH} go build -o bin/${BINARY}-linux-${GOARCH} . ; \
	cd - >/dev/null

darwin:
	cd ${BUILD_DIR}; \
	GOOS=darwin GOARCH=${GOARCH} go build -o bin/${BINARY}-darwin-${GOARCH} . ; \
	cd - >/dev/null

windows:
	cd ${BUILD_DIR}; \
	GOOS=windows GOARCH=${GOARCH} go build -o bin/${BINARY}-windows-${GOARCH}.exe . ; \
	cd - >/dev/null

clean:
	-rm -f bin/*

install:
	@go install

.PHONY: link linux darwin windows clean
