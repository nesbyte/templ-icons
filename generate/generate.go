package generate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

const (
	iconifyBase = "https://api.iconify.design"
	batchSize   = 50
)

// collectionResponse is the shape of /collection?prefix=X
type collectionResponse struct {
	Prefix        string              `json:"prefix"`
	Total         int                 `json:"total"`
	Uncategorized []string            `json:"uncategorized"`
	Categories    map[string][]string `json:"categories"`
}

// iconsResponse is the shape of /{prefix}.json?icons=a,b,c
type iconsResponse struct {
	Icons map[string]struct {
		Body string `json:"body"`
	} `json:"icons"`
}

// Results holds the final counts
type Results struct {
	Success int
	Failed  int
}

// templFile is the template for a single generated .templ file
const templFile = `package lucide

templ {{ .ComponentName }}(class string) {
	<svg
		class={ class }
		xmlns="http://www.w3.org/2000/svg"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="2"
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden="true"
	>
		@templ.Raw(` + "`{{ .SVGBody }}`" + `)
	</svg>
}
`

type iconData struct {
	ComponentName string
	SVGBody       string
}

// FetchIconList calls the Iconify collection API and returns all icon names
func FetchIconList(prefix string) ([]string, error) {
	url := fmt.Sprintf("%s/collection?prefix=%s", iconifyBase, prefix)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	var col collectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&col); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// collect all icons whether categorised or not
	seen := make(map[string]bool)
	var all []string

	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			all = append(all, name)
		}
	}

	for _, name := range col.Uncategorized {
		add(name)
	}
	for _, names := range col.Categories {
		for _, name := range names {
			add(name)
		}
	}

	return all, nil
}

// DownloadAndGenerate fetches icons in batches and writes a .templ file for each
func DownloadAndGenerate(prefix string, icons []string, outDir string, _ int) Results {
	tmpl := template.Must(template.New("icon").Parse(templFile))

	var results Results

	// split into batches of batchSize
	for i := 0; i < len(icons); i += batchSize {
		end := i + batchSize
		if end > len(icons) {
			end = len(icons)
		}
		batch := icons[i:end]

		bodies, err := fetchBatch(prefix, batch)
		if err != nil {
			fmt.Printf("  ✗ batch %d-%d: %v\n", i, end, err)
			results.Failed += len(batch)
			continue
		}

		for _, name := range batch {
			body, ok := bodies[name]
			if !ok {
				fmt.Printf("  ✗ %s: not in response\n", name)
				results.Failed++
				continue
			}

			if err := writeTemplFile(name, body, outDir, tmpl); err != nil {
				fmt.Printf("  ✗ %s: %v\n", name, err)
				results.Failed++
			} else {
				fmt.Printf("  ✓ %s\n", name)
				results.Success++
			}
		}
	}

	return results
}

// fetchBatch hits /{prefix}.json?icons=a,b,c and returns name -> SVG body
func fetchBatch(prefix string, names []string) (map[string]string, error) {
	url := fmt.Sprintf("%s/%s.json?icons=%s", iconifyBase, prefix, strings.Join(names, ","))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var result iconsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	bodies := make(map[string]string, len(result.Icons))
	for name, icon := range result.Icons {
		bodies[name] = icon.Body
	}

	return bodies, nil
}

// writeTemplFile generates a single .templ file for an icon
func writeTemplFile(name, body, outDir string, tmpl *template.Template) error {
	fileName := strings.ReplaceAll(name, "-", "_") + ".templ"
	outPath := filepath.Join(outDir, fileName)

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, iconData{
		ComponentName: toComponentName(name),
		SVGBody:       body,
	})
}

// toComponentName converts kebab-case to PascalCase
// e.g. arrow-right -> ArrowRight, home -> Home
func toComponentName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, "")
}
