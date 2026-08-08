# GitHub Actions Workflow Analysis

## Executive Summary

1. **The current orchestrator correctly centralizes application selection, but the build fan-out is much larger than it first appears.** With 44 non-smoke apps and three platforms per app, a full run creates 132 package builds plus three smoke builds. The orchestrator caps concurrent reusable workflow calls, but each called workflow still runs three platform jobs in parallel.
2. **Release automation works only when every app/platform succeeds and currently publishes MSI, DEB, and DMG assets without checksums, AppImage assets, or per-platform debug bundles.** This creates an all-or-nothing release path and makes post-failure diagnosis harder.
3. **Caching is underdeveloped for the actual workload.** Rust caching is enabled, but Node/npm package reuse, Pake-generated project caches, Linux apt caching, and deterministic cache key controls are missing or minimal.

## 1. Workflow Structure Analysis

### Triggers

- `orchestrator.yaml` runs on pushes that touch `apps.json`, workflow files, or local composite actions.
- `orchestrator.yaml` supports manual `workflow_dispatch` with `max_parallel`.
- `orchestrator.yaml` runs weekly on Sunday at 03:00 UTC.
- `build-app.yaml` can run manually with app metadata inputs.
- `build-app.yaml` is reusable via `workflow_call` and is invoked by the orchestrator.
- There is no `pull_request` trigger, so workflow changes are not validated before merge unless run manually or pushed to a branch covered by `push`.

### Job dependency graph

```text
validate
  ├─ smoke-test (conditional reusable workflow)
  └─ build (waits for validate and smoke-test)
        └─ release (only if aggregate build succeeds)
validate + smoke-test + build + release
  └─ status-report (always)
```

Inside each `build-app.yaml` call:

```text
build-platform[windows, linux, macos]
```

### Parallelism observations

- App-level concurrency is constrained by `max_parallel`, but platform-level jobs inside each reusable workflow run independently. Effective maximum platform concurrency is approximately `max_parallel * 3` plus smoke jobs.
- `fail-fast: false` is appropriate for package matrices because it preserves signal across platforms and apps.
- The smoke test is a gate for every app build. This improves safety, but it also makes external network availability to `https://google.com` a single point of failure.

### Orchestrator reliability assessment

The orchestrator pattern is sound for a large, repetitive Pake build fleet because app metadata lives in one JSON file and the reusable workflow handles platform packaging consistently. Its main reliability risks are:

- app matrix output size limits when the app list grows significantly;
- all-or-nothing global release behavior;
- reliance on live web URLs during package generation;
- reusable workflow nesting that obscures the true job count and queue pressure;
- no pull-request validation path.

## 2. Dynamic Matrix Generation

### Current behavior

The `validate` job loads `apps.json`, requires a non-empty `.apps` array, validates each `name` and `url`, slugifies names, rejects duplicate slugs, extracts a special `__smoke__` app, and writes `apps`, `app_count`, and `smoke` outputs.

### Potential failure points

- JSON parsing failure stops the workflow without a schema-specific diagnostic.
- `except:` in the slug normalization block catches all exceptions. Use `except Exception as error` or avoid a broad catch.
- Matrix output is written as one line to `$GITHUB_OUTPUT`; very large app lists can hit GitHub output or matrix-size constraints.
- Slugs are ASCII-only, so non-Latin names may collapse to `app` and collide.
- `__smoke__` is excluded by exact name only; a duplicate slug such as `smoke` from another app is still possible because the smoke slug is not included in the non-smoke list.
- URL validation checks shape, not reachability, TLS health, redirects, or whether Pake can package the site.

### Large-list handling

For 44 non-smoke apps and three platforms, the current workflow schedules 132 package jobs. GitHub Actions matrices also have hard limits, and output payloads are not an ideal transport for very large manifests.

Recommended strategies:

```yaml
strategy:
  fail-fast: false
  max-parallel: ${{ inputs.max_parallel || vars.MAX_PARALLEL || 4 }}
  matrix:
    shard: [0, 1, 2, 3]
```

Then generate one app slice per shard. This keeps each matrix smaller and provides a recovery path for rerunning a subset.

Alternatively, generate a combined app/platform matrix in the orchestrator and call a reusable workflow for exactly one app/platform pair. That makes global concurrency explicit and easier to cap.

## 3. Build Matrix Optimization

### Current platform coverage

- Windows: `windows-latest`, MSI only.
- Linux: `ubuntu-latest`, DEB only.
- macOS: `macos-latest`, DMG only.

### Gaps vs target platforms and formats

