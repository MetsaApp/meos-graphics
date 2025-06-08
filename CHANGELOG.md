# Changelog

## [1.1.0](https://github.com/MetsaApp/meos-graphics/compare/v1.0.3...v1.1.0) (2025-06-08)


### Features

* add swagger hostname flag ([#52](https://github.com/MetsaApp/meos-graphics/issues/52)) ([e437eb9](https://github.com/MetsaApp/meos-graphics/commit/e437eb916200dd1dc386df6daace717553ac9b05))

## [1.0.3](https://github.com/MetsaApp/meos-graphics/compare/v1.0.2...v1.0.3) (2025-06-08)


### Bug Fixes

* add packages write permission for Docker registry push ([#50](https://github.com/MetsaApp/meos-graphics/issues/50)) ([5fb16dd](https://github.com/MetsaApp/meos-graphics/commit/5fb16dd6d93b249b290198d9d71c288e1f65b724))

## [1.0.2](https://github.com/MetsaApp/meos-graphics/compare/v1.0.1...v1.0.2) (2025-06-08)


### Bug Fixes

* convert repository name to lowercase for Docker registry ([#48](https://github.com/MetsaApp/meos-graphics/issues/48)) ([b10824d](https://github.com/MetsaApp/meos-graphics/commit/b10824df50131ff62db76846653c052d48300b7a))

## [1.0.1](https://github.com/MetsaApp/meos-graphics/compare/v1.0.0...v1.0.1) (2025-06-06)


### Bug Fixes

* add templ generation to release workflow and update version to 1.0.0 ([#38](https://github.com/MetsaApp/meos-graphics/issues/38)) ([ac0eb5e](https://github.com/MetsaApp/meos-graphics/commit/ac0eb5e5e457084b7beae453e88844cde1f4cb09))
* update release-please configuration for Go version updates ([#40](https://github.com/MetsaApp/meos-graphics/issues/40)) ([4a3a762](https://github.com/MetsaApp/meos-graphics/commit/4a3a76216215805b482f58f6baf7af38dbc4a163))
* use generic type with glob pattern for version.go updates ([#43](https://github.com/MetsaApp/meos-graphics/issues/43)) ([3808d45](https://github.com/MetsaApp/meos-graphics/commit/3808d457c0037ebe9f9d9985295393a9d02073fa))

## 1.0.0 (2025-05-26)


### Features

* add command-line flags for MeOS server configuration ([#12](https://github.com/MetsaApp/meos-graphics/issues/12)) ([c572f79](https://github.com/MetsaApp/meos-graphics/commit/c572f79d65a49e55a711af91358fef6df42ea2b9))
* add configurable poll interval for MeOS data updates ([#11](https://github.com/MetsaApp/meos-graphics/issues/11)) ([e9d474c](https://github.com/MetsaApp/meos-graphics/commit/e9d474cdc8f5427c9b4b22dc553051fb9644a142))
* add configurable timing for simulation mode ([#25](https://github.com/MetsaApp/meos-graphics/issues/25)) ([e3d0e58](https://github.com/MetsaApp/meos-graphics/commit/e3d0e58ce767da44085062c4b21a462b6c43d8c1))
* add pre-commit hooks with Lefthook ([#23](https://github.com/MetsaApp/meos-graphics/issues/23)) ([a101ae5](https://github.com/MetsaApp/meos-graphics/commit/a101ae519e1880708231c5ea1e0342a50ac419fe))
* Add web interface with real-time updates and API documentation ([#26](https://github.com/MetsaApp/meos-graphics/issues/26)) ([c30a6d2](https://github.com/MetsaApp/meos-graphics/commit/c30a6d2d91579bb699a4632ed6d732f332e325ea))
* auto-generate CLI documentation using Cobra ([#22](https://github.com/MetsaApp/meos-graphics/issues/22)) ([1814973](https://github.com/MetsaApp/meos-graphics/commit/1814973c371abdba946254b1a4401ee62d8ba32a))
* clean up REST API endpoints to return only necessary data ([#37](https://github.com/MetsaApp/meos-graphics/issues/37)) ([0e2580a](https://github.com/MetsaApp/meos-graphics/commit/0e2580a58a95a37d8d569dc87cf17058e12049e2))
* implement shared positions for competitors with identical times ([#36](https://github.com/MetsaApp/meos-graphics/issues/36)) ([1b1cce3](https://github.com/MetsaApp/meos-graphics/commit/1b1cce3c9845939b103afa9033f0a464c2c1b9a6))
* initial application structure with MeOS integration ([#1](https://github.com/MetsaApp/meos-graphics/issues/1)) ([e4cb035](https://github.com/MetsaApp/meos-graphics/commit/e4cb0352a9a4456b980d6ac369e9096df00d895d))
* migrate to templ templates with local assets and optimize simulation polling ([#34](https://github.com/MetsaApp/meos-graphics/issues/34)) ([4359666](https://github.com/MetsaApp/meos-graphics/commit/435966601c22f615eed91c4920c118bffbb85129))
* rename timeBehind field to timeDifference in API responses ([#13](https://github.com/MetsaApp/meos-graphics/issues/13)) ([3199cf4](https://github.com/MetsaApp/meos-graphics/commit/3199cf438e7df4dcd09fec21284c0d259cdefac6))


### Bug Fixes

* improve simulation timing distribution and phase boundary enforcement ([#33](https://github.com/MetsaApp/meos-graphics/issues/33)) ([a2e4faf](https://github.com/MetsaApp/meos-graphics/commit/a2e4faf0cd4b584519aa0eedf96acc339172a63e))
* resolve SSE updates not working in web interface ([#29](https://github.com/MetsaApp/meos-graphics/issues/29)) ([d2de5ab](https://github.com/MetsaApp/meos-graphics/commit/d2de5ab4956ae4a17d432e70df4ce00c979d6cc7))
