package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"templ-icons/generate"
)

func main() {
	prefix := flag.String("prefix", "lucide", "Iconify icon set prefix (e.g. lucide, ph, mdi)")
	outDir := flag.String("out", "./lucide", "Output directory for generated templ files")
	workers := flag.Int("workers", 20, "Number of parallel download workers")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}

	fmt.Printf("fetching icon list for prefix: %s\n", *prefix)

	icons, err := generate.FetchIconList(*prefix)
	if err != nil {
		log.Fatalf("failed to fetch icon list: %v", err)
	}

	fmt.Printf("found %d icons, downloading with %d workers...\n", len(icons), *workers)

	results := generate.DownloadAndGenerate(*prefix, icons, *outDir, *workers)

	fmt.Printf("\ndone: %d generated, %d failed\n", results.Success, results.Failed)
}
