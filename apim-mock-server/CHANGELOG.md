# Changelog

## [0.3.0](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-mock-server/v0.2.1...apim-mock-server/v0.3.0) (2026-11-26)


### Features

* **am-sdk:** generate from AM master 4.15.0 ([265ee38](https://github.com/gravitee-io/gravitee-automation-sdk/commit/265ee382cd55ceaa34dd3b5665349eaa0aa44eac))
* **apim-sdk:** generate from AIM 5.1.8 to carry the MCP auth policy refs ([2f82e60](https://github.com/gravitee-io/gravitee-automation-sdk/commit/2f82e607001ff2427c3d4211d121a7335361cba4))
* **apim-sdk:** generate the AuthZ Automation API ([3165f7d](https://github.com/gravitee-io/gravitee-automation-sdk/commit/3165f7d957cb5af46ba9ac74517750ee7b0ec4aa))


### Bug Fixes

* **apim-sdk:** drop studio tool entityIds from AimMcpProxyState.Spec() ([0cca0ed](https://github.com/gravitee-io/gravitee-automation-sdk/commit/0cca0ed2186187146485624cd834909775831bc2))
* **apim-sdk:** keep the PlanType constant names next to the AuthZ policy type ([6384483](https://github.com/gravitee-io/gravitee-automation-sdk/commit/6384483ed86eb5052cc2fe9b707462942dbdabfc))
* **ci:** read the AuthZ OAS through the GitHub API helper (APIM-16606) ([2f1a0af](https://github.com/gravitee-io/gravitee-automation-sdk/commit/2f1a0af60d26efeeee3dd3483b99b73e7b40ed7e))
* **deps:** bump common to v1.4.1, apim-sdk to v0.2.2 ([e42fa0e](https://github.com/gravitee-io/gravitee-automation-sdk/commit/e42fa0e6238560b3f11644b0e8673c3873cde849))

## [0.2.1](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-mock-server/v0.2.0...apim-mock-server/v0.2.1) (2026-11-19)


### Bug Fixes

* **apim-sdk:** accept a response template without status ([50a6e9d](https://github.com/gravitee-io/gravitee-automation-sdk/commit/50a6e9ded098804e3d3ddfe59e8affed6b30b4c5))
* **apim-sdk:** omit unset required fields APIM defaults ([9110ece](https://github.com/gravitee-io/gravitee-automation-sdk/commit/9110ece2104f78829c34698edb383d574899e0be))
* **deps:** bump apim-sdk to v0.2.1 ([1c818f4](https://github.com/gravitee-io/gravitee-automation-sdk/commit/1c818f46bc1bc94a391300c2c8b8969ad903a696))

## [0.2.0](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-mock-server/v0.1.1...apim-mock-server/v0.2.0) (2026-11-13)


### Features

* **apim-mock-server:** keep group members when ignoreMembers is set ([9ea5bc8](https://github.com/gravitee-io/gravitee-automation-sdk/commit/9ea5bc80748c983ce7b91f094606d8708d401065))
* **apim-sdk:** convert each resource State to its PUT payload ([99af048](https://github.com/gravitee-io/gravitee-automation-sdk/commit/99af048d69f244f2e4e16e173f7e106cc3515753))


### Bug Fixes

* **apim-sdk:** decode error details whatever their invalid value holds ([b76930e](https://github.com/gravitee-io/gravitee-automation-sdk/commit/b76930e2e65bfee41d19105e01cf979bc06a3cad))
* **deps:** bump apim-sdk to v0.2.0 ([ed0a973](https://github.com/gravitee-io/gravitee-automation-sdk/commit/ed0a9738c915f00f1ef2617c5e1084b90b814fbd))
* **deps:** bump common to v1.4.0 ([c0512c7](https://github.com/gravitee-io/gravitee-automation-sdk/commit/c0512c79ffd6ae08aeb7fa0d57c16f42cde94925))

## [0.1.1](https://github.com/gravitee-io/gravitee-automation-sdk/compare/apim-mock-server/v0.1.0...apim-mock-server/v0.1.1) (2026-11-10)


### Bug Fixes

* **deps:** bump apim-sdk to v0.1.1 ([c424ca1](https://github.com/gravitee-io/gravitee-automation-sdk/commit/c424ca1c6be3cc5b02ac9bbc9811c7a765461de5))
* **deps:** bump common to v1.3.0 ([623ed04](https://github.com/gravitee-io/gravitee-automation-sdk/commit/623ed0430585da3385631888ecc39e48a9f3fc0b))

## 0.1.0 (2026-11-09)


### Features

* **apim-mock-server:** add the APIM mock server ([ac7709a](https://github.com/gravitee-io/gravitee-automation-sdk/commit/ac7709af73ab205e4e5f46fe58bf332859af7b4e))
* **apim-mock-server:** release apim-mock-server binaries and document its differences from AM ([e68b9d7](https://github.com/gravitee-io/gravitee-automation-sdk/commit/e68b9d73822322c97b39a6496732f37155fe643a))
* **apim-mock-server:** test auth and OAS request validation ([26f84f8](https://github.com/gravitee-io/gravitee-automation-sdk/commit/26f84f89bcb59afffdd177f5ea6136f4fc4c79cf))
* **apim-mock-server:** test dry run on PUT and on the shared policy group DELETE ([cd98a7e](https://github.com/gravitee-io/gravitee-automation-sdk/commit/cd98a7e164a23cf9353c1365bded4e9470b89ca2))
* **apim-mock-server:** test sub-resources, parent checks, subscription references and cascades ([6328f6b](https://github.com/gravitee-io/gravitee-automation-sdk/commit/6328f6b784e34b577e6739d8c7bbbf66c99b2744))
* **apim-mock-server:** test the CRUD, defaults, ids and tenant isolation ([c91a08a](https://github.com/gravitee-io/gravitee-automation-sdk/commit/c91a08a473e5da6b795e3f6fe4229b33938254cd))


### Bug Fixes

* **deps:** bump apim-sdk to v0.1.0 ([b43683a](https://github.com/gravitee-io/gravitee-automation-sdk/commit/b43683ae4af0695f790404dbde4563a97778a280))
