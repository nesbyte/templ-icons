# templ-icons

Lucide icons as typed Go Templ components. No runtime dependencies, no CDN, full Tailwind colour control via `currentColor`.

## Installation

```bash
go get github.com/nesbyte/templ-icons
```

## Usage

```templ
import "github.com/nesbyte/templ-icons/lucide"

@lucide.Home("w-6 h-6 text-red-600")
@lucide.ArrowRight("w-4 h-4 text-white")
@lucide.Settings("w-5 h-5 text-gray-400 hover:text-white transition-colors")
```

Colour is controlled via Tailwind's `text-*` utilities since the SVG uses `currentColor` for stroke.

## Regenerating icons

Icons are pre-generated and committed. To regenerate (e.g. after a new Lucide release):

```bash
go run main.go --prefix lucide --out ./lucide --workers 20
```

Or via `go generate`:

```bash
go generate ./lucide/...
```

## Adding other icon sets

```bash
go run main.go --prefix ph --out ./phosphor --workers 20
go run main.go --prefix mdi --out ./material --workers 20
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--prefix` | `lucide` | Iconify icon set prefix |
| `--out` | `./lucide` | Output directory |
| `--workers` | `20` | Parallel download workers |