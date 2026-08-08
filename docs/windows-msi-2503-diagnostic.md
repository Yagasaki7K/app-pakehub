# Windows MSI Error 2503 Diagnostic and Workflow Fix

## Diagnostic Report

### Root cause assessment

The current Windows workflow had three issues that can plausibly produce broken or incomplete MSI output and make error 2503 much harder to diagnose:

1. **WiX and NSIS were not installed or cached explicitly before invoking Pake.** Tauri uses WiX v3 for MSI packages and NSIS for setup executables, so relying on implicit downloader behavior during `npx pake-cli` adds a network and path-sensitive failure point.
2. **The Windows runner and shell environment were not pinned tightly enough.** `windows-latest` can change under the workflow, and the build step did not force `COMSPEC`/`ComSpec` to `cmd.exe` before invoking the Tauri/Pake bundler stack.
3. **MSI validation was too shallow.** The previous workflow checked only that the file existed and had non-zero size. A non-empty MSI can still have an invalid Windows Installer database or broken cabinet stream.

### Findings against requested causes

| Requested check | Applies? | Finding | Fix |
| --- | --- | --- | --- |
| WiX Toolset missing | Yes | No Windows-specific WiX installation step existed. | Add cached install of `wix314-binaries.zip` before build. |
| NSIS missing | Yes | No Windows-specific NSIS installation step existed. | Add cached install of `nsis-3.zip` before build. |
| Tool downloads timing out or using the wrong path | Yes | The workflow delegated tool acquisition to Pake/Tauri at build time. | Pre-populate `%LOCALAPPDATA%\\tauri\\WixTools` and `%LOCALAPPDATA%\\tauri\\NSIS`, then add both to `PATH`. |
| COMSPEC wrong | Possible | The build step did not set `COMSPEC`/`ComSpec`. | Force both variables to `C:\\Windows\\System32\\cmd.exe`. |
| Output path mismatch | Partially mitigated already | The existing workflow used broad `find` discovery, so it was not limited to `src-tauri/target/release/bundle/msi`. | Keep broad discovery and add debug bundle preservation. |
| Hardcoded installer paths | Unknown from workflow only | No installer table inspection existed, so path-related installer failures would surface only on user machines. | Add `msiexec /a` administrative validation with verbose logging. |
| Visual Studio Build Tools / Windows SDK | Possible | Hosted images usually include them, but the workflow did not fail early if missing. | Add explicit verification for VC++ build tools and Windows SDK `10.0.19041.0+`. |

## Fixed workflow implementation

The fix is split into a reusable composite action and updates to the Pake build workflow:

- `.github/actions/install-windows-deps/action.yaml` installs/caches WiX and NSIS, normalizes COMSPEC, and verifies Visual Studio/Windows SDK prerequisites.
- `.github/workflows/build-app.yaml` pins Windows to `windows-2022`, invokes the Windows dependency action before the build, preserves debug files on failure, validates the MSI container, runs an administrative MSI validation, and uploads debug logs if validation fails.

## Why these changes target error 2503

Error 2503 is a generic Windows Installer failure. When it happens for every generated MSI, the CI-generated package should be treated as suspect until the bundler toolchain and MSI database are proven healthy. Explicit WiX/NSIS installation removes the most likely hidden dependency failure. Forcing `COMSPEC` avoids shell selection surprises in downstream Windows tooling. Running `msiexec /a` validates that Windows Installer can open and process the MSI database before users download it. Uploading `msiexec-validate.log` gives maintainers the installer engine's exact failure diagnostics instead of relying on end-user screenshots.

## Testing instructions

1. Run `Build Application` manually with the smoke app and the Windows platform enabled through the existing matrix.
2. Confirm the Windows job logs show successful WiX, NSIS, Visual Studio, and Windows SDK verification.
3. Confirm `Validate Windows MSI database` passes and produces no error in `msiexec-validate.log`.
4. Download the MSI artifact from GitHub Actions.
5. On a clean Windows 10/11 VM, run:

   ```powershell
   msiexec /i .\smoke.msi /L*v .\smoke-install.log
   ```

6. Verify the app installs and launches. If error 2503 persists, inspect `smoke-install.log` for failing custom actions, file table paths, permission errors, and cabinet extraction failures.
7. If CI validation fails first, download the `debug-<slug>-windows` artifact and inspect `msiexec-validate.log` plus the preserved Pake output directory.
