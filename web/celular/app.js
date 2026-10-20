// Controle da TV pelo celular.
const $ = s => document.querySelector(s);
const $$ = s => [...document.querySelectorAll(s)];
let token = localStorage.getItem('tvc-token') || '';
let ws = null, espera = 500, fila = [], versaoServidor = '';

const vibrar = (ms = 12) => navigator.vibrate && navigator.vibrate(ms);

function tela(nome) {
  $('#tela-parear').classList.toggle('oculto', nome !== 'parear');
  $('#tela-controle').classList.toggle('oculto', nome !== 'controle');
  $('#tela-ajustes').classList.toggle('oculto', nome !== 'ajustes');
}

async function api(url, opts = {}) {
  const r = await fetch(url, { ...opts, headers: { 'Content-Type': 'application/json', 'X-Token': token, ...(opts.headers || {}) } });
  const j = await r.json().catch(() => ({}));
  if (!r.ok) { const e = new Error(j.erro || r.statusText); e.status = r.status; throw e; }
  return j;
}

// ---------- Pareamento ----------
$('#pedir-codigo').onclick = async () => {
  $('#erro-parear').textContent = '';
  try { await api('api/parear/iniciar', { method: 'POST' }); $('#codigo').focus(); }
  catch (e) { $('#erro-parear').textContent = e.message; }
};
$('#form-codigo').onsubmit = async ev => {
  ev.preventDefault();
  try {
    const nome = /android/i.test(navigator.userAgent) ? 'Android' : /iphone|ipad/i.test(navigator.userAgent) ? 'iPhone' : 'Celular';
    const j = await api('api/parear', { method: 'POST', body: JSON.stringify({ codigo: $('#codigo').value, nome }) });
    token = j.token;
    localStorage.setItem('tvc-token', token);
    iniciar();
  } catch (e) { $('#erro-parear').textContent = e.message; vibrar([30, 40, 30]); }
};

// ---------- Conexão ----------
function conectar() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  const este = ws = new WebSocket(`${proto}://${location.host}/ws?token=${encodeURIComponent(token)}`);
  ws.onopen = () => {
    if (ws !== este) return;
    espera = 500;
    $('#status').classList.add('on');
    $('#reconectando').classList.add('oculto');
    fila.splice(0).forEach(m => ws.send(m));
  };
  ws.onclose = async () => {
    if (ws !== este) return; // conexão antiga: já existe outra
    $('#status').classList.remove('on');
    $('#reconectando').classList.remove('oculto');
    try { await api('api/eu'); } catch (e) { if (e.status === 401) { localStorage.removeItem('tvc-token'); token = ''; tela('parear'); return; } }
    // Continua tentando para sempre (o PC pode estar atualizando ou reiniciando)
    clearTimeout(conectar.t);
    conectar.t = setTimeout(conectar, espera);
    espera = Math.min(espera * 2, 3000);
  };
  ws.onmessage = ev => {
    const m = JSON.parse(ev.data);
    if (m.erro) console.warn(m.erro);
    if (m.t === 'pong' && m.id) mostrarLatencia(Date.now() - m.id);
    if (m.t === 'pong') mostrarEstado(m);
    // O PC foi atualizado: recarrega o controle para pegar a versão nova
    if (m.versao) { if (!versaoServidor) versaoServidor = m.versao; else if (m.versao !== versaoServidor) location.reload(); }
  };
}

function enviar(m) {
  const s = JSON.stringify(m);
  if (ws && ws.readyState === 1) ws.send(s);
  else if (m.t !== 'tecla' && m.t !== 'mover' && m.t !== 'rolar' && fila.length < 5) fila.push(s);
}

