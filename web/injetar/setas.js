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
      border-radius:8px;transition:outline-offset .12s,box-shadow .12s;z-index:2147483000}
    html{scroll-behavior:smooth}`;
  const colocarCss = () => { if (!css.isConnected) (document.head || document.documentElement).appendChild(css); };
  colocarCss();
  document.addEventListener('DOMContentLoaded', colocarCss);

  const ehCampo = el => el && (el.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(el.tagName));

  function videoGrande() {
    for (const v of document.querySelectorAll('video')) {
      const r = v.getBoundingClientRect();
      if (r.width * r.height > innerWidth * innerHeight * 0.5 && v.readyState > 1) return v;
    }
    return null;
  }

  // Retângulo útil do elemento, ou null se não dá para usar (invisível, coberto, minúsculo).
  function retangulo(el, margem) {
    const r = el.getBoundingClientRect();
    if (r.width < 6 || r.height < 6) return null;
    if (r.bottom < -margem || r.top > innerHeight + margem || r.right < -margem || r.left > innerWidth + margem) return null;
    const s = getComputedStyle(el);
    if (s.visibility === 'hidden' || s.display === 'none' || s.pointerEvents === 'none' || +s.opacity === 0) return null;
    if (el.disabled || el.getAttribute('aria-hidden') === 'true') return null;
    const dentro = r.top >= 0 && r.left >= 0 && r.bottom <= innerHeight && r.right <= innerWidth;
    if (dentro) {
      const topo = document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
      if (topo && topo !== el && !el.contains(topo) && !topo.contains(el) && topo.closest(SELETOR) !== el) return null;
    }
    return r;
  }

  function candidatos(margem) {
    const lista = [];
    for (const el of document.querySelectorAll(SELETOR)) {
      if (el === atual) continue;
      const r = retangulo(el, margem);
      if (r) lista.push({ el, r });
    }
    // Se um item está dentro de outro com quase o mesmo tamanho, fica só o de fora.
    return lista.filter(a => !lista.some(b => b !== a && b.el.contains(a.el) &&
      Math.abs(b.r.width - a.r.width) < 24 && Math.abs(b.r.height - a.r.height) < 24));
  }

  function centro(r) { return { x: r.left + r.width / 2, y: r.top + r.height / 2 }; }

  function melhor(dir, margem) {
    const de = atual.getBoundingClientRect(), c = centro(de);
    let escolhido = null, pontos = Infinity;
    for (const { el, r } of candidatos(margem)) {
      if (el.contains(atual) || atual.contains(el)) continue;
      const p = centro(r);
      let frente, lado;
      if (dir === 'direita') { frente = r.left - de.right; if (p.x <= c.x + 4) continue; lado = sobra(r.top, r.bottom, de.top, de.bottom, p.y, c.y); }
      if (dir === 'esquerda') { frente = de.left - r.right; if (p.x >= c.x - 4) continue; lado = sobra(r.top, r.bottom, de.top, de.bottom, p.y, c.y); }
      if (dir === 'baixo') { frente = r.top - de.bottom; if (p.y <= c.y + 4) continue; lado = sobra(r.left, r.right, de.left, de.right, p.x, c.x); }
      if (dir === 'cima') { frente = de.top - r.bottom; if (p.y >= c.y - 4) continue; lado = sobra(r.left, r.right, de.left, de.right, p.x, c.x); }
      const valor = Math.max(frente, -20) + lado * (dir === 'direita' || dir === 'esquerda' ? 4 : 1.5);
      if (valor < pontos) { pontos = valor; escolhido = el; }
    }
    return escolhido;
  }

  // Distância lateral: zero se os dois se sobrepõem nesse eixo.
  function sobra(a1, a2, b1, b2, pa, pb) {
    if (a1 < b2 && a2 > b1) return 0;
    return Math.abs(pa - pb);
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
      el.scrollIntoView({ block: 'center', inline: 'center', behavior: 'smooth' });
    }
  }

  function primeiro() {
    const lista = candidatos(0).filter(({ r }) => r.top >= 0 && r.bottom <= innerHeight);
    lista.sort((a, b) => (a.r.top - b.r.top) || (a.r.left - b.r.left));
    // Pula o menu do topo quando houver conteúdo
    return (lista.find(({ r }) => r.top > innerHeight * 0.15) || lista[0] || {}).el || null;
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

  function clicar(el) {
    const r = el.getBoundingClientRect(), o = { bubbles: true, cancelable: true, clientX: r.left + r.width / 2, clientY: r.top + r.height / 2, button: 0 };
    el.dispatchEvent(new PointerEvent('pointerdown', o));
    el.dispatchEvent(new MouseEvent('mousedown', o));
    el.dispatchEvent(new PointerEvent('pointerup', o));
    el.dispatchEvent(new MouseEvent('mouseup', o));
    el.click();
  }

  const DIRS = { ArrowUp: 'cima', ArrowDown: 'baixo', ArrowLeft: 'esquerda', ArrowRight: 'direita' };

  window.addEventListener('keydown', e => {
    if (e.ctrlKey || e.altKey || e.metaKey) return;
    const dir = DIRS[e.key];
    const foco = document.activeElement;

    // Vídeo em tela grande: o player cuida das setas; Enter pausa; Esc volta.
    const v = videoGrande();
    if (v) {
      if (document.fullscreenElement) return;
      if (e.key === 'Enter') { v.paused ? v.play() : v.pause(); parar(e); }
      else if (e.key === 'Escape') { history.back(); parar(e); }
      return;
    }

    // Digitando num campo: esquerda/direita e Enter ficam com o campo; cima/baixo saem dele.
    if (ehCampo(foco)) {
      if (e.key === 'Escape') { foco.blur(); parar(e); return; }
      if (!dir || dir === 'esquerda' || dir === 'direita') return;
    }

    if (dir) {
      parar(e);
      if (!atual || !atual.isConnected || !retangulo(atual, innerHeight)) { marcar(primeiro()); return; }
      const prox = melhor(dir, 0) || melhor(dir, innerHeight);
      if (prox) marcar(prox);
      else { rolarPara(dir); setTimeout(() => { const p = melhor(dir, 0); if (p) marcar(p); }, 400); }
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
