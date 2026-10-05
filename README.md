# Chinese Font Fix for Miro

中文字形修正（适用于 Miro）: a browser extension for Edge and Chrome
that makes Miro show Chinese in Chinese character forms (simplified,
SC), not Japanese ones. Not affiliated with Miro.

This is its source, published so anyone can see what it does before
installing it, or borrow the idea for a similar fix elsewhere.

## Install

- Chrome Web Store:
  https://chromewebstore.google.com/detail/gkpbpegjjmlibniojdfojholcmenmhkc
- Edge Add-ons:
  https://microsoftedge.microsoft.com/addons/detail/aoacfnjjmobhlhpkkhdffppfjfeapkjn

Or from this repo: in Edge, `edge://extensions`, turn on Developer
mode, Load unpacked, and choose the `extension` folder (Chrome: the
same at `chrome://extensions`). Reload any Miro board already open.

## The problem

Miro names no Chinese font. Its board text asks for `"Noto Sans",
Helvetica, OpenSans, Arial, sans-serif, …, "Noto Sans JP", "Noto Sans
KR"`, and the first of those with Chinese characters is Noto Sans JP.
So Chinese is drawn with a Japanese font, and characters whose forms
differ come out Japanese: 一直's 直, for one, and full-width punctuation
such as ；. (Serif text falls back the same way to Noto Serif JP.)

Reported to Miro:
https://community.miro.com/developer-platform-and-apis-57/chinese-characters-rendered-with-japanese-kanji-on-non-chinese-systems-28997

## How it works

It changes neither Miro's code nor its font lists. It gives the
browser a Chinese font under the name Miro already asks for.

`manifest.json` runs two scripts on `https://miro.com/app/*` pages at
`document_start`, before Miro's own code:

- `fonts/sc.js`: a list of 176 font faces, each a family name ("Noto
  Sans JP" or "Noto Serif JP"), a file, and a Unicode range;
- `js/fix.js`: about twenty lines that add each face to the page with
  the `FontFace` API, at weights 400 and 700:

  ```js
  const face = new FontFace("Noto Sans JP", "url(chrome-extension://…/fonts/sc/sans-17.woff2)",
                            { weight: "400", unicodeRange: "U+4E00-4F7F, …" });
  document.fonts.add(face);
  face.load();
  ```

Why this works:

- **It wins over Miro's font.** Of two faces with the same name that
  both have a character, the browser uses the one added later, and
  faces added by script come after every stylesheet's. (Putting the
  faces in a stylesheet lost to Miro's; so did declaring them for all
  weights, 100–900, against Miro's exact 400 and 700.)
- **Only Chinese changes.** `unicodeRange` limits each face to Chinese
  characters, CJK and full-width punctuation, and bopomofo. Latin,
  kana, Korean and the rest still come from Miro's fonts.
- **Board and page alike.** The board is drawn on a canvas and the
  search box is ordinary page text, but both look fonts up in the same
  `document.fonts`.
- **Ready before Miro draws.** `face.load()` loads every file at once
  rather than when a character first needs it.

The fonts are Noto Sans SC and Noto Serif SC from Google Fonts, the
WOFF2 slices Google serves to browsers (88 of each with Chinese in
them, 10.3 MB in all), under the SIL Open Font License
(`fonts/OFL-*.txt`). Only their ranges in `sc.js` are narrowed to
Chinese, so a slice that also has kana doesn't take kana from Miro.

`icons/icon.html` is the icon's source; `icons/render.sh` renders the
PNGs from it in headless Edge. Neither is in the published extension.

## Fetching the fonts again

`cmd/fetchfonts` (Go 1.27) made `extension/fonts/`, and running it
again remakes it:

    go run ./cmd/fetchfonts

It asks Google Fonts for Noto Sans SC and Noto Serif SC as a browser
would, keeps the slices with Chinese in them, writes them unchanged,
and writes `sc.js` with each slice's range cut down to Chinese, along
with the fonts' licenses. Run it and `git diff`: if Google hasn't
changed the fonts since, nothing differs, so you can check that the
font files are Google's own. `go test ./...` tests its range handling.

## What it doesn't do

No permissions, no background script, no network requests: the fonts
are files inside the extension, and nothing is sent anywhere. The font
files are `web_accessible_resources` only so that Miro's page may load
them.

## Limits

- It replaces Miro's Japanese fallback on the whole page, so Japanese
  on Miro gets Chinese forms for its kanji too. If you read Japanese
  on Miro, leave it off.
- It works only in Edge or Chrome on a computer, not on phones or in
  the Miro app.
- It depends on Miro's font lists as they are. If Miro changes its
  fallbacks, it quietly stops working.

## License

MIT (`LICENSE`), for the code. The fonts are under the SIL Open Font
License (`extension/fonts/OFL-*.txt`).
