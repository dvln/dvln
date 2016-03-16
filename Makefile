# This makefile was put together based on online Makefile 
# examples from around the net so thanks much for the quick
# training to all that worked on those training blogs/etc.
# - thanks much to all
#
# As we go forward the dvln focused version is under Apache 2.0:
# http://www.apache.org/licenses/LICENSE-2.0
# Copyright © 2015 Erik Brady <brady@dvln.org>

# Primary targets:
#   make -> produces 'dvln' binary in current dir by default (ie: 'build' target)
#   make build -> produces 'dvln' binary in the current directory
#   make build-race -> same as above but with -race added in for more race condition checking
#   make clean -> uses 'go clean' in all packages to clean up derived objects, also rm's dvln bin
#   make diff -> uses 'git diff' to run a diff on all non-index changes in ARK+modified vend pkgs
#   make diff-staged -> uses 'git diff --staged' to run a diff on all index changes in ARK+modified vend pkgs
#   make fmt -> formats all Go code in all ARK and modified open packages
#   make install -> "installs" the 'dvln' binary in $GOPATH/bin
#   make lint -> uses 'golint' to check for code issues in ARK packages (need golint in path)
#   make status -> uses 'git status --short' to run a cset on ARK and modified open packages
#   make status-long -> uses 'git status' to run a cset on ARK and modified open packages
#   make test -> runs 'go test -cover' on all packages
#   make test-race -> runs 'go test -race -cover' on all packages
#   make test-verbose -> runs 'go test -v -cover' on all packages
#   make vet -> uses 'go vet' to check for code issues on ARK packages

GO_CMD=go
GIT_CMD=git
GO_BUILD=$(GO_CMD) build
GO_BUILD_RACE=$(GO_CMD) build -race
GO_TEST=$(GO_CMD) test -cover
GO_TEST_RACE=$(GO_CMD) test -race -cover
GO_TEST_VERBOSE=$(GO_CMD) test -v -cover
GO_INSTALL=$(GO_CMD) install -v
GO_CLEAN=$(GO_CMD) clean
GO_VET=$(GO_CMD) vet
GO_FMT=$(GO_CMD) fmt
GO_LINT=golint
GIT_STATUS=$(GIT_CMD) status
GIT_STATUS_SHORT=$(GIT_CMD) status --short
GIT_DIFF=$(GIT_CMD) diff
GIT_DIFF_STAGED=$(GIT_CMD) diff --staged

# Package listings, basics are:
# - DVLN_TOOL is the main 'dvln' tool package
# - LOCAL_PACKAGE_LIST is all ARK focused Go packages
# - MODIFIED_VENDOR_PACKAGE_LIST is all vendor packages modified for dvln
# - VENDOR_PACKAGE_LIST is all vendor packages NOT modified for dvln
# Nested packages under top level packages are covered via the "/..." below
DVLN_TOOL := \
  github.com/dvln/dvln

LOCAL_PACKAGE_LIST := \
  github.com/dvln/api \
  github.com/dvln/codebase \
  github.com/dvln/devline \
  github.com/dvln/out \
  github.com/dvln/lib \
  github.com/dvln/pkg \
  github.com/dvln/toolver \
  github.com/dvln/util \
  github.com/dvln/vcs \
  github.com/dvln/wkspc

MODIFIED_VENDOR_PACKAGE_LIST := \
  github.com/dvln/afero \
  github.com/dvln/pflag \
  github.com/dvln/pretty \
  github.com/dvln/cobra \
  github.com/dvln/viper \
  github.com/dvln/nitro

VENDOR_PACKAGE_LIST := \
  github.com/dvln/cast \
  github.com/dvln/check \
  github.com/dvln/fsnotify \
  github.com/dvln/go-difflib \
  github.com/dvln/go-spew \
  github.com/dvln/hcl \
  github.com/dvln/mapstructure \
  github.com/dvln/objx \
  github.com/dvln/osext \
  github.com/dvln/properties \
  github.com/dvln/pty \
  github.com/dvln/str \
  github.com/dvln/testify \
  github.com/dvln/text \
  github.com/dvln/toml \
  github.com/dvln/yaml

PACKAGE_LIST := $(DVLN_TOOL) $(LOCAL_PACKAGE_LIST) $(MODIFIED_VENDOR_PACKAGE_LIST) $(VENDOR_PACKAGE_LIST)

.PHONY: all build build-race status status-long test test-race test-verbose test-fast test-fast-verbose install clean fmt vet lint

all: build

