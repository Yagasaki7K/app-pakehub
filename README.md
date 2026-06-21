# App Pake Builder

Automatically build desktop applications from websites using Pake and GitHub Actions.

## What is Pake?

Pake is an open-source tool that transforms any website into a lightweight native desktop application powered by Tauri.

Official project:

https://github.com/tw93/Pake

Benefits:

- Small executable size
- Low memory usage
- Native performance
- Windows support (.msi)
- Linux support (.deb)
- macOS support (.app and .dmg)

## Features

This repository automatically:

- Builds Windows installers (.msi)
- Builds Linux packages (.deb)
- Builds macOS applications (.app)
- Builds macOS installers (.dmg)
- Publishes GitHub Releases
- Stores generated installers inside `/projects`

## How It Works

Every workflow execution:

1. Builds the application with Pake.
2. Collects generated installers.
3. Stores them in `/projects`.
4. Creates a GitHub Release.
5. Commits updated installers back to the repository.

## Contributing

Contributions are welcome.

You can help by:

- Reporting bugs
- Improving workflows
- Improving documentation
- Adding support for additional packaging targets
- Testing on different operating systems

## License

This project is distributed under the MIT License.
