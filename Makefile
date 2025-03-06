.PHONY: git-init git-checkout golangci-lint-install lint blueprint fn-deploy fn-create fn-zip fn-create-version fn-timer fn-clear fn-list fn-version-list

REPO_NAME := $(shell basename $(CURDIR))
PROJECT := $(CURDIR)
LOCAL_BIN := $(CURDIR)/bin
ENV_PROD := .env.prod

ifneq (,$(wildcard ./.env))
ENV_FILE := .env
else ifneq (,$(wildcard ./.env.dev))
ENV_FILE := .env.dev
else
ENV_FILE :=
endif

ifneq ($(ENV_FILE),)
include $(ENV_FILE)
include $(ENV_PROD)
export
endif

# GIT
git-init:
	gh repo create $(USER)/$(REPO_NAME) --private
	git init
	git config user.name "$(USER)"
	git config user.email "$(EMAIL)"
	git add Makefile go.mod
	git commit -m "Init commit"
	git remote add origin git@github.com:$(USER)/$(REPO_NAME).git
	git remote -v
	git push -u origin master

BN ?= dev
.PHONY:
git-checkout:
	git checkout -b $(BN)

# LINT
lint-install:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2

lint:
	$(LOCAL_BIN)/golangci-lint run ./...


# PROJECT
blueprint:
	mkdir -p $(LOCAL_BIN)
	mkdir -p cmd && echo 'package main' > cmd/main.go
	mkdir -p internal/logger && echo 'package logger' > internal/logger/logger.go
	mkdir -p internal/model && echo 'package model' > internal/model/model.go
	mkdir -p internal/config && echo 'package config' > internal/config/config.go
	mkdir -p internal/service && echo 'package config' > internal/service/service.go
	mkdir -p internal/mailer && echo 'package mailer' > internal/mailer/mailer.go
	mkdir -p internal/filter && echo 'package filter' > internal/filter/filter.go
	mkdir -p internal/fetcher && echo 'package fetcher' > internal/fetcher/fetcher.go
	mkdir -p internal/cluster && echo 'package cluster' > internal/cluster/cluster.go

# Yandex Cloud
fn-create:
	@if yc serverless function get --name '$(YC_FUNC_NAME)' > /dev/null 2>&1; then \
  		echo "Function '$(YC_FUNC_NAME)' already exists, skipping creation."; \
	else \
		yc serverless function create --name "$(YC_FUNC_NAME)"; \
	fi

fn-zip:
	zip -r '$(YC_FUNC_NAME).zip' handler.go go.mod go.sum internal templates

fn-create-version:
	yc serverless function version create \
	--function-name '$(YC_FUNC_NAME)' \
 	--service-account-id '$(YC_SA_ID)' \
	--runtime golang121 \
	--entrypoint handler.Handler \
	--execution-timeout 120s \
	--memory 128m \
	--environment APP_MODE=prod \
	$(shell cat $(ENV_PROD) | grep -v '^#' | grep -v '^$$' | while read -r line; do echo "--environment $$line"; done) \
	--source-path "./$(YC_FUNC_NAME).zip"

fn-timer:
	@if yc serverless trigger get --name "run-$(YC_FUNC_NAME)" > /dev/null 2>&1; then \
  		echo "Trigger 'run-$(YC_FUNC_NAME)' already exists, skipping creation."; \
	else \
		yc serverless trigger create timer; \
		--cron-expression "$(YC_CRON)"; \
		--invoke-function-name "$(YC_FUNC_NAME)"; \
		--invoke-function-service-account-id "$(YC_SA_ID)"; \
		--name "run-$(YC_FUNC_NAME)"; \
	fi

fn-clear:
	rm "./$(YC_FUNC_NAME).zip"

fn-deploy: fn-create fn-zip fn-create-version fn-timer fn-clear

fn-list:
	yc serverless function list

fn-version-list:
	yc serverless function version list
