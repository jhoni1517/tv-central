// Bloqueio de anúncios do TV Central (só nas janelas abertas pelo programa).
// YouTube: tira os anúncios dos dados do vídeo antes do player ler (a mesma técnica dos
// bloqueadores conhecidos) e, se algum passar, pula ou acelera até o fim.
(() => {
  if (window.__tvCentralAnuncios) return;
  window.__tvCentralAnuncios = true;
  const host = location.hostname;
  if (!/(^|\.)youtube(kids)?\.com$/.test(host)) return;

  const CHAVES = ['adPlacements', 'adSlots', 'playerAds', 'adBreakHeartbeatParams'];
  function limpar(obj, n = 0) {
    if (!obj || typeof obj !== 'object' || n > 6) return obj;
    for (const k of CHAVES) if (k in obj) delete obj[k];
    for (const k of ['playerResponse', 'response', 'playerConfig']) if (obj[k]) limpar(obj[k], n + 1);
    return obj;
  }
  const parseOriginal = JSON.parse;
  JSON.parse = function () { return limpar(parseOriginal.apply(this, arguments)); };
  const jsonOriginal = Response.prototype.json;
  Response.prototype.json = function () { return jsonOriginal.apply(this, arguments).then(limpar); };
  try {
    let atual;
    Object.defineProperty(window, 'ytInitialPlayerResponse', { configurable: true, get: () => atual, set: v => { atual = limpar(v); } });
  } catch (_) {}

  // Se algum anúncio aparecer mesmo assim: aperta "Pular" ou vai direto ao fim dele
  const PULAR = '.ytp-skip-ad-button, .ytp-ad-skip-button, .ytp-ad-skip-button-modern, .ytp-ad-skip-button-container button';
  setInterval(() => {
    const emAnuncio = document.querySelector('.ad-showing, .ad-interrupting, ytlr-ad-ui, .ytlr-ad, [class*="ad-showing"]');
    if (!emAnuncio) return;
    const b = document.querySelector(PULAR);
    if (b) { b.click(); return; }
    const v = document.querySelector('video');
    if (v && isFinite(v.duration) && v.duration > 0 && v.currentTime < v.duration - 0.3) {
      v.muted = true;
      v.currentTime = v.duration - 0.1;
    }
  }, 300);
})();