// Mede a latência (tempo de ida e volta até o PC) a cada 3 s com a tela ligada
function mostrarLatencia(ms) {
  const el = $('#latencia');
  el.textContent = ms + ' ms';
  el.className = ms < 80 ? 'boa' : ms < 250 ? 'media' : 'ruim';
}
setInterval(() => { if (!document.hidden && ws && ws.readyState === 1) ws.send(JSON.stringify({ t: 'ping', id: Date.now() })); }, 3000);
const reconectarJa = () => { if (token && (!ws || ws.readyState > 1)) { espera = 500; conectar(); } };
document.addEventListener('visibilitychange', () => { if (!document.hidden) reconectarJa(); });
window.addEventListener('online', reconectarJa);
window.addEventListener('pageshow', reconectarJa);

// ---------- Botões (segurar repete) ----------
const pararTodos = new Set();
document.addEventListener('visibilitychange', () => { if (document.hidden) pararTodos.forEach(f => f()); });
function botaoRepete(el, acao) {
  let t1, t2;
  const parar = () => { clearTimeout(t1); clearInterval(t2); };
  pararTodos.add(parar);
  el.addEventListener('pointerdown', e => {
    e.preventDefault();
    parar(); // dois dedos no mesmo botão não deixam repetição presa
    acao(); vibrar();
    if (el.hasAttribute('data-repete')) t1 = setTimeout(() => (t2 = setInterval(acao, 110)), 400);
  });
  ['pointerup', 'pointerleave', 'pointercancel'].forEach(n => el.addEventListener(n, parar));
  el.addEventListener('contextmenu', e => e.preventDefault());
}

$$('[data-tecla]').forEach(b => botaoRepete(b, () => { enviar({ t: 'tecla', k: b.dataset.tecla }); if (b.dataset.tecla === 'ok') vibrar(25); }));
$$('[data-acao="inicio"], #btn-inicio-topo').forEach(b => botaoRepete(b, () => enviar({ t: 'inicio' })));
$$('[data-clique]').forEach(b => botaoRepete(b, () => enviar({ t: 'clique', k: b.dataset.clique })));
$$('[data-sistema]').forEach(b => b.onclick = () => {
  if (confirm(b.dataset.sistema === 'desligar' ? 'Desligar o computador?' : 'Reiniciar o computador?')) enviar({ t: 'sistema', k: b.dataset.sistema });
});

// ---------- Texto: envia, aperta Enter e limpa o campo ----------
$('#form-texto').onsubmit = ev => {
  ev.preventDefault();
  const campo = $('#campo-texto');
  if (!campo.value) { enviar({ t: 'tecla', k: 'ok' }); return; }
  enviar({ t: 'texto', s: campo.value, enter: true });
  campo.value = '';
  vibrar();
};

// ---------- Voz ----------
const Reconhecer = window.SpeechRecognition || window.webkitSpeechRecognition;
if (Reconhecer) {
  const btn = $('#btn-voz');
  btn.classList.remove('oculto');
  btn.onclick = () => {
    const r = new Reconhecer();
    r.lang = 'pt-BR';
    r.interimResults = false;
    btn.textContent = '…';
    r.onresult = ev => {
      const frase = ev.results[0][0].transcript;
      enviar({ t: 'texto', s: frase, enter: true });
      vibrar(20);
    };
    r.onend = () => (btn.textContent = '🎤');
    r.onerror = () => (btn.textContent = '🎤');
    r.start();
  };
}

