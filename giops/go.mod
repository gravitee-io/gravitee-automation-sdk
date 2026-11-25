module github.com/gravitee-io/gravitee-automation-sdk/giops

go 1.26.1

require (
	github.com/caarlos0/env/v11 v11.4.2
	github.com/go-playground/validator/v10 v10.30.6
	github.com/gravitee-io/gravitee-automation-sdk/am-mock-server v1.8.0
	github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2 v2.7.0
	github.com/gravitee-io/gravitee-automation-sdk/apim-mock-server v0.3.1
	github.com/gravitee-io/gravitee-automation-sdk/apim-sdk v0.3.0
	github.com/gravitee-io/gravitee-automation-sdk/common v1.4.1
	github.com/joho/godotenv v1.5.1
	github.com/labstack/gommon v0.4.2
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.9
	github.com/stretchr/testify v1.12.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/getkin/kin-openapi v0.149.0 // indirect
	github.com/go-chi/chi/v5 v5.3.2 // indirect
	github.com/go-openapi/jsonpointer v0.23.1 // indirect
	github.com/go-openapi/swag/jsonname v0.26.0 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/oapi-codegen/nethttp-middleware v1.1.2 // indirect
	github.com/oapi-codegen/oapi-codegen/v2 v2.8.0 // indirect
	github.com/oapi-codegen/runtime v1.7.0 // indirect
	github.com/oasdiff/yaml v0.1.1 // indirect
	github.com/oasdiff/yaml3 v0.0.14 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

replace github.com/gravitee-io/gravitee-automation-sdk/common => ../common

replace github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2 => ../am-sdk

replace github.com/gravitee-io/gravitee-automation-sdk/apim-sdk => ../apim-sdk

replace github.com/gravitee-io/gravitee-automation-sdk/am-mock-server => ../am-mock-server

replace github.com/gravitee-io/gravitee-automation-sdk/apim-mock-server => ../apim-mock-server
