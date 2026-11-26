# Changelog

## [0.4.0](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-sdk/v0.3.0...apim-sdk/v0.4.0) (2026-11-26)


### Features

* **giops:** CLI for the AM and APIM Automation APIs (proof of concept) ([f79f3e0](https://github.com/gravitee-io/gravitee-automation-sdk/commit/f79f3e0c34a27a237c10084f38c3bb2e18330488))


### Reverts

* feat(giops): CLI for the AM and APIM Automation APIs (proof of concept) ([892abdf](https://github.com/gravitee-io/gravitee-automation-sdk/commit/892abdf6df68605dd4d01267f0edcf298aa8a6ac))

## [0.3.0](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-sdk/v0.2.2...apim-sdk/v0.3.0) (2026-11-26)


### Features

* **am-sdk:** generate from AM master 4.15.0 ([265ee38](https://github.com/gravitee-io/gravitee-automation-sdk/commit/265ee382cd55ceaa34dd3b5665349eaa0aa44eac))
* **apim-sdk:** generate from AIM 5.1.8 to carry the MCP auth policy refs ([2f82e60](https://github.com/gravitee-io/gravitee-automation-sdk/commit/2f82e607001ff2427c3d4211d121a7335361cba4))
* **apim-sdk:** generate the AuthZ Automation API ([3165f7d](https://github.com/gravitee-io/gravitee-automation-sdk/commit/3165f7d957cb5af46ba9ac74517750ee7b0ec4aa))


### Bug Fixes

* **apim-sdk:** compare an AuthZ entity left without active as active ([f118024](https://github.com/gravitee-io/gravitee-automation-sdk/commit/f1180249252d554250a1102881896c22a22e2d19))
* **apim-sdk:** compare an AuthZ PDP tag left out and an empty one as equal ([7919c00](https://github.com/gravitee-io/gravitee-automation-sdk/commit/7919c002ed782a64b0109c313e9b1d000499aca9))
* **apim-sdk:** compare an AuthZ schema target left out and an empty one as equal ([821ae77](https://github.com/gravitee-io/gravitee-automation-sdk/commit/821ae77c7e3f28325950937250d10b17d89d84ae))
* **apim-sdk:** drop studio tool entityIds from AimMcpProxyState.Spec() ([0cca0ed](https://github.com/gravitee-io/gravitee-automation-sdk/commit/0cca0ed2186187146485624cd834909775831bc2))
* **apim-sdk:** keep the PlanType constant names next to the AuthZ policy type ([6384483](https://github.com/gravitee-io/gravitee-automation-sdk/commit/6384483ed86eb5052cc2fe9b707462942dbdabfc))
* **apim-sdk:** read an unset MCP proxy enableFGA and matchRequired as false ([65125aa](https://github.com/gravitee-io/gravitee-automation-sdk/commit/65125aaea68150c98d58367022ce4d395a0c442c))
* **ci:** read the AuthZ OAS through the GitHub API helper (APIM-16606) ([2f1a0af](https://github.com/gravitee-io/gravitee-automation-sdk/commit/2f1a0af60d26efeeee3dd3483b99b73e7b40ed7e))
* **deps:** bump common to v1.4.1, apim-sdk to v0.2.2 ([e42fa0e](https://github.com/gravitee-io/gravitee-automation-sdk/commit/e42fa0e6238560b3f11644b0e8673c3873cde849))

## [0.2.2](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-sdk/v0.2.1...apim-sdk/v0.2.2) (2026-11-23)


### Bug Fixes

* **apim-sdk:** tell expired and not yet active client certificates ([5af00e2](https://github.com/gravitee-io/gravitee-automation-sdk/commit/5af00e2b084af5f87f87212502904f00c057b861))

## [0.2.1](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-sdk/v0.2.0...apim-sdk/v0.2.1) (2026-11-19)


### Bug Fixes

* **apim-sdk:** accept a response template without status ([50a6e9d](https://github.com/gravitee-io/gravitee-automation-sdk/commit/50a6e9ded098804e3d3ddfe59e8affed6b30b4c5))
* **apim-sdk:** drift tags at GKO parity for v4 APIs ([63a2fb8](https://github.com/gravitee-io/gravitee-automation-sdk/commit/63a2fb896d4c5cf46bbdb2eecfa190eb24511300))
* **apim-sdk:** omit unset required fields APIM defaults ([9110ece](https://github.com/gravitee-io/gravitee-automation-sdk/commit/9110ece2104f78829c34698edb383d574899e0be))

## [0.2.0](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-sdk/v0.1.1...apim-sdk/v0.2.0) (2026-11-13)


### Features

* **apim-sdk:** convert each resource State to its PUT payload ([99af048](https://github.com/gravitee-io/gravitee-automation-sdk/commit/99af048d69f244f2e4e16e173f7e106cc3515753))
* **apim-sdk:** identify every Automation resource by its hrid ([009b141](https://github.com/gravitee-io/gravitee-automation-sdk/commit/009b1411e61220df19167ef8b389ca320b9bb420))


### Bug Fixes

* **apim-sdk:** decode error details whatever their invalid value holds ([b76930e](https://github.com/gravitee-io/gravitee-automation-sdk/commit/b76930e2e65bfee41d19105e01cf979bc06a3cad))
* **deps:** bump common to v1.4.0 ([c0512c7](https://github.com/gravitee-io/gravitee-automation-sdk/commit/c0512c79ffd6ae08aeb7fa0d57c16f42cde94925))

## [0.1.1](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-sdk/v0.1.0...apim-sdk/v0.1.1) (2026-11-10)


### Bug Fixes

* **deps:** bump common to v1.3.0 ([623ed04](https://github.com/gravitee-io/gravitee-automation-sdk/commit/623ed0430585da3385631888ecc39e48a9f3fc0b))

## 0.1.0 (2026-11-09)


### Features

* **apim-sdk:** add the models, operations and paths overlays ([6f3cc2c](https://github.com/gravitee-io/gravitee-automation-sdk/commit/6f3cc2c264beaec8035863e9119e66d8d7e4d416))
* **apim-sdk:** generate the models and client, with a cloud-token APIContext ([1d96aa6](https://github.com/gravitee-io/gravitee-automation-sdk/commit/1d96aa6efc08acbd1c9395c55eb076b31923f979))
* **apim-sdk:** implement store.Identifiable on list items ([45fdf4e](https://github.com/gravitee-io/gravitee-automation-sdk/commit/45fdf4ef4b0407956f2fe9866536c65d70a36218))
* **apim-sdk:** port the GKO drift tags in a drift overlay ([cf94b8e](https://github.com/gravitee-io/gravitee-automation-sdk/commit/cf94b8e263c5809e9c9f5852e01d3d203ed8ac17))
* **apim-sdk:** sync APIM OAS with the legacy UUID params ([7acaae4](https://github.com/gravitee-io/gravitee-automation-sdk/commit/7acaae4653e57ad344e25fefcd0b6a9a434af6f5))
* **apim-sdk:** sync APIM OAS with the masked dictionary reads ([e158897](https://github.com/gravitee-io/gravitee-automation-sdk/commit/e158897915757d5d6cac51d51c4ab63694fa7e69))
* **apim-sdk:** sync the APIM and AIM Automation OAS and merge them with mergespec ([739f9bb](https://github.com/gravitee-io/gravitee-automation-sdk/commit/739f9bb780f2afad83dab25e19c3cdaa7b69f361))
* **apim-sdk:** type polymorphic schemas as OneOf structs with one pointer per variant ([27d073d](https://github.com/gravitee-io/gravitee-automation-sdk/commit/27d073da98afa27961784c34a0040fc96bad7d43))
* move modules to github.com/gravitee-io/gravitee-automation-sdk ([c542010](https://github.com/gravitee-io/gravitee-automation-sdk/commit/c5420109ea40bbae76b948c9d8fce9f01033604f))