// ---------- Touchpad ----------
(() => {
  const pad = $('#touchpad');
  const dedos = new Map();
  let moveu = false, inicio = 0, acumX = 0, acumY = 0, rolagem = 0;
  pad.addEventListener('pointerdown', e => {
    pad.setPointerCapture(e.pointerId);
    dedos.set(e.pointerId, { x: e.clientX, y: e.clientY });
    if (dedos.size === 1) { moveu = false; inicio = Date.now(); }
  });
  pad.addEventListener('pointermove', e => {
    const d = dedos.get(e.pointerId);
    if (!d) return;
    const dx = e.clientX - d.x, dy = e.clientY - d.y;
    d.x = e.clientX; d.y = e.clientY;
    if (Math.abs(dx) + Math.abs(dy) > 1) moveu = true;
    if (dedos.size >= 2) {
      rolagem += dy;
      if (Math.abs(rolagem) > 18) { enviar({ t: 'rolar', dy: -Math.sign(rolagem) }); rolagem = 0; }
      return;
    }
    // Aceleração: movimento rápido anda mais
    const v = Math.hypot(dx, dy), f = v > 12 ? 2.6 : v > 5 ? 1.8 : 1.2;
    acumX += dx * f; acumY += dy * f;
    const ix = Math.trunc(acumX), iy = Math.trunc(acumY);
    if (ix || iy) { enviar({ t: 'mover', dx: ix, dy: iy }); acumX -= ix; acumY -= iy; }
  });
  const fim = e => {
    const eraUnico = dedos.size === 1;
    dedos.delete(e.pointerId);
    if (eraUnico && !moveu && Date.now() - inicio < 250) { enviar({ t: 'clique', k: 'esquerdo' }); vibrar(); }
  };
  pad.addEventListener('pointerup', fim);
  pad.addEventListener('pointercancel', e => dedos.delete(e.pointerId));
})();

// ---------- O que está na TV e timer ----------
const nomesApps = {};
function mostrarEstado(m) {
  const el = $('#na-tv');
  el.classList.toggle('oculto', !m.atual);
  if (m.atual) el.innerHTML = 'Na TV agora: <b></b>', el.querySelector('b').textContent = nomesApps[m.atual] || m.atual;
  $('#timer-cel').textContent = m.timer ? Math.ceil(m.timer / 60) + ' min' : 'desativado';
}
$$('[data-timer]').forEach(b => b.onclick = () => {
  enviar({ t: 'timer', min: +b.dataset.timer }); vibrar(20);
  setTimeout(() => ws && ws.readyState === 1 && ws.send(JSON.stringify({ t: 'ping', id: Date.now() })), 150);
});

// ---------- Gestos: deslizar = setas, toque = OK, segurar = voltar ----------
(() => {
  const area = $('#gestos'), seta = $('#gestos-seta');
  const PASSO = 45; // pixels de deslize para cada seta
  let ini = null, moveu = false, tSegurar = null;
  const mostrarSeta = txt => { seta.textContent = txt; seta.classList.add('ver'); clearTimeout(mostrarSeta.t); mostrarSeta.t = setTimeout(() => seta.classList.remove('ver'), 160); };
  area.addEventListener('pointerdown', e => {
    area.setPointerCapture(e.pointerId);
    ini = { x: e.clientX, y: e.clientY }; moveu = false;
    tSegurar = setTimeout(() => { if (!moveu) { enviar({ t: 'tecla', k: 'voltar' }); vibrar(40); mostrarSeta('↩'); ini = null; } }, 550);
  });
  area.addEventListener('pointermove', e => {
    if (!ini) return;
    const dx = e.clientX - ini.x, dy = e.clientY - ini.y;
    if (Math.max(Math.abs(dx), Math.abs(dy)) < PASSO) return;
    moveu = true; clearTimeout(tSegurar);
    const [k, s] = Math.abs(dx) > Math.abs(dy) ? (dx > 0 ? ['direita', '→'] : ['esquerda', '←']) : (dy > 0 ? ['baixo', '↓'] : ['cima', '↑']);
    enviar({ t: 'tecla', k }); vibrar(8); mostrarSeta(s);
    ini = { x: e.clientX, y: e.clientY }; // continuar deslizando anda mais
  });
  const fim = () => {
    clearTimeout(tSegurar);
    if (ini && !moveu) { enviar({ t: 'tecla', k: 'ok' }); vibrar(25); mostrarSeta('●'); }
    ini = null;
  };
  area.addEventListener('pointerup', fim);
  area.addEventListener('pointercancel', () => { clearTimeout(tSegurar); ini = null; });
})();

// ---------- Tela do celular acesa enquanto o controle está aberto ----------
let travaTela = null;
async function manterTelaAcesa() {
  try { if ('wakeLock' in navigator && !document.hidden) travaTela = await navigator.wakeLock.request('screen'); } catch (_) {}
}
document.addEventListener('visibilitychange', () => { if (!document.hidden) manterTelaAcesa(); });

