# skli

`skli` is a powerful CLI/TUI tool for managing and installing skills from Git repositories.

## 🚀 Installation

You can install `skli` on Mac, Linux, or Windows without needing Go installed.

### Mac / Linux

Run the following command in your terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/Bafoj/Skli/main/scripts/install.sh | bash
```

> [!NOTE]
> This script will attempt to install `skli` to `/usr/local/bin`. You may be prompted for your password.

### Windows

Run the following command in PowerShell (as Administrator for best results):

```powershell
powershell -c "irm https://raw.githubusercontent.com/Bafoj/Skli/main/scripts/install.ps1 | iex"
```

---

## 🛠 Usage

### 1. Add skills (interactive)
Use `add` to install skills. Without args, it opens the interactive TUI.

```bash
skli add
```

### 2. Add from a specific URL
Pass a Git repository URL directly:

```bash
skli add https://github.com/user/my-skills-repo
```

### 3. Remove skills
Delete one skill by name:

```bash
skli rm my-skill
```

Or run without args for a checkbox TUI:

```bash
skli rm
```

### 4. Synchronize installed skills
To update all your installed skills from their source repositories:

```bash
skli sync
```

### 5. Upload local skills
Upload directly:

```bash
skli upload https://github.com/user/repo.git ./skills/my-skill
```

Or run without args for the 2-step TUI flow:

```bash
skli upload
```

### 6. Configuration
To configure global settings and default remotes:

```bash
skli config
```

### 7. Help

```bash
skli --help
```

---

## 📂 Project Structure

- `cmd/skli`: Main application entry point.
- `internal/tui`: Terminal User Interface implementation.
- `internal/gitrepo`: Git repository handling and skill detection.
- `internal/config`: Global configuration management.
- `scripts`: Installation scripts.

---

## 📄 License
Check the repository for license details.
