// Miro names no Chinese font: the browser draws Chinese with Miro's
// fallbacks, Noto Sans JP and Noto Serif JP, in Japanese forms (直).
// This adds Chinese fonts to the page under those names, for the
// characters in each face's unicode-range only: kana and all else stay
// Miro's. Faces added by script come after every stylesheet's, Miro's
// included, and of two faces with the same name the later wins (in a
// stylesheet of the extension's, they lost). Each file has every
// weight; it is added at 400 and at 700, the weights of Miro's own
// faces: added once for all weights (100 900), it lost to them too.
// Each is loaded at once, not when first needed, to be ready before
// Miro first draws its text on its canvas.
for (const f of fontFaces) {
  for (const weight of ["400", "700"]) {
    const face = new FontFace(f.family, `url(${chrome.runtime.getURL("fonts/" + f.file)}) format("woff2")`, {
      weight,
      unicodeRange: f.unicodeRange,
    });
    document.fonts.add(face);
    face.load().catch((err) => console.warn("Miro Chinese Font Fix:", f.file, err));
  }
}
