# Changelog

## [1.0.6](https://github.com/dklesev/terraform-provider-pihole/compare/v1.0.5...v1.0.6) (2026-03-20)


### Miscellaneous

* Automate Go version updates via a new workflow and update Go and module dependencies, aligning CI to `go.mod`. ([1f68d15](https://github.com/dklesev/terraform-provider-pihole/commit/1f68d15426be86e7df0fb4d4da07f38012756bfa))
* **deps:** bump github.com/hashicorp/terraform-plugin-testing ([#17](https://github.com/dklesev/terraform-provider-pihole/issues/17)) ([50da5bc](https://github.com/dklesev/terraform-provider-pihole/commit/50da5bc665ab21e2f63dc0669b2f29d2673b9d72))
* Update GOLANGCI_LINT_VERSION to v2.11.3 in test workflow. ([00d36f1](https://github.com/dklesev/terraform-provider-pihole/commit/00d36f1ce3ff80e92e9869809e0e624f77f37a59))

## [1.0.5](https://github.com/dklesev/terraform-provider-pihole/compare/v1.0.4...v1.0.5) (2026-03-10)


### Miscellaneous

* **deps:** bump crazy-max/ghaction-import-gpg from 6.3.0 to 7.0.0 ([#14](https://github.com/dklesev/terraform-provider-pihole/issues/14)) ([c2b6931](https://github.com/dklesev/terraform-provider-pihole/commit/c2b6931f296f351dff0a9320868cd298795bf1ca))

## [1.0.4](https://github.com/dklesev/terraform-provider-pihole/compare/v1.0.3...v1.0.4) (2026-03-02)


### Miscellaneous

* **deps:** bump actions/setup-go from 6.2.0 to 6.3.0 ([#10](https://github.com/dklesev/terraform-provider-pihole/issues/10)) ([f0ccbf0](https://github.com/dklesev/terraform-provider-pihole/commit/f0ccbf0fd2e929dae5d58499f9fcfc5272414d3d))
* **deps:** bump github.com/hashicorp/terraform-plugin-framework ([#12](https://github.com/dklesev/terraform-provider-pihole/issues/12)) ([40e1ae5](https://github.com/dklesev/terraform-provider-pihole/commit/40e1ae51d50bed179e714a52ef0dadfcfaa0bd00))
* **deps:** bump github.com/hashicorp/terraform-plugin-go ([#11](https://github.com/dklesev/terraform-provider-pihole/issues/11)) ([05110ed](https://github.com/dklesev/terraform-provider-pihole/commit/05110ed48e9590116610b03ba07491e7a905e79a))

## [1.0.3](https://github.com/dklesev/terraform-provider-pihole/compare/v1.0.2...v1.0.3) (2026-02-26)


### Miscellaneous

* **deps:** bump github.com/cloudflare/circl ([#9](https://github.com/dklesev/terraform-provider-pihole/issues/9)) ([9b2960b](https://github.com/dklesev/terraform-provider-pihole/commit/9b2960bdd9833ac14a59956b6ff59a41a3452eee))
* **deps:** bump goreleaser/goreleaser-action from 6.4.0 to 7.0.0 ([#7](https://github.com/dklesev/terraform-provider-pihole/issues/7)) ([a3cebeb](https://github.com/dklesev/terraform-provider-pihole/commit/a3cebeb3a8efb7cca8120c256882d359b37170d1))

## [1.0.2](https://github.com/dklesev/terraform-provider-pihole/compare/v1.0.1...v1.0.2) (2026-01-31)


### Miscellaneous

* **deps:** bump actions/checkout from 6.0.1 to 6.0.2 ([#5](https://github.com/dklesev/terraform-provider-pihole/issues/5)) ([9235cc2](https://github.com/dklesev/terraform-provider-pihole/commit/9235cc2df80021fb6fab5bc91e332a693cedf6c8))

## [1.0.1](https://github.com/dklesev/terraform-provider-pihole/compare/v1.0.0...v1.0.1) (2026-01-19)


### Miscellaneous

* **deps:** bump actions/setup-go from 6.1.0 to 6.2.0 ([#3](https://github.com/dklesev/terraform-provider-pihole/issues/3)) ([9a4fac5](https://github.com/dklesev/terraform-provider-pihole/commit/9a4fac5e1f6c28964a5e1893762aef4027c74b3b))

## 1.0.0 (2025-12-27)


### Miscellaneous

* init ([d004193](https://github.com/dklesev/terraform-provider-pihole/commit/d004193822a302bf0f7d29dab0a43d4e3404fc22))

## [1.2.0](https://github.com/dklesev/terraform-provider-pihole/compare/v1.1.0...v1.2.0) (2025-12-27)


### Features

* update README to reflect v1.0 features, expanded resource coverage, and detailed configuration options. ([ea43913](https://github.com/dklesev/terraform-provider-pihole/commit/ea4391327f8f6bb8490f46e44d12741878a26826))

## [1.1.0](https://github.com/dklesev/terraform-provider-pihole/compare/v1.0.0...v1.1.0) (2025-12-27)


### Features

* Introduce Terraform resources for Pi-hole configuration management including DNS, DHCP, resolver, and webserver settings. ([0f84f82](https://github.com/dklesev/terraform-provider-pihole/commit/0f84f82a8ec92898fbe25430cba0d66d16a09280))
* Update client group references to use IDs, improve client update logic, and add new client-side tests. ([808519a](https://github.com/dklesev/terraform-provider-pihole/commit/808519ac6a5a61437c41b1700f1d2f7e52486eba))


### Miscellaneous

* add copyright and SPDX license identifier to tools file ([fb351bf](https://github.com/dklesev/terraform-provider-pihole/commit/fb351bfbf79fdeab7663bb66f9fdee5534f5f86b))
* Update license identifier to MIT and adjust example group references to use IDs instead of names. ([47ebd45](https://github.com/dklesev/terraform-provider-pihole/commit/47ebd45838dd3cb3297c71bb95ede164522e72a3))

## 1.0.0 (2025-12-25)


### Miscellaneous

* init ([2025142](https://github.com/dklesev/terraform-provider-pihole/commit/202514279511eadf5be954c7af13db57e64c3b40))