- Target Linux includes `.AppImage`, but the workflow only collects `.deb`.
- Target macOS includes `.app`, but the workflow only uploads `.dmg`.
- Target Windows includes Windows 10/11, but `windows-latest` tracks the hosted image selected by GitHub, not a pinned OS version.
- Target Ubuntu is 22.04+, but `ubuntu-latest` can move over time. Pin `ubuntu-22.04` or test `ubuntu-22.04` plus `ubuntu-24.04` deliberately.
- Target macOS is 11+, but `macos-latest` can move and may imply different CPU architecture defaults over time.
- macOS ARM64 is not explicitly built.

### Suggested matrix shape

For release builds:

```yaml
matrix:
  platform:
    - { name: windows-x64, runner: windows-2022, ext: msi }
    - { name: linux-x64-deb, runner: ubuntu-22.04, ext: deb }
    - { name: linux-x64-appimage, runner: ubuntu-22.04, ext: AppImage }
    - { name: macos-x64, runner: macos-13, ext: dmg }
    - { name: macos-arm64, runner: macos-14, ext: dmg }
```

For pull requests, use a reduced validation matrix:

```yaml
on:
  pull_request:
    paths:
      - apps.json
      - .github/workflows/**
      - .github/actions/**

jobs:
  validate:
    # same JSON/schema validation
  smoke-linux:
    # build one smoke app on ubuntu-22.04 only
```

## 4. Caching Strategy

### Current caching

- Node.js is installed but npm cache is not enabled.
- Rust stable is installed and `Swatinem/rust-cache@v2` is used.
- There is no pnpm usage in current workflow files.
- There is no explicit Cargo registry/git cache path tuning.
- There is no apt package cache.
- Pake is run through `npx --yes pake-cli@version`, which can redownload package metadata frequently.

### Improvements

Enable npm cache and pin a cache dependency path if a lockfile exists:

```yaml
- name: Setup Node.js
  uses: actions/setup-node@v5
  with:
    node-version: 22
    cache: npm
```

If Pake does not require per-run global install isolation, install once per job after Node setup:

```yaml
- name: Install Pake CLI
  shell: bash
  run: npm install --global pake-cli@${{ inputs.pake-version }}
```

Improve Rust cache separation:

```yaml
- name: Cache Rust dependencies
  uses: Swatinem/rust-cache@v2
  with:
    shared-key: pake-${{ runner.os }}-${{ runner.arch }}
    save-if: ${{ github.ref == 'refs/heads/main' }}
```

Consider apt caching only if Linux setup time is material and cache complexity is justified.

## 5. Security Audit

### Positive findings

- Top-level default permission in both workflows is `contents: read`.
- Release job escalates only to `contents: write`.
- Inputs are quoted in shell commands where app name, slug, URL, and version are used.
- No hardcoded tokens or credentials were found in the workflow files reviewed.

### Risks

- Third-party actions are pinned by tag, not full commit SHA (`actions/checkout`, `actions/setup-node`, `dtolnay/rust-toolchain`, `Swatinem/rust-cache`, upload/download artifact). Tag pinning is convenient but weaker than SHA pinning.
- `npx --yes pake-cli@$PAKE_VERSION` downloads executable code at build time. The Pake version is pinned by input value, but npm transitive dependencies can still be a supply-chain risk unless lockfile-driven installation or provenance controls are used.
- `PAT_TOKEN || GITHUB_TOKEN` uses a personal token if present. A PAT usually has broader blast radius than the ephemeral GitHub token. Prefer a fine-grained PAT only if cross-repository writes are required.
- Weekly scheduled builds package arbitrary external websites. A compromised or hostile web source could influence generated app content.
- Release assets are not checksummed in the global release path.

### Recommended permission model

```yaml
permissions:
  contents: read

jobs:
  release:
    permissions:
      contents: write
```

This is already mostly implemented. Add an explicit comment explaining why write access is necessary and avoid PAT fallback unless required.

## 6. Artifact Management

### Current behavior

- Each platform uploads one artifact named `${slug}-${platform}`.
- The release job downloads all artifacts and copies MSI, DEB, and DMG files into `release-assets`.
- Smoke artifacts are filtered by artifact directory name.
- The global release creates one release tagged `pakehub-X.Y.Z`.

### Issues

- Artifact filenames are `${slug}.${ext}`, so files from different platforms are distinguishable only by extension. This breaks down if multiple Linux formats are added or multiple architectures share one extension.
- No retention period is specified; default retention may be longer or shorter than intended.
- No SHA256 checksums are generated for the global release.
- `.AppImage` and `.app` are not collected.
- Release notes list apps but not exact artifact names, platform statuses, sizes, or checksums.
- macOS code signing/notarization is not implemented.
- Windows code signing is not implemented.
- The `prepare-release` and `publish-release` composite actions appear partly unused by the current orchestrator global release path.

### Suggested artifact naming

