# skli

`skli` is a powerful CLI/TUI tool for managing and installing skills from Git repositories.

## 🚀 Installation

You can install `skli` on Mac, Linux, or Windows without needing Go installed.

### Mac / Linux

Run the following command in your terminal:

```bash
curl -fsSL https://bitbucket.org/cuatroochenta/skli/raw/main/scripts/install.sh | bash
```

> [!NOTE]
> This script will attempt to install `skli` to `/usr/local/bin`. You may be prompted for your password.

### Windows

Run the following command in PowerShell (as Administrator for best results):

```powershell
powershell -c "irm https://bitbucket.org/cuatroochenta/skli/raw/main/scripts/install.ps1 | iex"
```

---

## 🛠 Usage

### 1. Basic Interactive Mode
Simply run `skli` to open the interactive TUI. From here you can explore remotes and install skills.

```bash
skli
```

### 2. Install from a specific URL
You can pass a Git repository URL directly to start exploring its skills:

```bash
skli https://github.com/user/my-skills-repo
```

### 3. Synchronize installed skills
To update all your installed skills from their source repositories:

```bash
skli sync
```

### 4. Manage installed skills
To list and delete your currently installed skills:

```bash
skli manage
```

### 5. Configuration
To configure global settings and default remotes:

```bash
skli config
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
