# Copyright 2015 The Prometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Ensure that 'all' is the default target otherwise it will be the first target from Makefile.common.
all::

# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 armv7 arm64

include Makefile.common

PROMTOOL_VERSION ?= 2.51.2
PROMTOOL_URL     ?= https://github.com/prometheus/prometheus/releases/download/v$(PROMTOOL_VERSION)/prometheus-$(PROMTOOL_VERSION).$(GO_BUILD_PLATFORM).tar.gz
PROMTOOL         ?= $(FIRST_GOPATH)/bin/promtool

DOCKER_IMAGE_NAME       ?= openziti-exporter
MACH                    ?= $(shell uname -m)

STATICCHECK_IGNORE =

# Use CGO for non-Linux builds.
ifeq ($(GOOS), linux)
	PROMU_CONF ?= .promu.yml
else
	ifndef GOOS
		ifeq ($(GOHOSTOS), linux)
			PROMU_CONF ?= .promu.yml
		else
			PROMU_CONF ?= .promu-cgo.yml
		endif
	else
		# Do not use CGO for openbsd/amd64 builds
		ifeq ($(GOOS), openbsd)
			ifeq ($(GOARCH), amd64)
				PROMU_CONF ?= .promu.yml
			else
				PROMU_CONF ?= .promu-cgo.yml
			endif
		else
			PROMU_CONF ?= .promu-cgo.yml
		endif
	endif
endif

PROMU := $(FIRST_GOPATH)/bin/promu --config $(PROMU_CONF)

# By default, "cross" test with ourselves to cover unknown pairings.
$(eval $(call goarch_pair,amd64,386))
$(eval $(call goarch_pair,mips64,mips))
$(eval $(call goarch_pair,mips64el,mipsel))

all:: vet common-all $(cross-test) $(test-e2e)

.PHONY: test-docker
test-docker:
	@echo ">> testing docker image"
	./test_image.sh "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)-linux-amd64:$(DOCKER_IMAGE_TAG)" 9100

.PHONY: promtool
promtool: $(PROMTOOL)

$(PROMTOOL):
	mkdir -p $(FIRST_GOPATH)/bin
	curl -fsS -L $(PROMTOOL_URL) | tar -xvzf - -C $(FIRST_GOPATH)/bin --strip 1 "prometheus-$(PROMTOOL_VERSION).$(GO_BUILD_PLATFORM)/promtool"