```yaml
with:
  name: ${{ inputs.slug }}-${{ matrix.platform.name }}-${{ github.run_number }}
  path: ${{ github.workspace }}/artifacts/${{ inputs.slug }}-${{ matrix.platform.name }}/*
  if-no-files-found: error
  retention-days: 14
```

Suggested final asset name:

```text
${slug}-${version}-${platform}-${arch}.${ext}
```

## 7. Error Handling & Debugging

### Strong points

- Build commands have retry logic for transient network-like errors.
- Generated artifact size is validated.
- Linux DEB files are checked with `file` and `dpkg-deb` when available.
- macOS DMG files are mounted to validate readability.
- Release creation retries conflicts and transient GitHub/network failures.

### Weak points

- Debug logs and generated intermediate Pake projects are deleted because the build directory is removed on exit.
- The retry wrapper also wraps commands such as `unzip`, where retries rarely help.
- The smoke site is Google, which can be rate limited, regionally altered, or blocked by bot detection.
- Status reporting requires smoke success; if no smoke app exists, `smoke-test` is skipped and the report currently treats that as not fully successful.
- The release notes are passed through a GitHub output and then shell interpolation. Prefer `--notes-file` to avoid escaping surprises.

### Debug preservation example

```yaml
- name: Upload debug bundle on failure
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: debug-${{ inputs.slug }}-${{ matrix.platform.name }}
    path: |
      ${{ github.workspace }}/pake-debug/**
      ${{ github.workspace }}/artifacts/**
    if-no-files-found: ignore
    retention-days: 7
```

## 8. Optimization Recommendations

### Build time reduction

- Pin runner images to avoid surprise updates and cache fragmentation.
- Build a single Linux smoke package on PRs and reserve full matrix builds for scheduled/manual/release-triggered runs.
- Move from nested app workflow calls to a single app/platform matrix if queue visibility and concurrency control are more important than workflow reuse boundaries.
- Cache npm package downloads and Rust artifacts per OS/architecture.
- Avoid rebuilding unchanged apps by comparing changed app entries or using a manual input to select app slugs.

### Resource optimization

- Add manual inputs such as `app_slug`, `platform`, and `release_enabled`.
- Use schedule for canary/smoke validation, not necessarily full rebuilds of all apps.
- Add `timeout-minutes` per expensive setup step or split Linux dependency install from packaging diagnostics.
- Use explicit `ubuntu-22.04`, `windows-2022`, and specific macOS runners.

### Dependency update automation

- Enable Dependabot for GitHub Actions and npm.
- Use Renovate or Dependabot grouping for workflow actions.
- Add a CI workflow linter such as `actionlint`.

### Cost optimization

- macOS minutes are the most expensive; build macOS only for releases, on manual demand, or for changed apps.
- Consider separate scheduled smoke and release workflows.
- Add `workflow_dispatch` inputs to skip platforms.
- Lower full-build cadence if generated assets do not need weekly refresh.

## Priority Matrix

| Priority | Recommendation | Impact | Effort |
| --- | --- | --- | --- |
| Critical | Add PR validation for workflow/app metadata changes | Prevent broken workflows from reaching main | S |
| Critical | Add checksums to global releases | Improves release integrity and user trust | S |
| High | Pin runner images and third-party actions more strictly | Reduces supply-chain and reproducibility risk | M |
| High | Add npm cache and refine Rust cache keys | Reduces repeated setup/build time | S |
| High | Support AppImage and optional `.app` artifacts | Aligns outputs with stated distribution targets | M |
| High | Preserve debug artifacts on failure | Speeds incident triage | S |
| Medium | Convert to explicit app/platform matrix or shards | Improves concurrency control at scale | M/L |
| Medium | Add app/platform manual filters | Reduces unnecessary build minutes | M |
| Medium | Replace Google smoke with a deterministic minimal local/static test URL if possible | Reduces false negatives | M |
| Low | Consolidate or remove unused release composite actions | Reduces maintenance overhead | S |

## Actionable Next Steps

1. **Add PR validation and `actionlint`** — estimated effort: 1-2 hours.
2. **Add global release SHA256 generation and upload** — estimated effort: 1 hour.
3. **Enable npm cache and configure Rust cache keys** — estimated effort: 1 hour.
4. **Add debug artifact upload on failure** — estimated effort: 1 hour.
5. **Pin runner images and decide macOS architecture policy** — estimated effort: 2-4 hours.
6. **Add AppImage collection and architecture-aware artifact naming** — estimated effort: 4-8 hours.
7. **Introduce app/platform filters or sharded matrix generation** — estimated effort: 0.5-2 days.
8. **Evaluate signing/notarization strategy** — estimated effort: 1-3 days depending on certificates and distribution policy.
