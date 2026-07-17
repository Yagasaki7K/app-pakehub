package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type App struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Slug string `json:"-"`
}

const (
	appsFile     = "apps.json"
	workflowsDir = ".github/workflows"
	projectsFile = "PROJECTS.md"
	readmeFile   = "README.md"
	readmeStart  = "<!-- PAKE:APPS:START -->"
	readmeEnd    = "<!-- PAKE:APPS:END -->"
)

var slugCleaner = regexp.MustCompile(`[^a-z0-9-]+`)

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	return slugCleaner.ReplaceAllString(s, "")
}

func downloadURL(repo, branch, slug, ext string) string {
	return fmt.Sprintf("https://github.com/%s/raw/refs/heads/%s/projects/%s.%s", repo, branch, slug, ext)
}

func renderWorkflow(app App) string {
	return strings.NewReplacer(
		"__APP_NAME__", app.Name,
		"__APP_SLUG__", app.Slug,
		"__APP_URL__", app.URL,
	).Replace(workflowTemplate)
}

func renderList(repo, branch string, apps []App) string {
	var b strings.Builder
	b.WriteString("# Projects to Download\n\n## Windows (.msi)\n")
	for _, a := range apps {
		fmt.Fprintf(&b, "- [%s (Windows)](%s)\n", a.Name, downloadURL(repo, branch, a.Slug, "msi"))
	}
	b.WriteString("\n## Linux (.deb)\n")
	for _, a := range apps {
		fmt.Fprintf(&b, "- [%s (Linux)](%s)\n", a.Name, downloadURL(repo, branch, a.Slug, "deb"))
	}
	return b.String()
}

func updateReadme(repo, branch string, apps []App) error {
	block := readmeStart + "\n\n" + renderList(repo, branch, apps) + "\n" + readmeEnd

	data, err := os.ReadFile(readmeFile)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(readmeFile, []byte("# README\n\n"+block+"\n"), 0644)
		}
		return err
	}
	content := string(data)
	start := strings.Index(content, readmeStart)
	end := strings.Index(content, readmeEnd)
	if start >= 0 && end > start {
		content = content[:start] + block + content[end+len(readmeEnd):]
	} else {
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + block + "\n"
	}
	return os.WriteFile(readmeFile, []byte(content), 0644)
}

func main() {
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		repo = "Yagasaki7K/app-pakebuilder"
	}
	branch := os.Getenv("GITHUB_REF_NAME")
	if branch == "" {
		branch = "main"
	}

	data, err := os.ReadFile(appsFile)
	if err != nil {
		fmt.Println("erro lendo", appsFile, ":", err)
		os.Exit(1)
	}
	var apps []App
	if err := json.Unmarshal(data, &apps); err != nil {
		fmt.Println("JSON inválido:", err)
		os.Exit(1)
	}

	seen := map[string]bool{}
	for i := range apps {
		if apps[i].Name == "" || apps[i].URL == "" {
			fmt.Println("todo app precisa de name e url")
			os.Exit(1)
		}
		apps[i].Slug = slugify(apps[i].Name)
		if seen[apps[i].Slug] {
			fmt.Println("slug duplicado:", apps[i].Slug)
			os.Exit(1)
		}
		seen[apps[i].Slug] = true
	}
	sort.Slice(apps, func(i, j int) bool { return apps[i].Slug < apps[j].Slug })

	// apaga apenas os workflows GERADOS (prefixo build-); orchestrator e outros ficam intactos
	old, _ := filepath.Glob(filepath.Join(workflowsDir, "build-*.yaml"))
	yml, _ := filepath.Glob(filepath.Join(workflowsDir, "build-*.yml"))
	for _, f := range append(old, yml...) {
		os.Remove(f)
		fmt.Println("removido:", f)
	}

	for _, app := range apps {
		path := filepath.Join(workflowsDir, "build-"+app.Slug+".yaml")
		if err := os.WriteFile(path, []byte(renderWorkflow(app)), 0644); err != nil {
			fmt.Println("erro escrevendo", path, ":", err)
			os.Exit(1)
		}
		fmt.Println("gerado:", path)
	}

	if err := os.WriteFile(projectsFile, []byte(renderList(repo, branch, apps)), 0644); err != nil {
		fmt.Println("erro escrevendo PROJECTS.md:", err)
		os.Exit(1)
	}
	if err := updateReadme(repo, branch, apps); err != nil {
		fmt.Println("erro atualizando README.md:", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %d apps processados\n", len(apps))
}

const workflowTemplate = `name: Build __APP_NAME__ Pake Installers

on:
  workflow_dispatch:

permissions:
  contents: read

concurrency:
  group: pake-__APP_SLUG__-${{ github.ref }}
  cancel-in-progress: false

jobs:
  build-windows:
    name: Build Windows MSI
    runs-on: windows-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '22'
      - name: Prepare artifact directory
        shell: bash
        run: mkdir -p "artifacts/windows"
      - name: Install Pake CLI with npm
        shell: bash
        run: npm install --global pake-cli
      - name: Verify Pake CLI
        shell: bash
        run: pake --version
      - name: Build Windows installer
        shell: bash
        run: pake "__APP_URL__" --name "__APP_NAME__"
      - name: Discover and collect Windows installer
        shell: bash
        run: scripts/ci/collect-pake-artifacts.sh --kind windows --search-root . --output-dir "artifacts/windows"
      - name: Normalize installer filename to lowercase
        shell: bash
        run: |
          set -euo pipefail
          find artifacts -type f \( -iname '*.msi' -o -iname '*.deb' \) | while read -r f; do
            dir=$(dirname "$f"); base=$(basename "$f")
            lower=$(printf '%s' "$base" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
            if [ "$base" != "$lower" ]; then
              mv "$f" "$dir/$lower"
              echo "renamed: $base → $lower"
            fi
          done
      - name: Upload Windows installer
        uses: actions/upload-artifact@v4
        with:
          name: __APP_SLUG__-windows-msi
          path: artifacts/windows/*.msi
          if-no-files-found: error

  build-linux:
    name: Build Linux DEB
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '22'
      - name: Prepare artifact directory
        run: mkdir -p "artifacts/linux"
      - name: Install Linux dependencies
        run: |
          set -euo pipefail
          sudo apt-get update
          sudo apt-get install -y build-essential libssl-dev libgtk-3-dev libwebkit2gtk-4.1-dev libayatana-appindicator3-dev librsvg2-dev libpango1.0-dev libcairo2-dev
      - name: Install Pake CLI with npm
        run: npm install --global pake-cli
      - name: Verify Pake CLI
        run: pake --version
      - name: Build Linux installer
        run: pake "__APP_URL__" --name "__APP_NAME__" --targets deb
      - name: Discover and collect Linux installer
        run: scripts/ci/collect-pake-artifacts.sh --kind linux --search-root . --output-dir "artifacts/linux"
      - name: Normalize installer filename to lowercase
        shell: bash
        run: |
          set -euo pipefail
          find artifacts -type f \( -iname '*.msi' -o -iname '*.deb' \) | while read -r f; do
            dir=$(dirname "$f"); base=$(basename "$f")
            lower=$(printf '%s' "$base" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
            if [ "$base" != "$lower" ]; then
              mv "$f" "$dir/$lower"
              echo "renamed: $base → $lower"
            fi
          done
      - name: Upload Linux installer
        uses: actions/upload-artifact@v4
        with:
          name: __APP_SLUG__-linux-deb
          path: artifacts/linux/*.deb
          if-no-files-found: error
`