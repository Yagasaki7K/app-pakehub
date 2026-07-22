<p align="center">
    <img src=/hub.png width=150/>
</p>
<h1 align="center">Pake Application Hub</h1>
<p align="center"><strong>Turn any webpage into a desktop app with Github Actions via Pake, supports Windows, MacOS and Linux</strong></p>

Pake Application Hub is an open-source automation repository that builds and publishes installable desktop wrappers for selected web applications. It is created and maintained by **Yagasaki7K**.

The repository combines GitHub Actions and Pake.

## Project Overview

This project provides a repeatable release pipeline for web applications that are useful as desktop apps. Each supported application has its own workflow under `.github/workflows` and each workflow builds native installer artifacts for Windows and Linux.

Final generated installers are stored in `/projects`. Temporary workflow files are stored in `/artifacts` and are not part of the published application catalog.

![](https://raw.githubusercontent.com/tw93/static/main/pake/YouTube.jpg)

## Features

- One GitHub Actions workflow per application.
- Windows `.msi`, MacOS `.dmg` and Linux `.deb` output validation.
- Automatic cleanup and replacement of outdated installers for the application being rebuilt.
- GitHub Release publication for generated installers.
- Dependency-free SPA with responsive desktop and mobile design.
- Automatic frontend discovery from `/projects` through a generated manifest.
- Search, platform filters, download buttons, contribution guidance, and project information sections.
- Professional contributor documentation for requesting new applications.

## What is Pake

[Pake](https://github.com/tw93/Pake) is an open-source project that turns web pages into desktop applications with a command-line workflow. The Pake project describes itself as a way to turn any webpage into a desktop app with one command and supports Windows and Linux.

Pake is built around the Tauri ecosystem, which uses Rust and native webview capabilities to package web experiences as desktop applications. This repository uses the Pake CLI to wrap public web application URLs into native desktop installers.

## How Pake Works

At a high level, Pake receives a web URL and an application name, creates a desktop application shell for that URL, and builds platform-specific installable outputs. In this repository, every application workflow hardcodes the verified public URL directly in the workflow build command.

Example workflow build command:

```bash
pake "https://chatgpt.com" --name "ChatGPT"
```

Linux builds explicitly request Debian output:

```bash
pake "https://chatgpt.com" --name "ChatGPT" --targets deb
```

## Why Pake Is Used

Pake is used because it is designed for packaging websites as lightweight desktop apps. It allows this repository to maintain simple, URL-driven application workflows while still producing native installer artifacts for multiple operating systems.

## Supported Platforms

| Platform | Runner | Output |
| --- | --- | --- |
| Windows | `windows-latest` | `.msi` |
| Linux | `ubuntu-latest` | `.deb` |

![](https://raw.githubusercontent.com/tw93/static/main/pake/ChatGPT.png)

## Available Applications

The web catalog only displays applications that have generated installer files in `release` and you can see better [in releases](https://github.com/Yagasaki7K/app-pakehub/releases).

## Installation

End users can download generated installers from the SPA catalog or from GitHub Releases after workflows have produced artifacts.

### Windows

1. Download the application `.msi` file.
2. Open the installer.
3. Follow the Windows installation prompts.

### Linux

1. Download the application `.deb` file.
2. Install it with apt:

```bash
sudo apt install ./application.deb
```

### macOS

1. Download the application `.dmg` file.
2. Open the `.dmg` file.
3. Drag the application icon to the `Applications` folder.
4. Launch the app from `Applications` (you may need to right-click and select "Open" the first time to bypass Gatekeeper).

> [!IMPORTANT]
> **Security Notice:** The generated installers are not officially signed by Microsoft or Apple. As a result, your system may display a security warning or antivirus alert when you try to run them. This happens **solely** because of the lack of digital certification — **not** because the software is malicious. All packages are built directly from the official application URL provided in `apps.json`, using [Pake](https://github.com/tw93/Pake), a widely used open-source tool. To install, you may need to:
>
> - On Windows: click **"More info"** and then **"Run anyway"**.
> - On macOS: right-click the application and select **"Open"**.
>
> Only download files from trusted sources (this official repository) to ensure integrity.

## Local Development

Install dependencies:

```bash
npm install
```

Generate the local projects manifest:

```bash
node scripts/generate-projects-manifest.mjs
```

Run the SPA locally:

```bash
npm run dev
```

## Running the Web Version

The web version is a dependency-free SPA served by the repository's Node.js development server. During startup and production builds, it reads `public/projects-manifest.json`, which is generated from the contents of `/projects`.

If `/projects` has no installer files, the SPA still runs and shows an empty-state message explaining that workflows must generate installers first.

## Building the Frontend

Build the production frontend:

```bash
npm run build
```

The build command first runs `scripts/generate-projects-manifest.mjs`, then creates the static site in `dist`.

## GitHub Actions Overview

Each application has its own workflow named with this pattern:

```text
build-application-name.yaml
```

Examples:

```text
build-discord.yaml
build-chatgpt.yaml
build-spotify.yaml
```

Each workflow:

1. Builds Windows `.msi` output on a Windows runner.
2. Builds Linux `.deb` output on an Ubuntu runner.
3. Uploads temporary build outputs through GitHub Actions artifacts.
4. Downloads those outputs in a publish job.
5. Copies final outputs into `/projects`.
6. Validates required installer types.
7. Commits changed generated installers.
8. Publishes a GitHub Release.

## Workflow Validation

This repository includes validation utilities for release maintenance:

```bash
npm run audit:workflows
npm run test:artifact-flow
npm test
```

The workflow audit checks every application workflow for hardcoded URLs and Pake project names, verifies that Pake is installed with npm instead of pnpm, and confirms that upload, download, validation, and release steps reference valid repository scripts. The artifact-flow test simulates generated installer files, collects them through the same shell scripts used in CI, copies them into `/projects`, and validates that temporary files remain under `/artifacts`.

## Release Process

Application workflows are manually runnable with `workflow_dispatch` and also run on a monthly schedule. A successful release job updates `/projects` for the application being built, commits changed generated outputs, and creates a release tagged with the application slug and workflow run number.

Release assets are uploaded from `/projects` using the generated `.msi`, `.deb`, and `.dmg` files.

## Contributing

Contributions are welcome. Keep changes focused, use English comments, and preserve the repository contracts:

- `/projects` is for final generated installers and application bundles.
- `/artifacts` is for temporary workflow output only.
- Application URLs and Pake project names must be hardcoded directly in each workflow.
- Do not add URL-related or application-name variables to `.env` or workflow-level application variables.
- Do not create a generic application build workflow.

Aqui estão as alterações solicitadas para o README.md, incluindo as instruções de instalação para macOS, um aviso sobre certificação e a atualização da seção "Creating New Applications" para refletir a nova arquitetura baseada em `apps.json`.

---

## Creating New Applications

This repository uses a **single-source-of-truth** approach. To add or update an application, you **do not** create a new workflow file. Instead, you edit the `apps.json` file, and the [orchestrator](.github/workflows/orchestrator.yaml) will automatically build all apps listed there.

### How to add a new application

1. Fork the repository.
2. Create a branch for your change.
3. Verify the public URL for the application.
4. Open `apps.json` and add a new entry with the following format:
   ```json
   {
     "name": "Application Name",
     "url": "https://example.com"
   }
   ```
5. Commit and push your branch.
6. Open a Pull Request.

### How the build system works

- The [`orchestrator.yaml`](.github/workflows/orchestrator.yaml) workflow is triggered on pushes to `apps.json` or manually via `workflow_dispatch`.
- It reads `apps.json`, generates a dynamic matrix, and triggers the reusable [`build-app.yaml`](.github/workflows/build-app.yaml) for each application.
- The `build-app.yaml` workflow builds installers for:
  - **Windows** (`.msi`)
  - **Linux** (`.deb`)
  - **macOS** (`.dmg`)
- All artifacts are collected, normalized (slug based on the application name), and committed to the `projects/` directory.

### Important notes

- **No manual workflow creation** – you never need to create `build-*.yaml` files.
- **No `slug` field in `apps.json`** – the slug is generated automatically from the `name` field.
- **No extra configuration** – the system handles everything from a single JSON file.
- The `projects/` directory is automatically updated with the latest builds on every successful orchestration run.

### Example `apps.json`

```json
[
  {
    "name": "ChatGPT",
    "url": "https://chatgpt.com"
  },
  {
    "name": "Twitter / X",
    "url": "https://x.com"
  },
  {
    "name": "Disney+",
    "url": "https://disneyplus.com"
  }
]
```

> The application name is used to generate the slug (e.g., `"Twitter / X"` becomes `twitter-x`). The slug is used for artifact filenames and directory names.

## Pull Request Guidelines

A Pull Request should include:

- The application name.
- The verified public URL.
- The new workflow path.
- Any metadata changes.
- Validation commands and results.
- A clear explanation of why the application is useful as a Pake desktop app.

## License

This project is licensed under the MIT License. See [`LICENSE`](LICENSE) for details.

## Credits

- Created and maintained by **Yagasaki7K**.
- Built with [Pake](https://github.com/tw93/Pake).
- Frontend powered by modern browser APIs and Node.js build scripts.
- Automation powered by GitHub Actions.

![](https://raw.githubusercontent.com/tw93/static/main/pake/Excalidraw.png)
