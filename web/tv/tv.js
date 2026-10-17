// Tela inicial da TV: navegação por setas entre linhas de cartões.
const $ = s => document.querySelector(s);
const linhas = [[], []];
let pos = { linha: 0, col: 0 };
let modal = null; // 'confirmar' quando a caixa de confirmação está aberta
let confirmarAcao = null;
let appsJson = '';
let versaoInicial = null;

function aviso(texto, ms = 6000) {
  const el = $('#aviso');
  el.textContent = texto;
  mostrar('#aviso', true);
  clearTimeout(aviso.t);
  aviso.t = setTimeout(() => mostrar('#aviso', false), ms);
}

function saudacao() {
  const h = new Date().getHours();
  return h < 5 ? 'Boa noite' : h < 12 ? 'Bom dia' : h < 18 ? 'Boa tarde' : 'Boa noite';
}

function relogio() {
  const d = new Date();
  $('#relogio').textContent = d.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });
  const t = d.toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' });
  $('#data').textContent = t.charAt(0).toUpperCase() + t.slice(1);
}

function montarApps(apps) {
  const caixa = $('#apps');
  caixa.innerHTML = '';
  for (const a of apps) {
    const b = document.createElement('button');
    b.className = 'cartao';
    b.dataset.id = a.ID;
    b.dataset.nome = a.Nome;
    b.style.setProperty('--c1', a.Cor);
    b.style.setProperty('--c2', a.Cor2 || '#111');
    b.style.setProperty('--capa', `url("/capa/${a.ID}?p=1")`);
    b.dataset.desc = a.Descricao || '';
    new Image().src = `/capa/${a.ID}`; // já deixa a arte grande carregada
    const n = document.createElement('span');
    n.className = 'nome';
    n.textContent = a.Nome;
    b.append(n);
    caixa.append(b);
  }
  linhas[0] = [...caixa.children];
  linhas[1] = [...$('#sistema').children];
  let ultimo = null;
  try { ultimo = localStorage.getItem('ultimo-app'); } catch (_) {}
  const i = linhas[0].findIndex(b => b.dataset.id === ultimo);
  if (!montarApps.feito && i >= 0) pos = { linha: 0, col: i };
  montarApps.feito = true;
  pos.col = Math.min(pos.col, linhas[pos.linha].length - 1);
  focar();
}

function focar(peloMouse) {
  document.querySelectorAll('.cartao.foco').forEach(e => e.classList.remove('foco'));
  const el = linhas[pos.linha][pos.col];
  if (!el) return;
  el.classList.add('foco');
  // Rola a linha para o cartão focado ficar visível
  const fila = el.parentElement;
  const visivel = innerWidth - fila.parentElement.getBoundingClientRect().left - innerWidth * 0.09;
  const ini = el.offsetLeft, fim = ini + el.offsetWidth; // a fila tem position: relative, então offsetLeft é dentro dela
  let x = fila._x || 0;
  if (!peloMouse) { // com mouse a linha não pula: o cartão já está embaixo do cursor
    if (fim - x > visivel) x = fim - visivel;
    if (ini - x < 0) x = ini;
  }
  fila._x = x;
  fila.style.transform = `translateX(${-x}px)`;
  const raiz = document.documentElement.style;
  if (el.dataset.id) {
    raiz.setProperty('--cor', el.style.getPropertyValue('--c1'));
    raiz.setProperty('--cor2', el.style.getPropertyValue('--c2'));
    titulo(el.dataset.nome);
    $('#subtitulo').textContent = el.dataset.desc || 'Pressione OK para abrir';
    $('#dica').classList.add('ver');
    trocarArte(el.dataset.id);
  } else {
    raiz.setProperty('--cor', '#3a6df0');
    raiz.setProperty('--cor2', '#0b1430');
    titulo(saudacao());
    $('#subtitulo').textContent = el.querySelector('span:last-child').firstChild.textContent.trim();
    $('#dica').classList.add('ver');
    trocarArte(null);
  }
}

function titulo(t) {
  const h = $('#titulo');
  if (h.textContent === t) return;
  h.textContent = t;
  h.classList.remove('troca'); void h.offsetWidth; h.classList.add('troca');
}

// Fundo com a arte do app: espera um instante para não carregar imagem a cada seta rápida
let arteAtual = null, arteVez = 0;
function trocarArte(id) {
  clearTimeout(trocarArte.t);
  trocarArte.t = setTimeout(() => {
    if (id === arteAtual) return;
    arteAtual = id;
    const camadas = [$('#arte-a'), $('#arte-b')];
    const nova = camadas[arteVez = 1 - arteVez], velha = camadas[1 - arteVez];
    if (id) nova.style.backgroundImage = `url("/capa/${id}")`;
    nova.classList.toggle('ativa', !!id);
    velha.classList.remove('ativa');
  }, 140);
}

function mover(dl, dc) {
  if (dl) {
    const nl = pos.linha + dl;
    if (nl < 0 || nl >= linhas.length || !linhas[nl].length) return;
    // Mantém a posição horizontal aproximada ao trocar de linha
    const x = linhas[pos.linha][pos.col].getBoundingClientRect().left;
    let melhor = 0, dist = Infinity;
    linhas[nl].forEach((el, i) => { const d = Math.abs(el.getBoundingClientRect().left - x); if (d < dist) { dist = d; melhor = i; } });
    pos = { linha: nl, col: melhor };
  } else {
    pos.col = Math.max(0, Math.min(linhas[pos.linha].length - 1, pos.col + dc));
  }
  focar();
}

async function post(url, corpo) {
  const r = await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(corpo) });
  if (!r.ok) throw new Error((await r.json().catch(() => ({}))).erro || r.statusText);
}

