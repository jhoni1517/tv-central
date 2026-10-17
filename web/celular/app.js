// Controle da TV pelo celular.
const $ = s => document.querySelector(s);
const $$ = s => [...document.querySelectorAll(s)];
let token = localStorage.getItem('tvc-token') || '';
let ws = null, espera = 500, fila = [], versaoServidor = '';

const vibrar = (ms = 12) => navigator.vibrate && navigator.vibrate(ms);

function tela(nome) {
  $('#tela-parear').classList.toggle('oculto', nome !== 'parear');
  $('#tela-controle').classList.toggle('oculto', nome !== 'controle');
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
  ws = new WebSocket(`${proto}://${location.host}/ws?token=${encodeURIComponent(token)}`);
  ws.onopen = () => {
    espera = 500;
    $('#status').classList.add('on');
    $('#reconectando').classList.add('oculto');
    fila.splice(0).forEach(m => ws.send(m));
  };
  ws.onclose = async () => {
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
function botaoRepete(el, acao) {
  let t1, t2;
  const parar = () => { clearTimeout(t1); clearInterval(t2); };
  el.addEventListener('pointerdown', e => {
    e.preventDefault();
    acao(); vibrar();
    if (el.hasAttribute('data-repete')) t1 = setTimeout(() => (t2 = setInterval(acao, 110)), 400);
  });
  ['pointerup', 'pointerleave', 'pointercancel'].forEach(n => el.addEventListener(n, parar));
  el.addEventListener('contextmenu', e => e.preventDefault());
}

$$('[data-tecla]').forEach(b => botaoRepete(b, () => enviar({ t: 'tecla', k: b.dataset.tecla })));
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

// ---------- Apps ----------
async function carregarApps() {
  try {
    const { apps } = await api('api/apps');
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

// ---------- Abas ----------
function trocarAba(nome) {
  $$('.abas button').forEach(b => b.classList.toggle('ativa', b.dataset.aba === nome));
  ['controle', 'touch', 'apps'].forEach(n => $('#aba-' + n).classList.toggle('oculto', n !== nome));
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
  carregarApps(); // para saber o nome do app que está na TV
}

if ('serviceWorker' in navigator) navigator.serviceWorker.register('sw.js').catch(() => {});
iniciar();
