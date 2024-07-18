# [2.7.0](https://github.com/jodevsa/wireguard-operator/compare/v2.6.2...v2.7.0) (2024-07-18)


### Features

* add liveness and readiness probe ([#205](https://github.com/jodevsa/wireguard-operator/issues/205)) ([3f57376](https://github.com/jodevsa/wireguard-operator/commit/3f573760194e52b5afe232e8efb9e08cee75329e))

## [2.6.2](https://github.com/jodevsa/wireguard-operator/compare/v2.6.1...v2.6.2) (2024-07-18)


### Bug Fixes

* revert golint setup as it introduce bug on netlink add ([c779a46](https://github.com/jodevsa/wireguard-operator/commit/c779a46689c38bf0e0e44df35f564e04ec352e9e))

# [2.4.0](https://github.com/jodevsa/wireguard-operator/compare/v2.3.2...v2.4.0) (2024-07-15)


### Features

* allow setting AllowedIps through WireguardPeer ([#191](https://github.com/jodevsa/wireguard-operator/issues/191)) ([e2d4d9d](https://github.com/jodevsa/wireguard-operator/commit/e2d4d9dc580517ee84048901b057e17b0a94cd73))

## [2.3.2](https://github.com/jodevsa/wireguard-operator/compare/v2.3.1...v2.3.2) (2024-07-14)


### Bug Fixes

* avoid hardcoding private key in config ([#200](https://github.com/jodevsa/wireguard-operator/issues/200)) ([018ae37](https://github.com/jodevsa/wireguard-operator/commit/018ae37687fd8221f8a0e15f7a651bfd83eaf48e))

## [2.3.1](https://github.com/jodevsa/wireguard-operator/compare/v2.3.0...v2.3.1) (2024-07-14)


### Bug Fixes

* privateKeyRef case ([#199](https://github.com/jodevsa/wireguard-operator/issues/199)) ([f880152](https://github.com/jodevsa/wireguard-operator/commit/f8801526e559365f85d7359846bc738239817643))


### Reverts

* Revert "fix manifest case ([#197](https://github.com/jodevsa/wireguard-operator/issues/197))" ([#198](https://github.com/jodevsa/wireguard-operator/issues/198)) ([88dd510](https://github.com/jodevsa/wireguard-operator/commit/88dd510f94b8d6ec9b14897adba8eaa5c13e6b5f))

# [2.3.0](https://github.com/jodevsa/wireguard-operator/compare/v2.2.0...v2.3.0) (2024-07-13)


### Features

* add node selector ([#194](https://github.com/jodevsa/wireguard-operator/issues/194)) ([7db3b17](https://github.com/jodevsa/wireguard-operator/commit/7db3b1745869613550f7f5d3054d94c5816acb6f))

# [2.2.0](https://github.com/jodevsa/wireguard-operator/compare/v2.1.0...v2.2.0) (2024-07-05)


### Features

* pass resources to deployment' ([d516579](https://github.com/jodevsa/wireguard-operator/commit/d516579f371c1af0cc37ade0b7adf47b8225d669))

# [2.1.0](https://github.com/jodevsa/wireguard-operator/compare/v2.0.30...v2.1.0) (2024-07-03)


### Features

* pass address to LoadBalancerIP ([#184](https://github.com/jodevsa/wireguard-operator/issues/184)) ([d954499](https://github.com/jodevsa/wireguard-operator/commit/d95449998bd8a3cbddebdc74a738c48698fe22ec))

## [2.0.30](https://github.com/jodevsa/wireguard-operator/compare/v2.0.29...v2.0.30) (2024-05-03)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.17.2 ([#155](https://github.com/jodevsa/wireguard-operator/issues/155)) ([931c9dd](https://github.com/jodevsa/wireguard-operator/commit/931c9dd49009c5004e8a9dc70b691575fb48ad1f))

## [2.0.29](https://github.com/jodevsa/wireguard-operator/compare/v2.0.28...v2.0.29) (2024-04-29)


### Bug Fixes

* **deps:** update module github.com/onsi/gomega to v1.33.0 ([#152](https://github.com/jodevsa/wireguard-operator/issues/152)) ([84140a8](https://github.com/jodevsa/wireguard-operator/commit/84140a8ba0ef273c77ccb8a54288bd09f145da13))

## [2.0.28](https://github.com/jodevsa/wireguard-operator/compare/v2.0.27...v2.0.28) (2024-04-24)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([#131](https://github.com/jodevsa/wireguard-operator/issues/131)) ([d045264](https://github.com/jodevsa/wireguard-operator/commit/d045264973027defd316bd9c48863f4e01df06f3))

## [2.0.27](https://github.com/jodevsa/wireguard-operator/compare/v2.0.26...v2.0.27) (2024-04-17)


### Bug Fixes

* **deps:** update kubernetes packages to v0.29.4 ([#150](https://github.com/jodevsa/wireguard-operator/issues/150)) ([3ebc5db](https://github.com/jodevsa/wireguard-operator/commit/3ebc5db5a8adab91b2191e0ab93cf06135c24b51))

## [2.0.26](https://github.com/jodevsa/wireguard-operator/compare/v2.0.25...v2.0.26) (2024-04-11)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.17.1 ([#137](https://github.com/jodevsa/wireguard-operator/issues/137)) ([bd151b0](https://github.com/jodevsa/wireguard-operator/commit/bd151b0b774f8dc052dc941a2acc85476744db07))

## [2.0.25](https://github.com/jodevsa/wireguard-operator/compare/v2.0.24...v2.0.25) (2024-04-03)


### Bug Fixes

* **deps:** update kubernetes packages to v0.29.3 ([#136](https://github.com/jodevsa/wireguard-operator/issues/136)) ([e05b4c1](https://github.com/jodevsa/wireguard-operator/commit/e05b4c142bce8aaee6ffb4652d7b3b4c0beb51a3))

## [2.0.24](https://github.com/jodevsa/wireguard-operator/compare/v2.0.23...v2.0.24) (2024-04-03)


### Bug Fixes

* **deps:** update module github.com/onsi/gomega to v1.32.0 ([#138](https://github.com/jodevsa/wireguard-operator/issues/138)) ([3fa6e68](https://github.com/jodevsa/wireguard-operator/commit/3fa6e685a49531186f7735614965c40dcb279db6))

## [2.0.23](https://github.com/jodevsa/wireguard-operator/compare/v2.0.22...v2.0.23) (2024-02-28)


### Bug Fixes

* update controller tools to 0.14.0 ([#146](https://github.com/jodevsa/wireguard-operator/issues/146)) ([c5c78f1](https://github.com/jodevsa/wireguard-operator/commit/c5c78f1245d04aa650619e4127568e2e3810d20c))

## [2.0.22](https://github.com/jodevsa/wireguard-operator/compare/v2.0.21...v2.0.22) (2024-02-28)


### Bug Fixes

* **deps:** update module sigs.k8s.io/kind to v0.22.0 ([#142](https://github.com/jodevsa/wireguard-operator/issues/142)) ([c856013](https://github.com/jodevsa/wireguard-operator/commit/c85601312666fdce5c7f0dd0f3b5b167c9ddf171))

## [2.0.21](https://github.com/jodevsa/wireguard-operator/compare/v2.0.20...v2.0.21) (2024-02-28)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.14.0 ([#135](https://github.com/jodevsa/wireguard-operator/issues/135)) ([e8a2fac](https://github.com/jodevsa/wireguard-operator/commit/e8a2fac1efae929a07b20cd814e42e5b434f7eda))
* **deps:** update module sigs.k8s.io/kind to v0.21.0 ([#140](https://github.com/jodevsa/wireguard-operator/issues/140)) ([73ff848](https://github.com/jodevsa/wireguard-operator/commit/73ff848b4c9e0b30627a3f639463cf8c3b2555f5))
* use nodejs 20.8.1 for semantic-release ([d1fa423](https://github.com/jodevsa/wireguard-operator/commit/d1fa4236ec143a45a295fc7d087ccfcace03648f))
* use nodejs 20.8.1 for semantic-release ([20825be](https://github.com/jodevsa/wireguard-operator/commit/20825bebe562d4516824f98eddacbd760fa44ca3))

## [2.0.20](https://github.com/jodevsa/wireguard-operator/compare/v2.0.19...v2.0.20) (2023-12-28)


### Bug Fixes

* **deps:** update module github.com/go-logr/logr to v1.4.1 ([#134](https://github.com/jodevsa/wireguard-operator/issues/134)) ([cd41df3](https://github.com/jodevsa/wireguard-operator/commit/cd41df3fd329a68b6414dbff160db2393ecb254b))

## [2.0.19](https://github.com/jodevsa/wireguard-operator/compare/v2.0.18...v2.0.19) (2023-12-14)


### Bug Fixes

* **deps:** update kubernetes packages to v0.29.0 ([#130](https://github.com/jodevsa/wireguard-operator/issues/130)) ([f30c389](https://github.com/jodevsa/wireguard-operator/commit/f30c38902d12a27b8621677287345394e1ad569f))

## [2.0.18](https://github.com/jodevsa/wireguard-operator/compare/v2.0.17...v2.0.18) (2023-12-07)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([#126](https://github.com/jodevsa/wireguard-operator/issues/126)) ([bdff4d8](https://github.com/jodevsa/wireguard-operator/commit/bdff4d836f4444a2be18994c7e42fe985dfce5f8))

## [2.0.17](https://github.com/jodevsa/wireguard-operator/compare/v2.0.16...v2.0.17) (2023-12-01)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.13.2 ([#125](https://github.com/jodevsa/wireguard-operator/issues/125)) ([89fab82](https://github.com/jodevsa/wireguard-operator/commit/89fab827da31fb5da7690ef6bf068a73782ddc8c))

## [2.0.16](https://github.com/jodevsa/wireguard-operator/compare/v2.0.15...v2.0.16) (2023-11-28)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([#124](https://github.com/jodevsa/wireguard-operator/issues/124)) ([f368544](https://github.com/jodevsa/wireguard-operator/commit/f36854452424c8e14029e768a0f50f9afb3f7c38))

## [2.0.15](https://github.com/jodevsa/wireguard-operator/compare/v2.0.14...v2.0.15) (2023-11-22)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([#119](https://github.com/jodevsa/wireguard-operator/issues/119)) ([7bc059f](https://github.com/jodevsa/wireguard-operator/commit/7bc059f2e23eb82f5084d0188621399eab4b4bef))

## [2.0.14](https://github.com/jodevsa/wireguard-operator/compare/v2.0.13...v2.0.14) (2023-11-16)


### Bug Fixes

* **deps:** update kubernetes packages to v0.28.4 ([#121](https://github.com/jodevsa/wireguard-operator/issues/121)) ([b22e613](https://github.com/jodevsa/wireguard-operator/commit/b22e6137d8dc4c4f4f393f18e179ab8e4ca76e26))

## [2.0.13](https://github.com/jodevsa/wireguard-operator/compare/v2.0.12...v2.0.13) (2023-11-15)


### Bug Fixes

* **deps:** update module github.com/onsi/gomega to v1.30.0 ([#117](https://github.com/jodevsa/wireguard-operator/issues/117)) ([57f4cfa](https://github.com/jodevsa/wireguard-operator/commit/57f4cfaccf8ba7d1f9c572f4f3642661a99e9edf))

## [2.0.12](https://github.com/jodevsa/wireguard-operator/compare/v2.0.11...v2.0.12) (2023-11-12)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v1.16.5 ([#107](https://github.com/jodevsa/wireguard-operator/issues/107)) ([76fb2d3](https://github.com/jodevsa/wireguard-operator/commit/76fb2d3fa05db39c84dd2a35da4010453d56add5))

## [2.0.11](https://github.com/jodevsa/wireguard-operator/compare/v2.0.10...v2.0.11) (2023-11-11)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.13.1 ([#118](https://github.com/jodevsa/wireguard-operator/issues/118)) ([15a95fc](https://github.com/jodevsa/wireguard-operator/commit/15a95fccb019fa26ac244d39fa7a320052ac9f48))

## [2.0.10](https://github.com/jodevsa/wireguard-operator/compare/v2.0.9...v2.0.10) (2023-11-08)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([#108](https://github.com/jodevsa/wireguard-operator/issues/108)) ([df48a2d](https://github.com/jodevsa/wireguard-operator/commit/df48a2dfb1acf36d777b20a54679561f37d1db0e))

## [2.0.9](https://github.com/jodevsa/wireguard-operator/compare/v2.0.8...v2.0.9) (2023-11-07)


### Bug Fixes

* **deps:** update module github.com/fsnotify/fsnotify to v1.7.0 ([#112](https://github.com/jodevsa/wireguard-operator/issues/112)) ([23edb6d](https://github.com/jodevsa/wireguard-operator/commit/23edb6db8a65f96eebb272ce9eb400ac3de43fd4))
* **deps:** update module github.com/onsi/gomega to v1.29.0 ([#113](https://github.com/jodevsa/wireguard-operator/issues/113)) ([2a62db3](https://github.com/jodevsa/wireguard-operator/commit/2a62db35f156be331da7601e09d29e919c21e68c))

## [2.0.8](https://github.com/jodevsa/wireguard-operator/compare/v2.0.7...v2.0.8) (2023-10-30)


### Bug Fixes

* **deps:** update module github.com/go-logr/logr to v1.3.0 ([#114](https://github.com/jodevsa/wireguard-operator/issues/114)) ([38755d7](https://github.com/jodevsa/wireguard-operator/commit/38755d7907e08ee90e89a4c21d23a95132e38bd2))

## [2.0.7](https://github.com/jodevsa/wireguard-operator/compare/v2.0.6...v2.0.7) (2023-10-19)


### Bug Fixes

* **deps:** update kubernetes packages to v0.28.3 ([#111](https://github.com/jodevsa/wireguard-operator/issues/111)) ([06ecbff](https://github.com/jodevsa/wireguard-operator/commit/06ecbff1fafed1da5bb5327a382d66d87f6d62a3))

## [2.0.6](https://github.com/jodevsa/wireguard-operator/compare/v2.0.5...v2.0.6) (2023-10-14)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.13.0 ([#109](https://github.com/jodevsa/wireguard-operator/issues/109)) ([060ffca](https://github.com/jodevsa/wireguard-operator/commit/060ffca156872f92cde9249b04a50f114cf8c1d7))

## [2.0.5](https://github.com/jodevsa/wireguard-operator/compare/v2.0.4...v2.0.5) (2023-10-04)


### Bug Fixes

* **deps:** update module github.com/onsi/gomega to v1.28.0 ([#105](https://github.com/jodevsa/wireguard-operator/issues/105)) ([ae5343f](https://github.com/jodevsa/wireguard-operator/commit/ae5343f3dba230e548b1530999d06a7ab585d966))

## [2.0.4](https://github.com/jodevsa/wireguard-operator/compare/v2.0.3...v2.0.4) (2023-09-23)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([#102](https://github.com/jodevsa/wireguard-operator/issues/102)) ([a6f72b5](https://github.com/jodevsa/wireguard-operator/commit/a6f72b51da4cb4ab96757cfbd4dea2d6a60208cd))

## [2.0.3](https://github.com/jodevsa/wireguard-operator/compare/v2.0.2...v2.0.3) (2023-09-23)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.12.1 ([#104](https://github.com/jodevsa/wireguard-operator/issues/104)) ([0a12404](https://github.com/jodevsa/wireguard-operator/commit/0a12404ba7710f980b8fb65043deca41c162cc6d))

## [2.0.2](https://github.com/jodevsa/wireguard-operator/compare/v2.0.1...v2.0.2) (2023-09-15)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([#84](https://github.com/jodevsa/wireguard-operator/issues/84)) ([f05a67f](https://github.com/jodevsa/wireguard-operator/commit/f05a67fc2957a74897b7f7c629e8d9e0c4e96380))

## [2.0.1](https://github.com/jodevsa/wireguard-operator/compare/v2.0.0...v2.0.1) (2023-09-14)


### Bug Fixes

* **deps:** update kubernetes packages to v0.28.2 ([#100](https://github.com/jodevsa/wireguard-operator/issues/100)) ([cd25cf6](https://github.com/jodevsa/wireguard-operator/commit/cd25cf6a4e8343a14c0ab9648b3214acf21f3fa4))

# [2.0.0](https://github.com/jodevsa/wireguard-operator/compare/v1.2.20...v2.0.0) (2023-09-02)


### Bug Fixes

* Change apiversion to canonical one ([#93](https://github.com/jodevsa/wireguard-operator/issues/93)) ([4bd7537](https://github.com/jodevsa/wireguard-operator/commit/4bd7537b9a71a77eaa71842fa7dd1da649cd0948))


### BREAKING CHANGES

* change in apiversion will result in a new crd

## [1.2.20](https://github.com/jodevsa/wireguard-operator/compare/v1.2.19...v1.2.20) (2023-08-27)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.12.0 ([feb6d9a](https://github.com/jodevsa/wireguard-operator/commit/feb6d9a13dc082a16cdccefa0eea8390894c8d1c))

## [1.2.19](https://github.com/jodevsa/wireguard-operator/compare/v1.2.18...v1.2.19) (2023-08-27)


### Bug Fixes

* **deps:** update kubernetes packages to v0.28.1 ([23b7a7c](https://github.com/jodevsa/wireguard-operator/commit/23b7a7ce533300b790d01f024184678210696fe7))

## [1.2.18](https://github.com/jodevsa/wireguard-operator/compare/v1.2.17...v1.2.18) (2023-08-14)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v1.16.5 ([133fd58](https://github.com/jodevsa/wireguard-operator/commit/133fd589f3de8ef32dd330e4607350133333cc60))

## [1.2.17](https://github.com/jodevsa/wireguard-operator/compare/v1.2.16...v1.2.17) (2023-08-04)


### Bug Fixes

* **deps:** update module sigs.k8s.io/controller-runtime to v0.15.1 ([55da080](https://github.com/jodevsa/wireguard-operator/commit/55da08053104abc1451ee8fc45375129e32071d2))

## [1.2.16](https://github.com/jodevsa/wireguard-operator/compare/v1.2.15...v1.2.16) (2023-07-29)


### Bug Fixes

* **deps:** update module github.com/onsi/gomega to v1.27.10 ([a0e42ff](https://github.com/jodevsa/wireguard-operator/commit/a0e42ffce9785ccec67cb0c613f1548ec1e398a3))
* **deps:** update module github.com/onsi/gomega to v1.27.9 ([b635d9b](https://github.com/jodevsa/wireguard-operator/commit/b635d9b0507ae6d5e2a77d644e9fd35e6ec1edd4))

## [1.2.15](https://github.com/jodevsa/wireguard-operator/compare/v1.2.14...v1.2.15) (2023-07-20)


### Bug Fixes

* **deps:** update kubernetes packages to v0.27.4 ([9cc9f45](https://github.com/jodevsa/wireguard-operator/commit/9cc9f4534f80c2e00c9ac6f1f3062dbc2d0d1705))

## [1.2.14](https://github.com/jodevsa/wireguard-operator/compare/v1.2.13...v1.2.14) (2023-06-25)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([6ca758f](https://github.com/jodevsa/wireguard-operator/commit/6ca758ffe008b437484a8399a711aeb920596a0c))

## [1.2.13](https://github.com/jodevsa/wireguard-operator/compare/v1.2.12...v1.2.13) (2023-06-25)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([900a4e8](https://github.com/jodevsa/wireguard-operator/commit/900a4e8ba74f9bd39ec1f9aee2f94ea7dfeae35a))

## [1.2.12](https://github.com/jodevsa/wireguard-operator/compare/v1.2.11...v1.2.12) (2023-06-25)


### Bug Fixes

* **deps:** update kubernetes packages to v0.27.3 ([964b192](https://github.com/jodevsa/wireguard-operator/commit/964b1924c202a6c1975a9f1054fc52d0a3318c6d))
* **deps:** update module github.com/onsi/ginkgo to v2 ([71f22b3](https://github.com/jodevsa/wireguard-operator/commit/71f22b3f5b922f65decaf50ef6f29de6acd6481f))
* **deps:** update module sigs.k8s.io/kind to v0.20.0 ([1a01801](https://github.com/jodevsa/wireguard-operator/commit/1a01801a6684377bfb6602fb578ea2efdde7f691))

## [1.2.11](https://github.com/jodevsa/wireguard-operator/compare/v1.2.10...v1.2.11) (2023-06-10)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([85a2a2b](https://github.com/jodevsa/wireguard-operator/commit/85a2a2b86c2fec84f6d37f6e7dd367d9e904dadc))

## [1.2.10](https://github.com/jodevsa/wireguard-operator/compare/v1.2.9...v1.2.10) (2023-06-10)


### Bug Fixes

* **deps:** update module github.com/onsi/gomega to v1.27.8 ([d9e398d](https://github.com/jodevsa/wireguard-operator/commit/d9e398d1a6e6cebe5479f772cb13a68506c37875))

## [1.2.9](https://github.com/jodevsa/wireguard-operator/compare/v1.2.8...v1.2.9) (2023-06-10)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v1.16.5 ([d26cb03](https://github.com/jodevsa/wireguard-operator/commit/d26cb036fe332fd675e4c9d2fb7e62883732bd97))

## [1.2.8](https://github.com/jodevsa/wireguard-operator/compare/v1.2.7...v1.2.8) (2023-05-23)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([ec79bd8](https://github.com/jodevsa/wireguard-operator/commit/ec79bd802e648dd9500a0f4868f1a2f26eb1f742))

## [1.2.7](https://github.com/jodevsa/wireguard-operator/compare/v1.2.6...v1.2.7) (2023-05-23)


### Bug Fixes

* **deps:** update module sigs.k8s.io/controller-runtime to v0.15.0 ([48b078b](https://github.com/jodevsa/wireguard-operator/commit/48b078bd0f2058dc8d8ea40dd457a3214353672c))

## [1.2.6](https://github.com/jodevsa/wireguard-operator/compare/v1.2.5...v1.2.6) (2023-05-19)


### Bug Fixes

* **tests:** add integration test ([e703aff](https://github.com/jodevsa/wireguard-operator/commit/e703aff91b0a8b9e9ac928335f6cb3b73085d2d7))

## [1.2.5](https://github.com/jodevsa/wireguard-operator/compare/v1.2.4...v1.2.5) (2023-05-15)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v2 ([0fe071f](https://github.com/jodevsa/wireguard-operator/commit/0fe071fcb369f20eafeeb774017886e5cbe5b103))

## [1.2.4](https://github.com/jodevsa/wireguard-operator/compare/v1.2.3...v1.2.4) (2023-05-15)


### Bug Fixes

* **deps:** update module github.com/onsi/gomega to v1.27.6 ([e6132b8](https://github.com/jodevsa/wireguard-operator/commit/e6132b8ceb10310ba850cee6820b33a8dde7ef12))

## [1.2.3](https://github.com/jodevsa/wireguard-operator/compare/v1.2.2...v1.2.3) (2023-05-14)


### Bug Fixes

* do not include package*.json in the bot commit ([74f8134](https://github.com/jodevsa/wireguard-operator/commit/74f813401ce75586df61d9b54373fb79667c361f))
* remove duplicate step in release workflow ([62158b2](https://github.com/jodevsa/wireguard-operator/commit/62158b21a3ca8a1eb46ae6cf0f6959ff5288ad0d))
* remove package.json ([4002f19](https://github.com/jodevsa/wireguard-operator/commit/4002f198b306ff42bc213d3948e535c539301a20))

## [1.2.2](https://github.com/jodevsa/wireguard-operator/compare/v1.2.1...v1.2.2) (2023-05-14)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo to v1.16.5 ([09d7a61](https://github.com/jodevsa/wireguard-operator/commit/09d7a61e165432d71c88100ea2a82e9f37755bac))

## [1.2.1](https://github.com/jodevsa/wireguard-operator/compare/v1.2.0...v1.2.1) (2023-05-14)


### Bug Fixes

* **deps:** update golang.zx2c4.com/wireguard/wgctrl digest to 925a1e7 ([628f92e](https://github.com/jodevsa/wireguard-operator/commit/628f92e643c73324194ccea1a720a92ed0b5f658))
* fix releases with protected main branch ([65ee033](https://github.com/jodevsa/wireguard-operator/commit/65ee0338e81d890de20de0091a410ee0a825525d))
* **workflow:** do not persist credentials on checkout ([6de8279](https://github.com/jodevsa/wireguard-operator/commit/6de82792e27d42eb997aeb76a48fc5ecfbe52b68))

# [1.2.0](https://github.com/jodevsa/wireguard-operator/compare/v1.1.0...v1.2.0) (2023-05-13)


### Features

* create a pipeline on PR ([ba2c0d9](https://github.com/jodevsa/wireguard-operator/commit/ba2c0d9e1ef7bc2d05c75c83bc3187166c743eaa))

# [1.1.0](https://github.com/jodevsa/wireguard-operator/compare/v1.0.9...v1.1.0) (2023-05-13)


### Features

* automate changelog creation ([310836f](https://github.com/jodevsa/wireguard-operator/commit/310836f6bd27a1a41196d2fcb4bc9705f0eb4810))

## [1.0.7](https://github.com/jodevsa/wireguard-operator/compare/v1.0.6...v1.0.7) (2023-05-13)


### Bug Fixes

* fix ([3f33021](https://github.com/jodevsa/wireguard-operator/commit/3f33021b96cfcb04603199436a87dfc7ba79c3f9))
* fix ([80cd65d](https://github.com/jodevsa/wireguard-operator/commit/80cd65d73bba3925229f3a700d40adfec36c80ad))
* fix ([754b639](https://github.com/jodevsa/wireguard-operator/commit/754b639a6cf5da377ba0dac7b542eea44d95a191))
* fix ([4236e3b](https://github.com/jodevsa/wireguard-operator/commit/4236e3bd20bf04a3c5f2d1d5196ec15847cb1f1a))
* fix ([137c226](https://github.com/jodevsa/wireguard-operator/commit/137c22636ffa5332c91c37966f50a5837983f748))
* fix ([849c1e1](https://github.com/jodevsa/wireguard-operator/commit/849c1e1017304109bfa8cafe238e51b477e98f99))
* fix ([a069a6b](https://github.com/jodevsa/wireguard-operator/commit/a069a6bc23daccb87600094ec29a4e1a017e7d97))
* fix ([7418851](https://github.com/jodevsa/wireguard-operator/commit/7418851f6f838c19b9d7712c0bbac1f50cbcd060))
* fix ([c623979](https://github.com/jodevsa/wireguard-operator/commit/c62397923a940d7825f1ddaa14b311596c748c0b))
* fix ([f32abaf](https://github.com/jodevsa/wireguard-operator/commit/f32abaf6b1bd643a7267b83c4733daef148aabde))
* upload release file ([a6f1e58](https://github.com/jodevsa/wireguard-operator/commit/a6f1e588a57906476f23a781d822d09e69142ee5))
* upload release file ([5e8cb31](https://github.com/jodevsa/wireguard-operator/commit/5e8cb313516987ca03b887581d77d02be730393e))
* use main as ref ([d512549](https://github.com/jodevsa/wireguard-operator/commit/d512549aba2dcea11577bb2063b90d0ed08c82be))
* workflow ([9df4c47](https://github.com/jodevsa/wireguard-operator/commit/9df4c473de0b803bea8ab21d5cc3e0f291507711))
* workflow ([372b683](https://github.com/jodevsa/wireguard-operator/commit/372b683e0b8a43d93686a75ecddab7f417f0a8ee))

## [1.0.6](https://github.com/jodevsa/wireguard-operator/compare/v1.0.5...v1.0.6) (2023-05-12)


### Bug Fixes

* fix release.yaml ([bcc4a2c](https://github.com/jodevsa/wireguard-operator/commit/bcc4a2c9485446650ad77d96b9de09b064432c85))
* imporvements to workflows ([9ccdfa3](https://github.com/jodevsa/wireguard-operator/commit/9ccdfa383ab1dca8ea55b947aa4c089649480e91))
* remove debug mode ([1356019](https://github.com/jodevsa/wireguard-operator/commit/13560190e9faa2569b4b75ecd1a896bff3497de5))
* workfllows ([fa449a1](https://github.com/jodevsa/wireguard-operator/commit/fa449a13fd283e0142420987f407531e16146660))
* workflow ([3361bfa](https://github.com/jodevsa/wireguard-operator/commit/3361bfa81d44c886e320d380b4a5cbc2aa43c060))

## [1.0.5](https://github.com/jodevsa/wireguard-operator/compare/v1.0.4...v1.0.5) (2023-05-12)


### Bug Fixes

* fix image version ([918d459](https://github.com/jodevsa/wireguard-operator/commit/918d45947f7b3ce98c00107fdbd758f7162fd877))

## [1.0.4](https://github.com/jodevsa/wireguard-operator/compare/v1.0.3...v1.0.4) (2023-05-12)


### Bug Fixes

* additions ([c98a08d](https://github.com/jodevsa/wireguard-operator/commit/c98a08dd0f4008211964e8bc9e502fdfb8b3ba44))

## [1.0.3](https://github.com/jodevsa/wireguard-operator/compare/v1.0.2...v1.0.3) (2023-05-12)


### Bug Fixes

* build images on tag push ([9d12662](https://github.com/jodevsa/wireguard-operator/commit/9d126625ff35f0b2b37659333b8a5e0eb24883ed))

## [1.0.2](https://github.com/jodevsa/wireguard-operator/compare/v1.0.1...v1.0.2) (2023-05-12)


### Bug Fixes

* add debug mode ([19998b5](https://github.com/jodevsa/wireguard-operator/commit/19998b5252d0a821cfe91b410316b34a4d8feb08))
* add more permissions ([2440f8d](https://github.com/jodevsa/wireguard-operator/commit/2440f8d7580928797f9496a70ec8eae5fe0665d8))
* fix semantic-release ([8a4b13c](https://github.com/jodevsa/wireguard-operator/commit/8a4b13cf996d26703bbafb3cfc9b02d46b001ddf))
* install semantic-release plugins in release job ([4060a84](https://github.com/jodevsa/wireguard-operator/commit/4060a84f839381b5dfa7c392e44052fc8c1fbcaf))
* Update README.md ([9f4da9f](https://github.com/jodevsa/wireguard-operator/commit/9f4da9fe95ff614e24d868d3a4a251dabec85b2c))
* use semantic-release directly ([74658ab](https://github.com/jodevsa/wireguard-operator/commit/74658ab502b07f81a027f3c966a8fba7d24c5ca6))

## [1.0.1](https://github.com/jodevsa/wireguard-operator/compare/v1.0.0...v1.0.1) (2023-05-12)


### Bug Fixes

* upload release file ([77c0c21](https://github.com/jodevsa/wireguard-operator/commit/77c0c210cc69e9e83efe87ad7bbe78ac38cf56b1))

# 1.0.0 (2023-05-12)


### Features

* add semantic-release config: ([2371f8e](https://github.com/jodevsa/wireguard-operator/commit/2371f8e6f2f425ec62ae8403de39e2bb45335b8f))
