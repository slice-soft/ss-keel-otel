<img src="https://cdn.slicesoft.dev/boat.svg" width="400" />

# ss-keel-addon-template
Base template for building Keel addons with a ready `keel-addon.json` contract, CI/CD scaffolding, and release automation.

[![Template Repo](https://img.shields.io/badge/Template-ss--keel--addon--template-0A7F5A)](https://github.com/slice-soft/ss-keel-addon-template)
[![Use this template](https://img.shields.io/badge/GitHub-Use%20this%20template-1f6feb)](https://github.com/slice-soft/ss-keel-addon-template/generate)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)
![Made in Colombia](https://img.shields.io/badge/Made%20in-Colombia-FCD116?labelColor=003893)


This repository is a **base template** for building **Keel Framework** addons.
It helps developers and companies quickly create functional addons with CI/CD, testing, and automated releases.
It also serves as a reference for the `keel-addon.json` contract used by the CLI to download and integrate addons into Keel projects.

---

## 🚀 Template structure

```
ss-keel-addon-template/
├── .github/workflows/    # CI and release workflows (commented by default)
│   ├── ci.yml
│   └── release.yml
├── .gitignore
├── .release-please-manifest.json
├── .release-please-config.json
├── CONTRIBUTING.md       # Contribution guide
├── keel-addon.json       # Addon contract for the CLI
├── LICENSE
├── README.md
└── go.mod
```

---

## 🛠️ Create a new addon

Recommended option (GitHub Template):

1. Open this repository on GitHub.
2. Click **Use this template**.
3. Create your new repository from this template.
4. Clone your new repository locally.

Alternative option (manual clone):

```bash
# Clone the template into a new project
git clone https://github.com/slice-soft/ss-keel-addon-template.git my-addon
cd my-addon

# Delete the existing git history
rm -rf .git

# Initialize a new git repository
git init

# Update the Go module path
go mod edit -module github.com/my-company/my-addon
go mod tidy
```

> `my-addon` is now ready to be developed and registered in Keel.

Edit `keel-addon.json` with your addon's real values (`name`, `repo`, `version`, `steps`, etc.).

---

## ⚡️ Keel integration

* Place your addon logic in `internal/addon`.
* Define metadata in `keel-addon.json`. This file is the contract the Keel CLI validates to install and integrate the addon.

```json
{
  "name": "my-addon",
  "version": "0.1.0",
  "description": "Short addon description",
  "repo": "github.com/your-user/your-repo",
  "steps": [
    {
      "type": "go_get",
      "package": "github.com/your-user/your-repo"
    },
    {
      "type": "property",
      "key": "addon.example-key",
      "example": "${ADDON_EXAMPLE_KEY:}",
      "description": "Example runtime setting exposed through application.properties."
    },
    {
      "type": "env",
      "key": "ADDON_EXAMPLE_KEY",
      "example": "",
      "description": "Example environment variable consumed by application.properties."
    },
    {
      "type": "create_provider_file",
      "filename": "cmd/setup_my_addon.go",
      "guard": "func setupMyAddon(",
      "content": "package main\\n\\nimport (\\n  \\\"github.com/slice-soft/ss-keel-core/config\\\"\\n  \\\"github.com/slice-soft/ss-keel-core/core\\\"\\n)\\n\\ntype myAddonConfig struct {\\n  ExampleKey string `keel:\\\"addon.example-key\\\"`\\n}\\n\\nfunc setupMyAddon(app *core.App) {\\n  _ = app\\n  _ = config.MustLoadConfig[myAddonConfig]()\\n}\\n"
    },
    {
      "type": "main_code",
      "anchor": "before_modules",
      "guard": "setupMyAddon(",
      "code": "setupMyAddon()"
    }
  ]
}
```

* The Keel CLI uses this file to:
  * Resolve the module to download (`repo`).
  * Validate that the addon matches the expected format.
  * Execute `steps` to integrate changes automatically across `application.properties`, `.env`, and `cmd/main.go`.
* Prefer `config.MustLoadConfig[T]` in generated provider files so addon runtime config comes from `application.properties` with env overrides, instead of reading individual keys manually.

---

## 🧭 `keel add` flow in the ecosystem

The CLI supports two installation paths:

1. **Official or verified addons**

```bash
keel add gorm
```

* `gorm` is interpreted as an alias.
* The CLI checks the `ss-keel-addons` alias repository.
* If the alias exists, it gets the addon URL, downloads it, and validates its `keel-addon.json`.
* Then it executes the defined `steps` to integrate it automatically into the project.

2. **Unofficial addons or addons not verified by SliceSoft/community**

```bash
keel add github.com/user/repo
```

* The CLI uses the provided repository directly.
* It downloads the module and validates its `keel-addon.json`.
* If validation passes, it applies the automatic integration steps the same way as official addons.

---

## 📚 Alias library: `ss-keel-addons`

`ss-keel-addons` works as an alias catalog/library for addons.

* Stores the relationship `alias -> repository URL`.
* Lets the CLI verify whether an alias exists before installing.
* Centralizes official or community-verified addons.
* Acts as the entry point for pre-validation before the automatic download and integration process.

---

## 🤚 CI/CD and releases

This repository is a template, so workflows are intentionally shipped **commented out** to avoid accidental executions after cloning:

* `.github/workflows/ci.yml`
* `.github/workflows/release.yml`

To enable CI/CD in your new addon repository:

1. Uncomment both workflow files.
2. Push to GitHub to validate that Actions run correctly.

---

## 💡 Recommendations

* Keep your addons independent and modular.
* Use Keel events and guards to extend functionality without touching the core.
* Document each addon in its own project README.

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for setup and repository-specific rules.
The base workflow, commit conventions, and community standards live in [ss-community](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md).

## Community

| Document | |
|---|---|
| [CONTRIBUTING.md](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md) | Workflow, commit conventions, and PR guidelines |
| [GOVERNANCE.md](https://github.com/slice-soft/ss-community/blob/main/GOVERNANCE.md) | Decision-making, roles, and release process |
| [CODE_OF_CONDUCT.md](https://github.com/slice-soft/ss-community/blob/main/CODE_OF_CONDUCT.md) | Community standards |
| [VERSIONING.md](https://github.com/slice-soft/ss-community/blob/main/VERSIONING.md) | SemVer policy and breaking changes |
| [SECURITY.md](https://github.com/slice-soft/ss-community/blob/main/SECURITY.md) | How to report vulnerabilities |
| [MAINTAINERS.md](https://github.com/slice-soft/ss-community/blob/main/MAINTAINERS.md) | Active maintainers |

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- Website: [keel-go.dev](https://keel-go.dev)
- GitHub: [github.com/slice-soft/ss-keel-addon-template](https://github.com/slice-soft/ss-keel-addon-template)
- Documentation: [docs.keel-go.dev](https://docs.keel-go.dev)

---

Made by [SliceSoft](https://slicesoft.dev) — Colombia 💙
