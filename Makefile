MODULES := am-sdk apim-sdk common am-mock-server apim-mock-server giops

HAS_GO = [ -n "$$(find $$mod -name '*.go' -print -quit)" ]

# Every OAS is read through the GitHub API with the gh token, so a source repository turning private keeps working.
GH_RAW := gh api -H "Accept: application/vnd.github.raw"

AM_OAS_BRANCH ?= master
AM_OAS_URL := repos/gravitee-io/gravitee-access-management/contents/docs/automation/openapi.yaml?ref=$(AM_OAS_BRANCH)
AM_OAS_FILE := am-sdk/openapi/openapi.yaml

APIM_OAS_BRANCH ?= master
APIM_OAS_URL := repos/gravitee-io/gravitee-api-management/contents/gravitee-apim-rest-api/gravitee-apim-rest-api-automation/gravitee-apim-rest-api-automation-rest/src/main/resources/open-api.yaml?ref=$(APIM_OAS_BRANCH)
APIM_OAS_FILE := apim-sdk/openapi/sources/apim.yaml

AIM_OAS_REF ?= 5.1.8
AIM_OAS_FILE := apim-sdk/openapi/sources/aim.yaml

AUTHZ_OAS_REF ?= 3.23.4
AUTHZ_OAS_FILE := apim-sdk/openapi/sources/authz.yaml

##@ 🧹 Lint

.PHONY: lint
lint: lint-vet lint-staticcheck lint-revive lint-licenses ## Run all linters

.PHONY: lint-vet
lint-vet: ## Run go vet
	@for mod in $(MODULES); do \
		if $(HAS_GO); then \
			echo "==> go vet $$mod ..."; \
			(cd $$mod && go vet ./...) || exit 1; \
		fi; \
	done

.PHONY: lint-staticcheck
lint-staticcheck: ## Run staticcheck
	@for mod in $(MODULES); do \
		if $(HAS_GO); then \
			echo "==> staticcheck $$mod ..."; \
			(cd $$mod && staticcheck ./...) || exit 1; \
		fi; \
	done

.PHONY: lint-revive
lint-revive: ## Run revive
	@for mod in $(MODULES); do \
		if $(HAS_GO); then \
			echo "==> revive $$mod ..."; \
			(cd $$mod && revive -config ../.revive.toml -formatter friendly -exclude "./**/*.gen.go" ./...) || exit 1; \
		fi; \
	done

.PHONY: lint-licenses
lint-licenses: ## Check license headers
	@echo "Checking license headers ..."
	@addlicense -check -f LICENSE_TEMPLATE.txt \
		-ignore "**/*.gen.go" \
		-ignore "**/*.gen.yaml" \
		-ignore "apim-sdk/openapi/sources/aim.yaml" \
		-ignore "apim-sdk/openapi/sources/authz.yaml" \
		-ignore "**/node_modules/**" \
		-ignore ".github/**" \
		-ignore ".idea/**" \
		.

.PHONY: lint-fix
lint-fix: ## Auto-fix linting issues and add license headers
	@for mod in $(MODULES); do \
		find $$mod -name '*.go' -not -name '*.gen.go' -exec goimports -w {} +; \
	done
	@addlicense -f LICENSE_TEMPLATE.txt \
		-ignore "**/*.gen.go" \
		-ignore "**/*.gen.yaml" \
		-ignore "apim-sdk/openapi/sources/aim.yaml" \
		-ignore "apim-sdk/openapi/sources/authz.yaml" \
		-ignore "**/node_modules/**" \
		-ignore ".github/**" \
		-ignore ".idea/**" \
		.

##@ 🔄 Generate

.PHONY: sync-oas
sync-oas: ## Download AM and APIM Automation OAS from gravitee-access-management and gravitee-api-management (needs a gh login)
	$(GH_RAW) "$(AM_OAS_URL)" > $(AM_OAS_FILE)
	$(GH_RAW) "$(APIM_OAS_URL)" > $(APIM_OAS_FILE)

.PHONY: sync-aim-oas
sync-aim-oas: ## Download the AIM Automation OAS fragment from the private gravitee-gamma-module-aim (needs a gh login)
	$(GH_RAW) "repos/gravitee-io/gravitee-gamma-module-aim/contents/src/main/resources/openapi/openapi-automation.yaml?ref=$(AIM_OAS_REF)" > $(AIM_OAS_FILE)

.PHONY: verify-aim-oas
verify-aim-oas: sync-aim-oas ## Fail if the committed AIM OAS differs from gravitee-gamma-module-aim at $(AIM_OAS_REF)
	@git diff --exit-code -- $(AIM_OAS_FILE) || (echo "$(AIM_OAS_FILE) differs from gravitee-gamma-module-aim at $(AIM_OAS_REF)" && exit 1)

.PHONY: sync-authz-oas
sync-authz-oas: ## Download the AuthZ Automation OAS fragment from the private gravitee-gamma-module-authz (needs a gh login)
	$(GH_RAW) "repos/gravitee-io/gravitee-gamma-module-authz/contents/src/main/resources/openapi/openapi-automation.yaml?ref=$(AUTHZ_OAS_REF)" > $(AUTHZ_OAS_FILE)

.PHONY: verify-authz-oas
verify-authz-oas: sync-authz-oas ## Fail if the committed AuthZ OAS differs from gravitee-gamma-module-authz at $(AUTHZ_OAS_REF)
	@git diff --exit-code -- $(AUTHZ_OAS_FILE) || (echo "$(AUTHZ_OAS_FILE) differs from gravitee-gamma-module-authz at $(AUTHZ_OAS_REF)" && exit 1)

.PHONY: check-oas
check-oas: ## Fail if committed AM or APIM OAS differs from upstream $(AM_OAS_BRANCH) / $(APIM_OAS_BRANCH) (needs a gh login)
	$(GH_RAW) "$(AM_OAS_URL)" > /tmp/am-openapi.yaml
	@diff -u $(AM_OAS_FILE) /tmp/am-openapi.yaml || { echo "OAS out of date. Run 'make sync-oas' and commit."; exit 1; }
	$(GH_RAW) "$(APIM_OAS_URL)" > /tmp/apim-openapi.yaml
	@diff -u $(APIM_OAS_FILE) /tmp/apim-openapi.yaml || { echo "OAS out of date. Run 'make sync-oas' and commit."; exit 1; }

.PHONY: generate
generate: ## Run go generate across all modules
	@for mod in $(MODULES); do \
		if $(HAS_GO); then \
			echo "==> generate $$mod ..."; \
			(cd $$mod && go generate ./...) || exit 1; \
		fi; \
	done

##@ 🧪 Test

.PHONY: test
test: ## Run tests
	@for mod in $(MODULES); do \
		if $(HAS_GO); then \
			echo "==> testing $$mod ..."; \
			if [ "$$mod" = "common" ]; then \
				(cd $$mod && go test -race ./...) || exit 1; \
			else \
				(cd $$mod && go test ./...) || exit 1; \
			fi; \
		fi; \
	done

##@ 🔧 Tools

.PHONY: tools
tools: ## Install development tools
	go install honnef.co/go/tools/cmd/staticcheck@v0.8.1
	go install github.com/mgechev/revive@v1.16.0
	go install github.com/google/addlicense@v1.2.0
	go install golang.org/x/tools/cmd/goimports@v0.50.0