// ---------- Apps ----------
async function carregarApps() {
  try {
    const { apps, crianca } = await api('api/apps');
    // Modo criança: sem desligar, reiniciar e timer
    document.querySelectorAll('#aba-apps .sistema, #aba-apps .timer, #aba-apps .sem-crianca').forEach(el => el.classList.toggle('oculto', !!crianca));
    if (!crianca) { carregarDepois(); carregarFotos(); }
    const lista = $('#lista-apps');
    lista.innerHTML = '';
    for (const a of apps) {
      nomesApps[a.ID] = a.Nome;
      const b = document.createElement('button');
      b.textContent = a.Nome;
      b.style.setProperty('--c1', a.Cor);
      b.style.setProperty('--c2', a.Cor2 || '#111');
      b.style.background = `linear-gradient(180deg, transparent 40%, rgba(0,0,0,.6)), url("/capa/${a.ID}?p=1") center / cover, linear-gradient(135deg, ${a.Cor}, ${a.Cor2 || '#111'})`;
      b.onclick = () => { enviar({ t: 'abrir', k: a.ID }); vibrar(20); trocarAba('controle'); };
      lista.append(b);
    }
  } catch (_) {}
}

// ---------- Assistir depois ----------
function el(tag, props = {}, ...filhos) {
  const e = document.createElement(tag);
  for (const [k, v] of Object.entries(props)) { if (k.startsWith('on')) e.addEventListener(k.slice(2), v); else if (k === 'class') e.className = v; else if (k === 'style') e.style.cssText = v; else e.setAttribute(k, v); }
  e.append(...filhos.filter(f => f != null));
  return e;
}
function aviso(txt) {
  const t = el('div', { class: 'toast' }, txt); document.body.append(t); setTimeout(() => t.remove(), 2500);
}
async function carregarDepois() {
  try {
    const itens = (await api('api/depois')).itens || [];
    $('#lista-depois').replaceChildren(...itens.map(it => el('div', { class: 'item-depois' },
      el('div', {}, el('b', {}, it.titulo), el('small', {}, new URL(it.url).hostname.replace(/^www\./, ''))),
      el('button', { class: 'aj-btn perigo', onclick: async () => { try { await api('api/depois/' + it.id, { method: 'DELETE' }); carregarDepois(); } catch (e) { aviso(e.message); } } }, '✕'))));
  } catch (_) {}
}
async function salvarDepois(texto) {
  try { await api('api/depois', { method: 'POST', body: JSON.stringify({ texto }) }); aviso('Salvo em Assistir depois ✓'); vibrar(25); carregarDepois(); return true; }
  catch (e) { aviso(e.message); return false; }
}
$('#form-depois').onsubmit = async ev => {
  ev.preventDefault();
  const c = $('#campo-depois');
  if (c.value.trim() && await salvarDepois(c.value)) c.value = '';
};
// Compartilhar → TV Central (o Android abre /compartilhar?url=...&texto=...)
async function receberCompartilhado() {
  if (location.pathname !== '/compartilhar') return;
  const p = new URLSearchParams(location.search);
  const texto = [p.get('url'), p.get('texto'), p.get('titulo')].filter(Boolean).join(' ');
  history.replaceState(null, '', '/');
  if (texto) { trocarAba('apps'); await salvarDepois(texto); }
}

