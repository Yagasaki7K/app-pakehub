<p align="center">
<img src="https://raw.githubusercontent.com/tw93/Pake/main/assets/logo.png" width="200" />
</p>

# Pake Application Hub

Pake Application Hub is a community catalog of desktop applications generated from web applications. The project packages useful websites as installable Windows, Linux, and macOS applications, so they can be opened from the desktop without keeping another browser tab pinned.

The project is intentionally simple: `apps.json` is the catalog, GitHub Actions builds the applications, and the generated installers are published in GitHub Releases and exposed by the web catalog.

## What is Pake?

[Pake](https://github.com/tw93/Pake) is an open-source command-line tool that turns a web URL into a native desktop application. It uses the Tauri ecosystem and the operating system's webview, so the result behaves like a focused desktop window while still loading the original website.

This repository uses Pake CLI `3.18.0`. It does not recreate or modify the services being packaged; each application continues to load its official public URL.

## How the project works

1. An application is added to [`apps.json`](apps.json) with a display name and public URL.
2. The [`orchestrator`](.github/workflows/orchestrator.yaml) validates the catalog and creates a build matrix.
3. The reusable [`build-app`](.github/workflows/build-app.yaml) workflow builds Windows (`.msi`), Linux (`.deb`), and macOS (`.dmg`) installers on their native GitHub-hosted runners.
4. Builds run in parallel with a bounded concurrency level, so the workflow makes steady progress without exhausting the repository's GitHub Actions runner quota. The default is six applications at a time; a manual run can override it with `max_parallel`, or the repository variable `MAX_PARALLEL` can define a default.
5. Validated artifacts are collected and published in a global GitHub Release. The catalog manifest is generated from the files in `projects/`.

## Supported platforms

| Platform | Output |
| --- | --- |
| Windows | `.msi` |
| Linux | `.deb` |
| macOS | `.dmg` |

## Using the applications

Download an installer from the [GitHub Releases](https://github.com/Yagasaki7K/app-pakehub/releases) page or from the web catalog when an application is available.

- Windows: open the `.msi` installer.
- Linux: install a `.deb` with `sudo apt install ./application.deb`.
- macOS: open the `.dmg` and move the application to `Applications`.

The generated packages are not signed with Microsoft or Apple certificates. Windows, macOS, or an antivirus product may therefore display a warning. Only download releases from this repository and verify that the URL in `apps.json` is the official service URL.

## Add or update an application

No application-specific workflow is needed. Add an entry to `apps.json`:

```json
{
  "name": "Application Name",
  "url": "https://example.com"
}
```

The slug used for artifact names is generated from `name`; do not add a `slug` field. After opening a pull request, the catalog validation and the orchestrator can be run from GitHub Actions. Please include the reason the application is useful as a desktop app and confirm that the URL is official.

## Local development

The repository contains a dependency-free static web catalog and scripts compatible with Bun. Install [Bun](https://bun.sh) and run:

```bash
bun scripts/generate-projects-manifest.mjs
bun scripts/dev-server.mjs
```

Then open <http://localhost:5173>. The frontend reads generated installers from `projects/` and writes the catalog manifest to `public/projects-manifest.json`.

## GitHub Actions maintenance

The Pake version is defined consistently in the reusable workflow, the setup action, and the orchestrator. When upgrading it, update all three defaults and run a manual application build before building the complete catalog.

The build workflow uses native runners, retries transient network failures, caches Bun packages, validates each installer, and uploads one artifact per application and platform. This separation keeps platform builds independent and prevents one slow runner from blocking the other platforms for the same application.

## License and credits

This project is licensed under the [MIT License](LICENSE).

Created and maintained by **Yagasaki7K**. Built with [Pake](https://github.com/tw93/Pake) and [GitHub Actions](https://github.com/features/actions).
