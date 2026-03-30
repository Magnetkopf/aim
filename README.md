# aim

**aim** is a AppImage installer. Make you love your AppImage!

## Features

- **Fast & small:** GoLang♥️, and use `//go:embed` for frontend.
- **Natively integrated:** You can simply double-click any `.AppImage` in your file explorer to install them.
- **Immutable paths:** Applications will be isolated under `<AppName>/<SHA256>`.
- **Desktop friendly:** Automatically copies icons and `.desktop`!

## Build

1. **Build the Web UI**
   ```bash
   cd web
   pnpm install
   pnpm run build
   cd ..
   ```

2. **Build the Go CLI binary**
   ```bash
   go build -o aim ./cmd/aim
   ```

## Usage

**1. First time?**
Install binary, and register aim to be the default handler for all `.AppImage` files:
```bash
# Install
sudo mv aim /usr/local/bin/

# Register
aim --register
```

**2. General use**
After registration, simply **double-click** any `.AppImage` file in your system file browser, or trigger it via the command line:

```bash
aim /path/to/downloaded/software.AppImage
```

Your default web browser will instantly pop up the installation guide!