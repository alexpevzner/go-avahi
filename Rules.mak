# CGo binding for Avahi
#
# Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
# See LICENSE for license terms and conditions
#
# Common part for Makefiles

# ----- Environment -----

TOPDIR  := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
GOFILES := $(wildcard *.go)
GOTESTS := $(wildcard *_test.go)
PACKAGE	:= $(shell basename $(shell pwd))

# ----- Parameters -----

GO	:= go
GOTAGS	:= $(shell which gotags 2>/dev/null)
GOLINT	:= $(shell which golint 2>/dev/null)

# ----- Common targets -----

.PHONY: all
.PHONY: clean
.PHONY: cover
.PHONY: tags
.PHONY: test
.PHONY: vet

# Recursive targets
all:	subdirs_all do_all
clean:	subdirs_clean do_clean
lint:	subdirs_lint do_lint
test:	subdirs_test do_test
vet:	subdirs_vet do_vet

# Non-recursive targets
cover:	do_cover

# Dependencies
do_all:	tags

# Default actions
tags:
do_all:
do_clean:
do_cover:
do_lint:
do_test:
do_vet:

# Conditional actions
ifneq   ($(GOFILES),)

do_all:
	$(GO) build
ifneq	($(GOTESTS),)
	$(GO) test -c
	rm -f $(PACKAGE).test
endif

do_cover:
ifneq	($(GOTESTS),)
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm -f coverage.out
endif

do_lint:
ifneq	($(GOLINT),)
	$(GOLINT) -set_exit_status
endif

do_test:
ifneq	($(GOTESTS),)
	$(GO) test
endif

do_vet:
	$(GO) vet

endif

ifneq	($(GOTAGS),)
tags:
	cd $(TOPDIR); gotags -R . > tags
endif

ifneq	($(CLEAN),)
do_clean:
	rm -f $(CLEAN)
endif

# ----- Subdirs handling

subdirs_all subdirs_lint subdirs_test subdirs_vet subdirs_clean:
	@for i in $(SUBDIRS); do \
                $(MAKE) -C $$i $(subst subdirs_,,$@) || exit 1; \
        done