async function ativar() {
  const el = linhas[pos.linha][pos.col];
  if (!el) return;
  if (el.dataset.id) {
    try { localStorage.setItem('ultimo-app', el.dataset.id); } catch (_) {}
    $('#abrindo-texto').textContent = `Abrindo ${el.dataset.nome}…`;
    mostrar('#abrindo', true);
    setTimeout(() => mostrar('#abrindo', false), 4000);
    try { await post('api/abrir', { id: el.dataset.id }); } catch (e) { $('#abrindo-texto').textContent = 'Não abriu: ' + e.message; }
    return;
  }
  const acao = el.dataset.acao;
  if (acao === 'celular') { $('#qr').classList.toggle('oculto'); return; }
  if (acao === 'atualizar') {
    aviso('Procurando versão nova…');
    try {
      const r = await fetch('api/atualizar', { method: 'POST' });
      aviso((await r.json()).mensagem || 'Pronto');
    } catch (_) { aviso('Sem conexão com o programa'); }
    return;
  }
  confirmarAcao = acao;
  $('#confirmar-texto').textContent = acao === 'desligar' ? 'Desligar o computador?' : 'Reiniciar o computador?';
  abrirModal();
}

function mostrar(sel, sim) { $(sel).classList.toggle('oculto', !sim); }

const botoesModal = () => [...document.querySelectorAll('#confirmar .botao')];
let focoModal = 1;
function abrirModal() { modal = 'confirmar'; focoModal = 1; mostrar('#confirmar', true); pintarModal(); }
function fecharModal() { modal = null; mostrar('#confirmar', false); }
function pintarModal() { botoesModal().forEach((b, i) => b.classList.toggle('foco', i === focoModal)); }

document.addEventListener('keydown', async e => {
  const k = e.key;
  if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Enter', 'Escape', ' '].includes(k)) e.preventDefault();
  if (modal) {
    if (k === 'ArrowLeft' || k === 'ArrowRight') { focoModal = 1 - focoModal; pintarModal(); }
    else if (k === 'Escape' || k === 'Backspace') fecharModal();
    else if (k === 'Enter') {
      const sim = focoModal === 0;
      fecharModal();
      if (sim) post('api/sistema', { acao: confirmarAcao }).catch(err => alert(err.message));
    }
    return;
  }
  if (k === 'ArrowLeft') mover(0, -1);
  else if (k === 'ArrowRight') mover(0, 1);
  else if (k === 'ArrowUp') mover(-1, 0);
  else if (k === 'ArrowDown') mover(1, 0);
  else if (k === 'Enter' || k === ' ') ativar();
  else if (k === 'Escape') $('#qr').classList.add('oculto');
});

// Mouse também funciona (modo touchpad do celular)
document.addEventListener('click', e => {
  const c = e.target.closest('.cartao');
  if (c) {
    linhas.forEach((l, li) => { const i = l.indexOf(c); if (i >= 0) pos = { linha: li, col: i }; });
    focar(); ativar();
  }
  const b = e.target.closest('#confirmar .botao');
  if (b) { focoModal = botoesModal().indexOf(b); document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' })); }
});
// Mouse: o cursor aparece ao mexer (inclusive em cima dos cartões) e some depois de parado.
// Passar o mouse por cima de um cartão já seleciona ele.
document.addEventListener('mousemove', e => {
  document.body.classList.add('com-mouse');
  clearTimeout(window.__cur);
  window.__cur = setTimeout(() => document.body.classList.remove('com-mouse'), 4000);
  const c = e.target.closest && e.target.closest('.cartao');
  if (c && !modal) {
    linhas.forEach((l, li) => { const i = l.indexOf(c); if (i >= 0 && (pos.linha !== li || pos.col !== i)) { pos = { linha: li, col: i }; focar(true); } });
  }
});
document.addEventListener('keydown', () => document.body.classList.remove('com-mouse'), true);

async function estado() {
  try {
    const r = await fetch('api/estado', { cache: 'no-store' });
    const s = await r.json();
    // Programa atualizado: recarrega a tela para pegar o visual novo
    if (versaoInicial === null) versaoInicial = s.versao;
    else if (s.versao !== versaoInicial) { location.reload(); return; }
    $('#versao').textContent = 'versão ' + s.versao;
    if (s.novidade && !sessionStorage.getItem('aviso-' + s.versao)) { sessionStorage.setItem('aviso-' + s.versao, '1'); aviso(s.novidade, 8000); }
    if (s.atualizacao && s.atualizacao !== estado.ultimoStatus) aviso(s.atualizacao);
    estado.ultimoStatus = s.atualizacao;
    const j = JSON.stringify(s.apps);
    if (j !== appsJson) { appsJson = j; montarApps(s.apps); }
    const end = (s.enderecos || [])[0] || '';
    if (end !== estado.end) { estado.end = end; $('#qr img').src = 'qr.png?e=' + encodeURIComponent(end); }
    $('#endereco').textContent = end || 'sem rede: conecte o Wi-Fi';
    if (s.celulares === 0 && !estado.mostrouQr) { $('#qr').classList.remove('oculto'); estado.mostrouQr = true; }
    $('#codigo').textContent = s.codigo || '';
    mostrar('#pareamento', !!s.codigo);
    if (s.codigo) $('#qr').classList.add('oculto');
  } catch (_) { /* servidor reiniciando */ }
}

relogio();
setInterval(relogio, 5000);
linhas[1] = [...$('#sistema').children];
estado();
setInterval(estado, 1500);
window.addEventListener('focus', () => mostrar('#abrindo', false));