build:
	@echo "==> Building $(DVLN_TOOL) (results in current dir: dvln):"
	$(GO_BUILD) $(DVLN_TOOL)
	@echo "==> Build Complete"

build-race:
	@echo "==> Building $(DVLN_TOOL) (results in current dir: dvln):"
	$(GO_BUILD_RACE) $(DVLN_TOOL)
	@echo "==> Build Complete"

test:
	@for p in $(PACKAGE_LIST); do \
echo "==> Unit Testing $$p/...:"; \
$(GO_TEST) $$p/... || exit 1; \
done
	@echo "==> Unit Testing Complete"

test-race:
	@for p in $(PACKAGE_LIST); do \
echo "==> Unit Testing $$p/... (race):"; \
$(GO_TEST_RACE) $$p/... || exit 1; \
done
	@echo "==> Unit Testing Complete"

test-verbose:
	@for p in $(PACKAGE_LIST); do \
echo "==> Unit Testing $$p/... (verbose):"; \
$(GO_TEST_VERBOSE) $$p/... || exit 1; \
done
	@echo "==> Unit Testing Complete"

test-fast:
	@for p in $(LOCAL_PACKAGE_LIST) $(MODIFIED_VENDOR_PACKAGE_LIST) $(DVLN_TOOL); do \
echo "==> Unit Testing $$p/... (fast):"; \
$(GO_TEST) $$p... || exit 1; \
done
	@echo "==> Unit Testing Complete"

test-fast-verbose:
	@for p in $(LOCAL_PACKAGE_LIST) $(MODIFIED_VENDOR_PACKAGE_LIST) $(DVLN_TOOL); do \
echo "==> Unit Testing $$p/... (fast, verbose):"; \
$(GO_TEST_VERBOSE) $$p/... || exit 1; \
done
	@echo "==> Unit Testing Complete"

install: build
	@echo "==> Installing dvln into $(GOPATH)/bin/dvln:"
	@-mkdir -p $(GOPATH)/bin
	mv dvln $(GOPATH)/bin/dvln
	@echo "==> Installation Complete"

clean:
	@for p in $(PACKAGE_LIST); do \
echo "==> Cleaning $$p/...:"; \
$(GO_CLEAN) $$p; \
done
	@-rm -f dvln $(GOPATH)/bin/dvln $(GOPATH)/dvln
	@echo "==> Cleaning Complete"

fmt:
	@for p in $(LOCAL_PACKAGE_LIST) $(MODIFIED_VENDOR_PACKAGE_LIST) $(DVLN_TOOL); do \
echo "==> Formatting $$p/...:"; \
$(GO_FMT) $$p/... || exit 1; \
done
	@echo "==> Formatting Complete"

diff:
	@for p in $(LOCAL_PACKAGE_LIST) $(MODIFIED_VENDOR_PACKAGE_LIST) $(DVLN_TOOL); do \
echo "==> Diff (git diff) for $$p:"; \
cd $(GOPATH)/src/$$p;$(GIT_DIFF) || exit 1; \
done
	@echo "==> Status Complete"

diff-staged:
	@for p in $(LOCAL_PACKAGE_LIST) $(MODIFIED_VENDOR_PACKAGE_LIST) $(DVLN_TOOL); do \
echo "==> Diff (git diff) for $$p:"; \
cd $(GOPATH)/src/$$p;$(GIT_DIFF_STAGED) || exit 1; \
done
	@echo "==> Diff Complete"

status:
	@for p in $(LOCAL_PACKAGE_LIST) $(MODIFIED_VENDOR_PACKAGE_LIST) $(DVLN_TOOL); do \
echo "==> Status (git status -s) for $$p:"; \
cd $(GOPATH)/src/$$p;$(GIT_STATUS_SHORT) || exit 1; \
done
	@echo "==> Status Complete"

status-long:
	@for p in $(LOCAL_PACKAGE_LIST) $(MODIFIED_VENDOR_PACKAGE_LIST) $(DVLN_TOOL); do \
echo "==> Status (git status) for $$p:"; \
cd $(GOPATH)/src/$$p;$(GIT_STATUS) || exit 1; \
done
	@echo "==> Status Complete"

vet:
	@for p in $(DVLN_TOOL) $(LOCAL_PACKAGE_LIST); do \
echo "==> Vetting $$p/...:"; \
$(GO_VET) $$p/...; \
done
	@echo "==> Vetting Complete"

lint:
	@for p in $(DVLN_TOOL) $(LOCAL_PACKAGE_LIST); do \
echo "==> Linting $$p/...:"; \
$(GO_LINT) $$p/...; \
done
	@echo "==> Linting Complete"

