// Command fetchfonts fetches the Chinese fonts the extension carries,
// from Google Fonts, in the slices it gives browsers (each a WOFF2 file
// with every weight, for a range of characters; the browser loads only
// the slices a page's text needs). It keeps the slices with Chinese
// characters, and only those characters, and writes them to
// extension/fonts/<variant>/, with extension/fonts/<variant>.js
// listing them under Miro's fallbacks' names for Chinese, Noto Sans JP
// and Noto Serif JP, for js/fix.js to add to the page. Run again only
// to update them:
//
//	go run ./cmd/fetchfonts
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gilramir/argparse/v2"
)

// source is one of Google's families, and the Miro fallback it stands in for.
type source struct {
	family  string // on Google Fonts
	weights string // its weight axis, as css2 asks for it
	alias   string // Miro's fallback, the name it takes on the page
	file    string // its slices' file names' start
}

// variants are the forms of Chinese the extension can show. Traditional
// would be Noto Sans TC and Noto Serif TC (Taiwan), or HK (Hong Kong).
var variants = map[string][]source{
	"sc": {
		{family: "Noto Sans SC", weights: "100..900", alias: "Noto Sans JP", file: "sans"},
		{family: "Noto Serif SC", weights: "200..900", alias: "Noto Serif JP", file: "serif"},
	},
}

// han is what the fonts take over from Miro's Japanese ones: Chinese
// characters, their radicals and strokes, CJK and full-width
// punctuation, bopomofo. Not kana (3040–30FF, 31F0–31FF, half-width
// FF61–FF9F) or Japan's squared words (3300–33FF): Japanese keeps those.
var han = mustRanges("U+2E80-2FDF, U+2FF0-2FFF, U+3000-303F, U+3100-312F, U+3190-31EF, " +
	"U+3200-321E, U+3220-3247, U+3280-32B0, U+3400-4DBF, U+4E00-9FFF, U+F900-FAFF, " +
	"U+FE10-FE1F, U+FE30-FE4F, U+FF00-FF60, U+FFE0-FFE6, U+20000-3134F")

// Google gives browsers WOFF2 slices; another client gets whole TTFs.
const browser = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"

type options struct {
	Variant string
	Out     string
}

func main() {
	opts := &options{Variant: "sc", Out: "extension/fonts"}
	ap := argparse.New(&argparse.Command{
		Name:        "fetchfonts",
		Description: "Fetch the extension's Chinese fonts from Google Fonts",
		Values:      opts,
	})
	ap.Add(&argparse.Argument{Switches: []string{"--variant"}, Choices: []string{"sc"}, Help: "Which Chinese (default sc, simplified)"})
	ap.Add(&argparse.Argument{Switches: []string{"--out"}, MetaVar: "DIR", Help: "Where the fonts go (default extension/fonts)"})
	if _, err := ap.ParseArgs(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := fetch(context.Background(), opts); err != nil {
		fmt.Fprintln(os.Stderr, "fetchfonts:", err)
		os.Exit(1)
	}
}

func fetch(ctx context.Context, opts *options) error {
	dir := filepath.Join(opts.Out, opts.Variant)
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var list []fontFace
	var bytes int64
	for _, src := range variants[opts.Variant] {
		q := "https://fonts.googleapis.com/css2?family=" + strings.ReplaceAll(src.family, " ", "+") + ":wght@" + src.weights
		sheet, err := get(ctx, q)
		if err != nil {
			return err
		}
		faces := parseFaces(string(sheet))
		if len(faces) == 0 {
			return fmt.Errorf("no faces in %s", q)
		}
		n := 0
		for _, f := range faces {
			keep := intersect(f.ranges, han)
			if len(keep) == 0 {
				continue
			}
			font, err := get(ctx, f.url)
			if err != nil {
				return err
			}
			name := fmt.Sprintf("%s-%02d.woff2", src.file, n)
			n++
			if err := os.WriteFile(filepath.Join(dir, name), font, 0o644); err != nil {
				return err
			}
			bytes += int64(len(font))
			list = append(list, fontFace{Family: src.alias, File: opts.Variant + "/" + name, UnicodeRange: formatRanges(keep)})
		}
		fmt.Printf("%s as %s: %d of %d slices\n", src.family, src.alias, n, len(faces))
		name := strings.ReplaceAll(src.family, " ", "")
		license, err := get(ctx, "https://raw.githubusercontent.com/google/fonts/main/ofl/"+strings.ToLower(name)+"/OFL.txt")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(opts.Out, "OFL-"+name+".txt"), license, 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("%.1f MB\n", float64(bytes)/1e6)
	js, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	head := "// Made by cmd/fetchfonts from Google Fonts: SIL Open Font License (OFL-*.txt).\n" +
		"// The faces js/fix.js adds to the page; files under fonts/.\n"
	return os.WriteFile(filepath.Join(opts.Out, opts.Variant+".js"), []byte(head+"var fontFaces = "+string(js)+";\n"), 0o644)
}

// fontFace is one face for js/fix.js to add: its file, under fonts/.
type fontFace struct {
	Family       string `json:"family"`
	File         string `json:"file"`
	UnicodeRange string `json:"unicodeRange"`
}

func get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browser)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// face is one @font-face rule of Google's.
type face struct {
	url    string
	ranges []span
}

var (
	ruleRE   = regexp.MustCompile(`@font-face\s*\{([^}]*)\}`)
	urlRE    = regexp.MustCompile(`src:\s*url\(([^)]+)\)`)
	rangeRE  = regexp.MustCompile(`unicode-range:\s*([^;]+);`)
)

func parseFaces(sheet string) []face {
	var faces []face
	for _, m := range ruleRE.FindAllStringSubmatch(sheet, -1) {
		u, r := urlRE.FindStringSubmatch(m[1]), rangeRE.FindStringSubmatch(m[1])
		if u == nil || r == nil {
			continue
		}
		spans, err := parseRanges(r[1])
		if err != nil {
			continue
		}
		faces = append(faces, face{url: u[1], ranges: spans})
	}
	return faces
}
