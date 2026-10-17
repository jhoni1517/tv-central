// Navegação por setas para sites feitos para mouse (Netflix, Disney+, HBO Max, Apple TV...).
// Setas: anda entre botões e capas. Enter: abre. Esc: volta.
// Com vídeo tocando em tela grande: Enter pausa/continua e as setas ficam com o player.
(() => {
  if (window.top !== window || window.__tvCentralSetas) return;
  window.__tvCentralSetas = true;

  const SELETOR = [
    'a[href]', 'button', 'input:not([type=hidden])', 'select', 'textarea', 'summary',
    '[role=button]', '[role=link]', '[role=tab]', '[role=menuitem]', '[role=option]',
    '[role=checkbox]', '[role=switch]', '[tabindex]:not([tabindex="-1"])', '[onclick]',
  ].join(',');
  const CLASSE = '__tv-foco';
  let atual = null;

  const css = document.createElement('style');
  css.textContent = `
    .${CLASSE}{outline:4px solid #fff!important;outline-offset:2px!important;
      box-shadow:0 0 0 9px rgba(64,156,255,.6),0 10px 40px rgba(0,0,0,.6)!important;
      border-radius:8px;z-index:2147483000}`;
  const colocarCss = () => { if (!css.isConnected) (document.head || document.documentElement).appendChild(css); };
  colocarCss();
  document.addEventListener('DOMContentLoaded', colocarCss);

  const ehCampo = el => el && (el.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(el.tagName));

  function videoGrande() {
    for (const v of document.querySelectorAll('video')) {
      const r = v.getBoundingClientRect();
      if (r.width * r.height > innerWidth * innerHeight * 0.85 && v.readyState > 1) return v;
    }
    return null;
  }

  // Rápido: só lê o tamanho (getBoundingClientRect). As checagens caras ficam para os poucos finalistas.
  function perto(r, margem) {
    return r.width >= 6 && r.height >= 6 && r.bottom > -margem && r.top < innerHeight + margem &&
      r.right > -margem && r.left < innerWidth + margem;
  }

  // Caro: estilo e se está coberto por outra coisa. Roda só nos melhores colocados.
  function utilizavel(el, r) {
    if (el.disabled || el.getAttribute('aria-hidden') === 'true') return false;
    const s = getComputedStyle(el);
    if (s.visibility === 'hidden' || s.display === 'none' || s.pointerEvents === 'none' || +s.opacity === 0) return false;
    if (r.top >= 0 && r.left >= 0 && r.bottom <= innerHeight && r.right <= innerWidth) {
      const topo = document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
      if (topo && topo !== el && !el.contains(topo) && !topo.contains(el) && topo.closest(SELETOR) !== el) return false;
    }
    return true;
  }

  function centro(r) { return { x: r.left + r.width / 2, y: r.top + r.height / 2 }; }

  // Distância lateral: zero se os dois se sobrepõem nesse eixo.
  function sobra(a1, a2, b1, b2, pa, pb) {
    if (a1 < b2 && a2 > b1) return 0;
    return Math.abs(pa - pb);
  }

  // Se o escolhido está dentro de outro clicável quase do mesmo tamanho, usa o de fora.
  function externo(el) {
    const r = el.getBoundingClientRect();
    let p = el.parentElement && el.parentElement.closest(SELETOR);
    while (p) {
      const q = p.getBoundingClientRect();
      if (Math.abs(q.width - r.width) > 24 || Math.abs(q.height - r.height) > 24) break;
      el = p;
      p = p.parentElement && p.parentElement.closest(SELETOR);
    }
    return el;
  }

  function melhor(dir) {
    const de = atual.getBoundingClientRect(), c = centro(de), horiz = dir === 'direita' || dir === 'esquerda';
    const lista = [];
    for (const el of document.querySelectorAll(SELETOR)) {
      if (el === atual) continue;
      const r = el.getBoundingClientRect();
      if (!perto(r, innerHeight)) continue;
      const p = centro(r);
      let frente, lado;
      if (dir === 'direita') { if (p.x <= c.x + 4) continue; frente = r.left - de.right; lado = sobra(r.top, r.bottom, de.top, de.bottom, p.y, c.y); }
      else if (dir === 'esquerda') { if (p.x >= c.x - 4) continue; frente = de.left - r.right; lado = sobra(r.top, r.bottom, de.top, de.bottom, p.y, c.y); }
      else if (dir === 'baixo') { if (p.y <= c.y + 4) continue; frente = r.top - de.bottom; lado = sobra(r.left, r.right, de.left, de.right, p.x, c.x); }
      else { if (p.y >= c.y - 4) continue; frente = de.top - r.bottom; lado = sobra(r.left, r.right, de.left, de.right, p.x, c.x); }
      lista.push({ el, r, v: Math.max(frente, -20) + lado * (horiz ? 4 : 1.5) });
    }
    lista.sort((a, b) => a.v - b.v);
    for (const { el, r } of lista.slice(0, 12)) {
      if (el.contains(atual) || atual.contains(el)) continue;
      if (utilizavel(el, r)) return externo(el);
    }
    return null;
  }

  function marcar(el) {
    if (atual) atual.classList.remove(CLASSE);
    atual = el;
    if (!el) return;
    el.classList.add(CLASSE);
    try { el.focus({ preventScroll: true }); } catch (_) {}
    const r = el.getBoundingClientRect(), opts = { bubbles: true, clientX: r.left + r.width / 2, clientY: r.top + r.height / 2 };
    for (const t of ['pointerover', 'pointerenter', 'mouseover', 'mouseenter', 'mousemove']) {
      el.dispatchEvent(new (t.startsWith('pointer') ? PointerEvent : MouseEvent)(t, opts));
    }
    if (r.top < 60 || r.bottom > innerHeight - 60 || r.left < 0 || r.right > innerWidth) {
      el.scrollIntoView({ block: 'center', inline: 'nearest' });
    }
  }

  function primeiro() {
    const lista = [];
    for (const el of document.querySelectorAll(SELETOR)) {
      const r = el.getBoundingClientRect();
      if (r.width >= 6 && r.height >= 6 && r.top >= 0 && r.bottom <= innerHeight && r.left >= 0 && r.right <= innerWidth) lista.push({ el, r });
    }
    lista.sort((a, b) => (a.r.top - b.r.top) || (a.r.left - b.r.left));
    // Pula o menu do topo quando houver conteúdo
    const ordem = lista.filter(({ r }) => r.top > innerHeight * 0.15).concat(lista);
    for (const { el, r } of ordem.slice(0, 30)) if (utilizavel(el, r)) return externo(el);
    return null;
  }

  function rolarPara(dir) {
    const passo = { cima: [0, -0.6], baixo: [0, 0.6], esquerda: [-0.6, 0], direita: [0.6, 0] }[dir];
    let alvo = atual && atual.parentElement;
    while (alvo && alvo !== document.body) {
      const s = getComputedStyle(alvo);
      const rolaX = /(auto|scroll)/.test(s.overflowX) && alvo.scrollWidth > alvo.clientWidth;
      const rolaY = /(auto|scroll)/.test(s.overflowY) && alvo.scrollHeight > alvo.clientHeight;
      if ((passo[0] && rolaX) || (passo[1] && rolaY)) break;
      alvo = alvo.parentElement;
    }
    const el = alvo && alvo !== document.body ? alvo : window;
    const w = el === window ? innerWidth : el.clientWidth, h = el === window ? innerHeight : el.clientHeight;
    el.scrollBy({ left: passo[0] * w, top: passo[1] * h, behavior: 'smooth' });
  }

  // Só o clique: disparar também pointerdown/mousedown fazia alguns sites agirem duas vezes
  function clicar(el) { el.click(); }

  const DIRS = { ArrowUp: 'cima', ArrowDown: 'baixo', ArrowLeft: 'esquerda', ArrowRight: 'direita' };

  window.addEventListener('keydown', e => {
    if (e.ctrlKey || e.altKey || e.metaKey) return;
    const dir = DIRS[e.key];
    const foco = document.activeElement;

    // Vídeo em tela grande: o player cuida das setas; Enter pausa; Esc volta.
    const v = videoGrande();
    if (v) {
      if (e.key === 'Enter') { v.paused ? v.play() : v.pause(); parar(e); return; }
      if (document.fullscreenElement) return; // Esc sai da tela cheia normalmente
      if (e.key === 'Escape') { history.back(); parar(e); }
      return;
    }

    // Digitando num campo: esquerda/direita e Enter ficam com o campo; cima/baixo saem dele.
    if (ehCampo(foco)) {
      if (e.key === 'Escape') { foco.blur(); parar(e); return; }
      if (!dir || dir === 'esquerda' || dir === 'direita') return;
    }

    if (dir) {
      parar(e);
      if (!atual || !atual.isConnected || !perto(atual.getBoundingClientRect(), innerHeight)) { marcar(primeiro()); return; }
      const prox = melhor(dir);
      if (prox) marcar(prox);
      else { rolarPara(dir); setTimeout(() => { const p = melhor(dir); if (p) marcar(p); }, 350); }
      return;
    }
    if (e.key === 'Enter' && atual && atual.isConnected) {
      parar(e);
      if (ehCampo(atual)) atual.focus(); else clicar(atual);
      return;
    }
    if (e.key === 'Escape') {
      parar(e);
      if (history.length > 1) history.back();
    }
  }, true);

  function parar(e) { e.preventDefault(); e.stopImmediatePropagation(); }

  // Se a página trocar o conteúdo (navegação sem recarregar), começa do primeiro item de novo.
  let ultimaUrl = location.href;
  setInterval(() => {
    if (location.href !== ultimaUrl) { ultimaUrl = location.href; setTimeout(() => marcar(primeiro()), 800); }
  }, 500);
})();
