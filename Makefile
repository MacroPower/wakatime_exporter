DOCKER_ARCHS      ?= amd64 armv7 arm64
DOCKER_IMAGE_NAME ?= wakatime-exporter
DOCKER_REPO       ?= macropower

all:: vet common-all

include Makefile.common