// ---------- Fotos ----------
async function carregarFotos() {
  try {
    const fotos = (await api('api/fotos')).fotos || [];
    $('#fotos-status').textContent = fotos.length ? `${fotos.length} na TV` : 'nenhuma ainda';
    $('#grade-fotos').replaceChildren(...fotos.slice(0, 60).map(n => el('div', {
      class: 'ft', style: `background-image:url("/foto/${encodeURIComponent(n)}?token=${encodeURIComponent(token)}")`,
      onclick: async () => { try { await api('api/fotos/mostrar', { method: 'POST', body: JSON.stringify({ nome: n }) }); aviso('Mostrando na TV'); vibrar(20); } catch (e) { aviso(e.message); } },
    }, el('button', { class: 'x', onclick: async ev => {
      ev.stopPropagation();
      if (!confirm('Apagar esta foto da TV?')) return;
      try { await api('api/fotos/' + encodeURIComponent(n), { method: 'DELETE' }); carregarFotos(); } catch (e) { aviso(e.message); }
    } }, '✕'))));
  } catch (_) {}
}
$('#enviar-fotos').onclick = () => {
  const inp = el('input', { type: 'file', accept: 'image/jpeg,image/png,image/webp,image/gif', multiple: '', style: 'display:none' });
  inp.onchange = () => {
    const arquivos = [...inp.files]; inp.remove();
    if (!arquivos.length) return;
    const fd = new FormData(); arquivos.forEach(f => fd.append('foto', f, f.name));
    const prog = $('#fotos-prog'), barra = prog.querySelector('i');
    prog.classList.remove('oculto'); barra.style.width = '0';
    const x = new XMLHttpRequest();
    x.open('POST', 'api/fotos'); x.setRequestHeader('X-Token', token);
    x.upload.onprogress = e => { if (e.lengthComputable) barra.style.width = (100 * e.loaded / e.total) + '%'; };
    x.onload = () => {
      prog.classList.add('oculto');
      let j = {}; try { j = JSON.parse(x.responseText); } catch (_) {}
      if (x.status !== 200) { aviso(j.erro || 'Não enviou'); return; }
      const salvas = j.salvas || [], falhas = j.falhas || [];
      aviso(`${salvas.length} foto(s) enviada(s)` + (falhas.length ? ` · ${falhas.length} com problema` : ''));
      carregarFotos();
    };
    x.onerror = () => { prog.classList.add('oculto'); aviso('Sem conexão'); };
    x.send(fd);
  };
  document.body.append(inp); inp.click();
};

// ---------- Abas ----------
function trocarAba(nome) {
  $$('.abas button').forEach(b => b.classList.toggle('ativa', b.dataset.aba === nome));
  ['controle', 'gestos', 'touch', 'apps'].forEach(n => $('#aba-' + n).classList.toggle('oculto', n !== nome));
  try { localStorage.setItem('aba', nome); } catch (_) {}
  if (nome === 'apps') carregarApps();
}
$$('.abas button').forEach(b => b.onclick = () => trocarAba(b.dataset.aba));

// Diagnóstico: versão, rede e últimas linhas do registro do PC
$('#diag').addEventListener('toggle', async () => {
  if (!$('#diag').open) return;
  try {
    const d = await api('api/diagnostico');
    $('#diag-texto').textContent =
      `Versão: ${d.versao}\nNavegador: ${d.navegador || 'não encontrado'}\nEndereço: ${(d.enderecos || []).join(', ')}\n` +
      `App aberto: ${d.appAberto || 'nenhum'}\nTela inicial: ${d.telaInicial ? 'aberta' : 'fechada'}\n` +
      (d.atualizacao ? `Atualização: ${d.atualizacao}\n` : '') + `\n--- Registro ---\n${(d.log || []).join('\n')}`;
  } catch (e) { $('#diag-texto').textContent = 'Erro: ' + e.message; }
});

// ---------- Início ----------
async function iniciar() {
  if (!token) { tela('parear'); return; }
  try { await api('api/eu'); } catch (e) {
    if (e.status === 401) { localStorage.removeItem('tvc-token'); token = ''; tela('parear'); return; }
  }
  tela('controle');
  conectar();
  receberCompartilhado();
  manterTelaAcesa();
  let aba = null; try { aba = localStorage.getItem('aba'); } catch (_) {}
  if (aba && aba !== 'apps' && $('#aba-' + aba)) trocarAba(aba); // volta na última aba usada
  carregarApps(); // para saber o nome do app que está na TV
}

if ('serviceWorker' in navigator) navigator.serviceWorker.register('sw.js').catch(() => {});
iniciar();
