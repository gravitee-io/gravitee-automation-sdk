# Changelog

## [2.3.1](https://github.com/gravitee-io/gravitee-automation-sdk/compare/am-sdk/v2.3.0...am-sdk/v2.3.1) (2026-10-05)


### Bug Fixes

* add ':masked' to unstructed drift tag + OAS regen ([52ec51b](https://github.com/gravitee-io/gravitee-automation-sdk/commit/52ec51ba279bfb2b31395502661c8e8130ffe3b5))

## [2.3.0](https://github.com/gravitee-io/gravitee-automation-sdk/compare/am-sdk/v2.2.1...am-sdk/v2.3.0) (2026-09-29)


### Features

* move modules to github.com/gravitee-io/gravitee-automation-sdk ([c542010](https://github.com/gravitee-io/gravitee-automation-sdk/commit/c5420109ea40bbae76b948c9d8fce9f01033604f))

## [2.2.1](https://github.com/gravitee-io-labs/gravitee-automation-tools/compare/am-sdk/v2.2.0...am-sdk/v2.2.1) (2026-09-29)


### Bug Fixes

* sync oas (docs only) ([b38b5e3](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/b38b5e387a15e6613dd2df6b5f8fcb9a4b89b2d7))

## [2.2.0](https://github.com/gravitee-io-labs/gravitee-automation-tools/compare/am-sdk/v2.1.0...am-sdk/v2.2.0) (2026-09-28)


### Features

* **am-sdk:** type plugin configurations as unstructured objects ([ddb1d11](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/ddb1d11bc37dd349353c295a7897580d3986e39f))


### Bug Fixes

* **deps:** bump common to v1.1.0 ([f60bd10](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/f60bd100c7570bad6e9d64889612f5c77afab4b3))

## [2.1.0](https://github.com/gravitee-io-labs/gravitee-automation-tools/compare/am-sdk/v2.0.2...am-sdk/v2.1.0) (2026-09-25)


### Features

* **am-sdk:** make NewClient timeout optional and add WithHTTPClient ([e598e97](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/e598e97814e3efed9b06275998a00dbab23c47f7))

## [2.0.2](https://github.com/gravitee-io-labs/gravitee-automation-tools/compare/am-sdk/v2.0.1...am-sdk/v2.0.2) (2026-09-24)


### Bug Fixes

* regen following OAS changes ([81fd8a3](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/81fd8a32eaec6bcaa4c2a36119f55b5ddfe189e0))

## [2.0.1](https://github.com/gravitee-io-labs/gravitee-automation-tools/compare/am-sdk/v2.0.0...am-sdk/v2.0.1) (2026-09-24)


### Bug Fixes

* defaultAllowedAlgorithms wrong tag ([16d198d](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/16d198df9525d9f4e1fe954e8e245c810d55878e))
* regen following OAS changes ([07b0cf4](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/07b0cf4e5a7d9ab7e89a7b95e8dd3b3f0f75fe09))

## [2.0.0](https://github.com/gravitee-io-labs/gravitee-automation-tools/compare/am-sdk/v1.0.2...am-sdk/v2.0.0) (2026-09-24)


### ⚠ BREAKING CHANGES

* **am-sdk:** module path is now `github.com/gravitee-io-labs/gravitee-automation-tools/am-sdk/v2`. Per-resource packages (`pkg/sdk/domain`, `pkg/sdk/certificate`, ...) are replaced by `pkg/sdk`; `AMClient.Domains.X(...)` becomes `AMClient.X(...)`; optional slice fields change from `*[]T` to `[]T`.

### Features

* **am-sdk:** single sdk package with WithDefaults() on models ([c68cfd2](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/c68cfd2a312d29718c1ed4d4712b691a60c0c917))
* **am-sdk:** tag read-only properties with drift:"ignore" ([e6c91e6](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/e6c91e60fae6477fccc40a4686744aea1574fe68))


### Bug Fixes

* add IdJagSettings and KeyRetrievalSettings ([ffc00fa](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/ffc00fa6241e38ab637bbc58768f4fcf9083a344))

## [1.0.2](https://github.com/gravitee-io-labs/gravitee-automation-tools/compare/am-sdk/v1.0.1...am-sdk/v1.0.2) (2026-09-21)


### Bug Fixes

* implement identifiable interface ([9e8fe9f](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/9e8fe9f8fb68757475c09cf24f67b8cd861c6be8))

## [1.0.1](https://github.com/gravitee-io-labs/gravitee-automation-tools/compare/am-sdk/v1.0.0...am-sdk/v1.0.1) (2026-09-21)


### Bug Fixes

* implement identifiable interface ([e80bf3d](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/e80bf3d280011bb9f2543962a807f1778dc88fd1))
* sync OAS (dryrun) ([0a4230a](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/0a4230a67108630b84d1dd02be8369c33433cf98))

## 1.0.0 (2026-09-11)


### Features

* add Data Plane SDK client and mock after OAS sync ([6e11349](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/6e11349ed5e9e6b8a2cd47602772ee54d774c4db))
* add make generate target and CI generation check ([cb00e82](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/cb00e8227874d5a99946b23775ae0d007acc2b37))
* add mock server ([ed96de3](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/ed96de302de812728f383fcbd707c1daa7f18fca))
* add mock server ([6a7443e](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/6a7443e5d6ea5ad06de2165fdda8bb09cb4be060))
* AM client facade for domains ([496dbcd](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/496dbcdd0dcebe4eefde6bee4f9f775f93e68ff7))
* AM client gen ([a7a2b24](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/a7a2b24a1127a9f023637da48272f0a7b000fa23))
* complete AM mock resources and response helpers ([b2da3b6](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/b2da3b6ab5bbfcd0b180e343207f63f23b470a46))
* extract AM mock server module and shared store ([ff4b252](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/ff4b2520f9cd3e8ab995b9cc280a762311a79772))
* scope mock nested resources to their domain ([7d043f7](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/7d043f7757762c9d397444a714cd895ced00023e))


### Bug Fixes

* **deps:** bump common to v1.0.0 ([f61d720](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/f61d7202a701362d379a110759c73a0173a4adbf))
* **deps:** bump common to v1.0.0 ([bf4a596](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/bf4a596ce3350b51d03394bfe8b45761d2b49b52))
* drop Automation suffix ([62daa55](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/62daa55f9c5e8f0bb7cc764e2ca0822ec96638a2))
* org/env overlay, specialized client per resource ([3923841](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/392384177030eb0695cbaeceb2e7f6dc2e93d300))
* rename AM operationIds to get, list, upsert, delete ([9aa1dcb](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/9aa1dcb13264ef51723b1b6d65d9a02548b07a6d))
* split clients ([5637742](https://github.com/gravitee-io-labs/gravitee-automation-tools/commit/5637742557062a5835494bde8df37d853035fc4c))
