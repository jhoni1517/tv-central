# Contributing / Como contribuir

Thanks for helping! / Valeu por ajudar!

## Reporting a bug / Relatar um problema
Open an issue with the **Diagnostics** text from the phone (Apps → Diagnóstico/Diagnostics): it has the version,
browser, network and the last lines of the log. Say which PC (CPU/GPU) and Windows version.

Abra uma issue com o texto de **Diagnóstico** do celular (Apps → Diagnóstico): ele traz a versão, o navegador, a
rede e as últimas linhas do registro. Diga qual PC (processador/placa de vídeo) e qual Windows.

## Code / Código
- Go 1.22+ for the program, plain HTML/CSS/JS (no build step) for the screens in `web/`.
- `go test ./...` must pass. Screens should stay light: only `transform`/`opacity` animations, no blur
  (they must run on an Intel HD 3000).
- User-facing text is written in Portuguese and translated in `web/i18n.js` (add the English and Spanish next to it).
- Keep the security rules: phone routes need pairing, TV routes only answer `127.0.0.1`, and nothing new opens a port.

## Ideas that would help / Ideias bem-vindas
- Arrow-navigation tweaks for specific streaming sites (`web/injetar/setas.js`).
- Translations to more languages (`web/i18n.js`).
- Art for more apps (`recursos/gerar_artes.py`) — original art only, no logos.
