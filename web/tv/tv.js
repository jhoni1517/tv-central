// Tela inicial da TV: navegação por setas entre linhas de cartões.
const $ = s => document.querySelector(s);
const linhas = [[], [], []]; // apps, assistir depois, sistema
let pos = { linha: 0, col: 0 };
let modal = null; // 'confirmar' (caixa de confirmação) ou 'perfil' (quem está assistindo)
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
  $('#relogio').textContent = d.toLocaleTimeString(window.TVC_LOCALE || 'pt-BR', { hour: '2-digit', minute: '2-digit' });
  const t = d.toLocaleDateString(window.TVC_LOCALE || 'pt-BR', { weekday: 'long', day: 'numeric', month: 'long' });
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
  linhas[2] = sistemaVisivel();
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
  rolarTela(el);
  const raiz = document.documentElement.style;
  if (el.dataset.depois) {
    raiz.setProperty('--cor', '#3a6df0'); raiz.setProperty('--cor2', '#0b1430');
    titulo(el.dataset.titulo);
    $('#subtitulo').textContent = 'Assistir depois · ' + el.lastChild.textContent;
    $('#dica').classList.add('ver');
    trocarArte(el.dataset.app || null);
  } else if (el.dataset.id) {
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

// Com várias linhas, a tela sobe para a linha em foco caber (e volta ao topo na primeira linha)
function rolarTela(el) {
  const main = document.querySelector('main');
  const linha = el.closest('.linha');
  let y = main._y || 0;
  if (pos.linha === 0) y = 0;
  else {
    const r = linha.getBoundingClientRect(), baixo = r.bottom + y, limite = innerHeight * 0.9;
    y = Math.max(0, baixo - limite);
  }
  if (y === main._y) return;
  main._y = y;
  main.style.transform = y ? `translateY(${-y}px)` : '';
  $('#rodape').style.opacity = y ? '0' : '';
  document.querySelector('.destaque').style.opacity = y ? '0' : ''; // o título não passa por cima do topo
}

function titulo(t) {
  const h = $('#titulo');
  if (h.textContent === t) return;
  h.textContent = t;
  h.classList.remove('troca'); void h.offsetWidth; h.classList.add('troca');
}

// Fundo com a arte (ou o vídeo) do app: espera um instante para não carregar a cada seta rápida
let arteAtual = null, arteVez = 0, videosFundo = new Set(), videosLigados = true;
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
    tocarVideoFundo(id || 'fundo');
  }, 140);
}

// Vídeo só toca com a tela inicial na frente: com um app aberto ele para (economiza a máquina)
const videoFundo = () => $('#video-fundo');
function tocarVideoFundo(id) {
  const v = videoFundo();
  clearTimeout(tocarVideoFundo.t);
  if (!videosLigados || !videosFundo.has(id) || document.body.classList.contains('leve')) { pararVideo(v); return; }
  tocarVideoFundo.t = setTimeout(() => {
    const src = `/video/${id}`;
    if (v.dataset.id !== id) {
      v.classList.remove('ativa');
      v.dataset.id = id;
      v.src = src;
    }
    if (document.hasFocus() && !document.hidden) v.play().catch(() => {});
  }, 250); // um pouco depois da imagem: troca de vídeo pesa mais
}
function pararVideo(v) {
  v.classList.remove('ativa');
  if (v.id === 'video-fundo') $('#fundo').classList.remove('com-video');
  v.pause();
  if (v.dataset.id) { delete v.dataset.id; v.removeAttribute('src'); v.load(); } // libera o decodificador
}
videoFundo().addEventListener('playing', e => { e.target.classList.add('ativa'); $('#fundo').classList.add('com-video'); });
videoFundo().addEventListener('error', e => pararVideo(e.target));
window.addEventListener('blur', () => videoFundo().pause());
document.addEventListener('visibilitychange', () => { if (document.hidden) videoFundo().pause(); else if (videoFundo().dataset.id) videoFundo().play().catch(() => {}); });
window.addEventListener('focus', () => { if (videoFundo().dataset.id) videoFundo().play().catch(() => {}); });

function mover(dl, dc) {
  if (dl) {
    let nl = pos.linha + dl;
    while (nl >= 0 && nl < linhas.length && !linhas[nl].length) nl += dl; // pula linhas vazias
    if (nl < 0 || nl >= linhas.length) { Som.limite(); return; }
    // Mantém a posição horizontal aproximada ao trocar de linha
    const x = linhas[pos.linha][pos.col].getBoundingClientRect().left;
    let melhor = 0, dist = Infinity;
    linhas[nl].forEach((el, i) => { const d = Math.abs(el.getBoundingClientRect().left - x); if (d < dist) { dist = d; melhor = i; } });
    pos = { linha: nl, col: melhor };
  } else {
    const nc = Math.max(0, Math.min(linhas[pos.linha].length - 1, pos.col + dc));
    if (nc === pos.col) { Som.limite(); return; }
    pos.col = nc;
  }
  Som.mover();
  focar();
}

// Botões de sistema que estão aparecendo (alguns somem conforme o perfil / modo criança)
function sistemaVisivel() {
  if (document.body.classList.contains('modo-crianca')) return [];
  return [...$('#sistema').children].filter(b => !b.classList.contains('oculto'));
}

// ---------- Quem está assistindo ----------
let perfis = [], perfilAtualId = null, focoPerfil = 0;
function abrirPerfis() {
  if (!perfis.length) return;
  modal = 'perfil';
  focoPerfil = Math.max(0, perfis.findIndex(p => p.id === perfilAtualId));
  const lista = $('#quem-lista');
  lista.replaceChildren(...perfis.map((p, i) => {
    const item = document.createElement('button');
    item.className = 'quem-item';
    const av = document.createElement('span');
    av.className = 'av'; av.style.background = p.cor; av.textContent = p.emoji;
    item.append(av, document.createTextNode(p.nome));
    item.onclick = () => { focoPerfil = i; escolherPerfil(); };
    item.onmousemove = () => { focoPerfil = i; pintarPerfis(); };
    return item;
  }));
  mostrar('#quem', true);
  pintarPerfis();
}
function pintarPerfis() { [...$('#quem-lista').children].forEach((el, i) => el.classList.toggle('foco', i === focoPerfil)); }
async function escolherPerfil() {
  const p = perfis[focoPerfil];
  modal = null; mostrar('#quem', false); Som.ok();
  if (p.id === perfilAtualId) return;
  try { await post('api/perfil', { id: p.id }); aviso(`${p.emoji} Olá, ${p.nome}!`, 2500); estado(); }
  catch (e) { aviso('Não deu: ' + e.message); }
}

async function post(url, corpo) {
  const r = await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(corpo) });
  if (!r.ok) throw new Error((await r.json().catch(() => ({}))).erro || r.statusText);
}

async function ativar() {
  const el = linhas[pos.linha][pos.col];
  if (!el) return;
  Som.ok();
  if (el.dataset.id) {
    if (Date.now() - (ativar.ultima || 0) < 2500) return; // já está abrindo
    ativar.ultima = Date.now();
    try { localStorage.setItem('ultimo-app', el.dataset.id); } catch (_) {}
    $('#abrindo-texto').textContent = `Abrindo ${el.dataset.nome}…`;
    $('#abrindo').style.backgroundImage = `linear-gradient(rgba(0,0,0,.55), rgba(0,0,0,.55)), url("/capa/${el.dataset.id}")`;
    mostrar('#abrindo', true);
    setTimeout(() => mostrar('#abrindo', false), 4000);
    try { await post('api/abrir', { id: el.dataset.id }); } catch (e) { $('#abrindo-texto').textContent = 'Não abriu: ' + e.message; }
    return;
  }
  if (el.dataset.depois) {
    $('#abrindo-texto').textContent = `Abrindo ${el.dataset.titulo}…`;
    $('#abrindo').style.backgroundImage = `linear-gradient(rgba(0,0,0,.55), rgba(0,0,0,.55)), url("/capa/${el.dataset.app || 'padrao'}")`;
    mostrar('#abrindo', true);
    setTimeout(() => mostrar('#abrindo', false), 4000);
    try { await post('api/depois/abrir', { id: el.dataset.depois }); } catch (e) { $('#abrindo-texto').textContent = 'Não abriu: ' + e.message; }
    return;
  }
  const acao = el.dataset.acao;
  if (acao === 'fotos') { abrirVisor(0); return; }
  if (acao === 'espelhar') {
    mostrar('#espelho', true); modal = 'espelho';
    try { await post('api/espelhar', {}); } catch (e) { aviso(e.message, 9000); }
    setTimeout(() => { if (modal === 'espelho') { modal = null; mostrar('#espelho', false); } }, 15000);
    return;
  }
  if (acao === 'celular') { $('#qr').classList.toggle('oculto'); return; }
  if (acao === 'timer') { proximoTimer(); return; }
  if (acao === 'perfil') { abrirPerfis(); return; }
  if (acao === 'desktop') {
    aviso('Indo para a área de trabalho. Aperte Início no celular para voltar.', 3000);
    setTimeout(() => post('api/desktop', {}).catch(() => {}), 800);
    return;
  }
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
  const k = e.key === 'BrowserBack' || e.key === 'GoBack' ? 'Escape' : e.key;
  if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Enter', 'Escape', ' '].includes(k)) e.preventDefault();
  if (modal === 'visor') {
    if (k === 'ArrowRight' || k === 'ArrowDown') passarFoto(1);
    else if (k === 'ArrowLeft' || k === 'ArrowUp') passarFoto(-1);
    else if (k === 'Escape' || k === 'Enter') fecharVisor();
    return;
  }
  if (modal === 'espelho') { modal = null; mostrar('#espelho', false); return; }
  if (modal === 'perfil') {
    if (k === 'ArrowLeft') { focoPerfil = Math.max(0, focoPerfil - 1); Som.mover(); pintarPerfis(); }
    else if (k === 'ArrowRight') { focoPerfil = Math.min(perfis.length - 1, focoPerfil + 1); Som.mover(); pintarPerfis(); }
    else if (k === 'Enter' || k === ' ') escolherPerfil();
    else if (k === 'Escape') { modal = null; mostrar('#quem', false); Som.voltar(); }
    return;
  }
  if (modal) {
    if (k === 'ArrowLeft' || k === 'ArrowRight') { focoModal = 1 - focoModal; pintarModal(); }
    else if (k === 'Escape' || k === 'Backspace') fecharModal();
    else if (k === 'Enter') {
      const sim = focoModal === 0;
      fecharModal();
      if (sim) post('api/sistema', { acao: confirmarAcao }).catch(err => aviso('Não deu: ' + err.message));
    }
    return;
  }
  if (k === 'ArrowLeft') mover(0, -1);
  else if (k === 'ArrowRight') mover(0, 1);
  else if (k === 'ArrowUp') mover(-1, 0);
  else if (k === 'ArrowDown') mover(1, 0);
  else if (k === 'Enter' || k === ' ') ativar();
  else if (k === 'Escape') { Som.voltar(); $('#qr').classList.add('oculto'); }
});

// Mouse também funciona (modo touchpad do celular)
document.addEventListener('click', e => {
  const c = e.target.closest('.cartao');
  if (c && !modal) {
    linhas.forEach((l, li) => { const i = l.indexOf(c); if (i >= 0) pos = { linha: li, col: i }; });
    focar(); ativar();
  }
  const b = e.target.closest('#confirmar .botao');
  if (b) { focoModal = botoesModal().indexOf(b); document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' })); }
});
// Mouse: o cursor aparece ao mexer (inclusive em cima dos cartões) e some depois de parado.
// Passar o mouse por cima de um cartão já seleciona ele.
let ultimoMouse = { x: 0, y: 0 };
const mouseMexeu = e => { const m = Math.abs(e.screenX - ultimoMouse.x) + Math.abs(e.screenY - ultimoMouse.y) > 4; ultimoMouse = { x: e.screenX, y: e.screenY }; return m; };
// Rodinha do mouse anda entre os cartões
document.addEventListener('wheel', e => { if (!modal && Math.abs(e.deltaY) > 20) { mover(0, e.deltaY > 0 ? 1 : -1); } }, { passive: true });
document.addEventListener('mousemove', e => {
  if (!mouseMexeu(e)) return;
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
    Som.ligado = s.sons !== false;
    if (typeof s.protetorMin === 'number') protetorMin = s.protetorMin;
    mostrarTimer(s.timer || 0);
    aplicarDesempenho(s.desempenho);
    if (s.idioma && window.tvcIdioma) window.tvcIdioma(s.idioma);
    const st = $('#stats');
    if (s.stats) st.textContent = `CPU ${s.stats.cpu}% · Memória livre ${(s.stats.memLivreMB / 1024).toFixed(1)} de ${(s.stats.memTotalMB / 1024).toFixed(1)} GB · Disco livre ${Math.round(s.stats.discoLivreGB)} GB`;
    mostrar('#stats', !!s.stats);
    const cl = $('#clima');
    if (s.clima) { cl.textContent = `${s.clima.icone} ${Math.round(s.clima.temp)}°`; cl.title = s.clima.texto; }
    mostrar('#clima', !!s.clima);
    $('#cel-chip').textContent = '📱 ' + (s.conectados || 0);
    mostrar('#cel-chip', s.conectados > 0);
    videosLigados = s.videosFundo !== false;
    const novos = (s.videos || []).join(',');
    if (novos !== estado.videos) {
      estado.videos = novos; videosFundo = new Set(s.videos || []);
      tocarVideoFundo(arteAtual || 'fundo'); // vídeo novo na pasta já aparece
    }
    if (s.novidade && !sessionStorage.getItem('aviso-' + s.versao)) { sessionStorage.setItem('aviso-' + s.versao, '1'); aviso(s.novidade, 8000); }
    if (s.atualizacao && s.atualizacao !== estado.ultimoStatus) aviso(s.atualizacao);
    estado.ultimoStatus = s.atualizacao;
    // Perfil atual / modo criança
    const crianca = !!s.crianca;
    document.body.classList.toggle('modo-crianca', crianca);
    perfis = s.perfis || []; perfilAtualId = s.perfil && s.perfil.id;
    const chip = $('#perfil-chip');
    chip.classList.toggle('crianca', crianca);
    if (crianca) chip.textContent = '🧸 Modo criança';
    else if (s.perfil && perfis.length > 1) {
      const av = document.createElement('span'); av.className = 'av'; av.style.background = s.perfil.cor; av.textContent = s.perfil.emoji;
      chip.replaceChildren(av, document.createTextNode(s.perfil.nome));
    }
    mostrar('#perfil-chip', crianca || perfis.length > 1);
    $('#sistema [data-acao="perfil"]').classList.toggle('oculto', crianca || perfis.length < 2);
    const dica = $('#rodape').lastElementChild;
    dica.firstElementChild.style.display = crianca ? 'none' : '';
    dica.lastChild.textContent = crianca ? 'Para sair do modo criança, use o celular' : ' no celular volta aqui';
    montarDepois(s.depois || []);
    totalFotos = s.fotos || 0;
    $('#sistema [data-acao="fotos"]').classList.toggle('oculto', !totalFotos);
    $('#fotos-txt').textContent = totalFotos ? `${totalFotos} foto${totalFotos > 1 ? 's' : ''}` : '';
    $('#sistema [data-acao="espelhar"]').classList.toggle('oculto', crianca);
    if (s.fotoAgora && s.fotoAgora.quando !== estado.fotoQuando) {
      const primeira = estado.fotoQuando === undefined; estado.fotoQuando = s.fotoAgora.quando;
      if (!primeira) abrirVisor(s.fotoAgora.nome);
    } else if (!s.fotoAgora && estado.fotoQuando === undefined) estado.fotoQuando = null;
    const sisAntes = linhas[2].length; linhas[2] = sistemaVisivel();
    if (pos.linha === 2 && pos.col >= linhas[2].length) { pos.col = Math.max(0, linhas[2].length - 1); focar(); }
    void sisAntes;
    const j = JSON.stringify(s.apps) + crianca + perfis.length;
    if (j !== appsJson) { appsJson = j; montarApps(s.apps); }
    // "Quem está assistindo?" ao ligar (depois da abertura)
    if (!estado.perguntou && s.perguntarPerfil && perfis.length > 1 && !crianca && !sessionStorage.getItem('perfil-perguntado')) {
      estado.perguntou = true;
      sessionStorage.setItem('perfil-perguntado', '1');
      (function esperarIntro() { document.getElementById('intro') ? setTimeout(esperarIntro, 300) : abrirPerfis(); })();
    }
    const end = (s.enderecos || [])[0] || '';
    if (end !== estado.end) { estado.end = end; $('#qr img').src = 'qr.png?e=' + encodeURIComponent(end); }
    $('#endereco').textContent = end || 'sem rede: conecte o Wi-Fi';
    if (s.celulares === 0 && !estado.mostrouQr) { $('#qr').classList.remove('oculto'); estado.mostrouQr = true; }
    $('#codigo').textContent = s.codigo || '';
    mostrar('#pareamento', !!s.codigo);
    if (s.codigo) $('#qr').classList.add('oculto');
  } catch (_) { /* servidor reiniciando */ }
}

// ---------- Desempenho: "auto" mede os quadros por segundo e liga o modo leve se o PC não der conta ----------
function aplicarDesempenho(modo) {
  let leve = modo === 'leve';
  if (modo === 'auto') {
    try { leve = localStorage.getItem('pc-fraco') === '1'; } catch (_) {}
    if (!aplicarDesempenho.medido && !document.getElementById('intro')) { aplicarDesempenho.medido = true; medirFPS(); }
  }
  if (document.body.classList.contains('leve') !== leve) {
    document.body.classList.toggle('leve', leve);
    if (leve) pararVideo(videoFundo()); else tocarVideoFundo(arteAtual || 'fundo');
  }
}
function medirFPS() {
  // Mede durante uma troca de cartão (o momento mais pesado da tela)
  let quadros = 0, inicio = 0;
  const conta = t => {
    if (!inicio) inicio = t;
    quadros++;
    if (t - inicio < 2000) requestAnimationFrame(conta);
    else {
      const fps = quadros * 1000 / (t - inicio);
      try { localStorage.setItem('pc-fraco', fps < 40 ? '1' : '0'); } catch (_) {}
      console.log('TV Central: ' + fps.toFixed(0) + ' quadros por segundo');
    }
  };
  setTimeout(() => requestAnimationFrame(conta), 1500);
}

// ---------- Assistir depois ----------
let depoisJson = '';
function montarDepois(itens) {
  const j = JSON.stringify(itens);
  if (j === depoisJson) return;
  depoisJson = j;
  const caixa = $('#depois');
  const idFoco = (linhas[pos.linha][pos.col] || {}).dataset;
  caixa.replaceChildren(...itens.map(it => {
    const b = document.createElement('button');
    b.className = 'cartao';
    b.dataset.depois = it.id; b.dataset.titulo = it.titulo; b.dataset.app = it.app || '';
    b.style.setProperty('--capa', `url("/capa/${it.app || 'padrao'}?p=1")`);
    b.style.setProperty('--c1', '#3a3f4d');
    const n = document.createElement('span'); n.className = 'nome'; n.textContent = it.titulo;
    const de = document.createElement('span'); de.className = 'de'; de.textContent = new URL(it.url).hostname.replace(/^www\./, '');
    b.append(n, de);
    return b;
  }));
  linhas[1] = [...caixa.children];
  mostrar('#linha-depois', itens.length > 0);
  if (pos.linha === 1) { pos.col = Math.min(pos.col, linhas[1].length - 1); if (pos.col < 0) pos = { linha: 0, col: 0 }; }
  void idFoco;
  focar();
}

// ---------- Visor de fotos (pelo botão Fotos ou quando o celular manda mostrar) ----------
let totalFotos = 0, fotosVisor = [], idxFoto = 0, vezFoto = 0;
async function abrirVisor(inicio) {
  try { fotosVisor = (await (await fetch('/api/fotos', { cache: 'no-store' })).json()).fotos || []; } catch (_) { fotosVisor = []; }
  if (!fotosVisor.length) { aviso('Nenhuma foto: mande pelo celular em Apps → Fotos'); return; }
  idxFoto = typeof inicio === 'string' ? Math.max(0, fotosVisor.indexOf(inicio)) : 0;
  modal = 'visor'; mostrar('#visor', true);
  if (protetorAtivo) desligarProtetor();
  pintarFoto();
  clearInterval(abrirVisor.t);
  abrirVisor.t = setInterval(() => passarFoto(1, true), 8000);
}
function pintarFoto() {
  const [a, b] = [$('#visor-a'), $('#visor-b')];
  const nova = vezFoto ? a : b, velha = vezFoto ? b : a; vezFoto = 1 - vezFoto;
  nova.onload = () => { nova.classList.add('ativa'); velha.classList.remove('ativa'); };
  nova.src = '/foto/' + encodeURIComponent(fotosVisor[idxFoto]);
  $('#visor-pos').textContent = `${idxFoto + 1} de ${fotosVisor.length}`;
}
function passarFoto(d, auto) {
  if (!fotosVisor.length) return;
  idxFoto = (idxFoto + d + fotosVisor.length) % fotosVisor.length;
  if (!auto) { Som.mover(); clearInterval(abrirVisor.t); abrirVisor.t = setInterval(() => passarFoto(1, true), 8000); }
  pintarFoto();
}
function fecharVisor() { clearInterval(abrirVisor.t); modal = null; mostrar('#visor', false); Som.voltar(); usou(); }

// ---------- Sons de navegação (curtinhos e baixos; "sons": false no config.json desliga) ----------
const Som = {
  ctx: null, ligado: true,
  tocar(freqs, dur = 0.05, vol = 0.035, tipo = 'sine') {
    if (!this.ligado) return;
    try {
      this.ctx = this.ctx || new (window.AudioContext || window.webkitAudioContext)();
      const t = this.ctx.currentTime;
      freqs.forEach((f, i) => {
        const o = this.ctx.createOscillator(), g = this.ctx.createGain();
        o.type = tipo; o.frequency.value = f;
        const t0 = t + i * dur * 0.8;
        g.gain.setValueAtTime(0.0001, t0);
        g.gain.exponentialRampToValueAtTime(vol, t0 + 0.006);
        g.gain.exponentialRampToValueAtTime(0.0001, t0 + dur);
        o.connect(g).connect(this.ctx.destination); o.start(t0); o.stop(t0 + dur + 0.02);
      });
    } catch (_) {}
  },
  mover() { this.tocar([1320], 0.04, 0.025); },
  limite() { this.tocar([180], 0.07, 0.04, 'triangle'); },
  ok() { this.tocar([880, 1320], 0.07, 0.035); },
  voltar() { this.tocar([700, 470], 0.06, 0.03); },
};

// ---------- Desligar em… (timer) ----------
$('#prot-video').addEventListener('playing', e => e.target.classList.add('ativa'));
$('#prot-video').addEventListener('error', e => pararVideo(e.target));

const OPCOES_TIMER = [0, 30, 60, 90, 120];
let timerRestante = 0;
function mostrarTimer(seg) {
  timerRestante = seg;
  const min = Math.ceil(seg / 60);
  $('#timer-txt').textContent = seg ? `desliga em ${min} min` : 'desativado';
  $('#timer-chip').textContent = seg ? `☾ ${min} min` : '';
  mostrar('#timer-chip', !!seg);
  if (seg && seg <= 60 && !mostrarTimer.avisou) { mostrarTimer.avisou = true; aviso('O computador vai desligar em 1 minuto'); }
  if (!seg || seg > 60) mostrarTimer.avisou = false;
}
async function proximoTimer() {
  const atual = Math.ceil(timerRestante / 60);
  const prox = OPCOES_TIMER.find(m => m > atual + 1) || 0;
  try {
    const r = await fetch('api/timer', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ min: prox }) });
    mostrarTimer((await r.json()).timer);
    aviso(prox ? `Desliga sozinho em ${prox} minutos` : 'Desligamento programado cancelado', 3000);
  } catch (_) { aviso('Sem conexão com o programa'); }
}

// ---------- Protetor de tela ----------
let protetorMin = 10, ultimoUso = Date.now(), protetorAtivo = false, protetorVez = 0, protetorIdx = 0;
function usou() { ultimoUso = Date.now(); }
function ligarProtetor() {
  protetorAtivo = true;
  pararVideo(videoFundo());
  mostrar('#protetor', true);
  trocarProtetor();
  ligarProtetor.t = setInterval(trocarProtetor, 15000);
}
function desligarProtetor() {
  protetorAtivo = false;
  clearInterval(ligarProtetor.t);
  pararVideo($('#prot-video'));
  mostrar('#protetor', false);
  tocarVideoFundo(arteAtual || 'fundo');
  usou();
}
async function trocarProtetor() {
  if (totalFotos && (!trocarProtetor.fotos || protetorIdx % 20 === 0)) {
    try { trocarProtetor.fotos = (await (await fetch('/api/fotos', { cache: 'no-store' })).json()).fotos || []; } catch (_) {}
  }
  const fotos = totalFotos ? (trocarProtetor.fotos || []) : [];
  if (fotos.length) { // com fotos, o protetor vira um porta-retratos
    const nome = fotos[protetorIdx++ % fotos.length];
    pararVideo($('#prot-video'));
    const camadas = [$('#prot-a'), $('#prot-b')];
    const nova = camadas[protetorVez = 1 - protetorVez];
    nova.style.backgroundImage = `url("/foto/${encodeURIComponent(nome)}")`;
    nova.style.backgroundSize = 'contain'; nova.style.backgroundPosition = 'center'; nova.style.backgroundColor = '#000';
    nova.classList.add('ativa'); camadas[1 - protetorVez].classList.remove('ativa');
    $('#prot-relogio').textContent = new Date().toLocaleTimeString(window.TVC_LOCALE || 'pt-BR', { hour: '2-digit', minute: '2-digit' });
    $('#prot-data').textContent = $('#data').textContent;
    return;
  }
  const ids = linhas[0].map(b => b.dataset.id);
  if (!ids.length) return;
  const id = ids[protetorIdx++ % ids.length];
  [$('#prot-a'), $('#prot-b')].forEach(c => { c.style.backgroundSize = ''; c.style.backgroundPosition = ''; });
  const pv = $('#prot-video');
  if (videosLigados && videosFundo.has(id)) {
    pv.classList.remove('ativa'); pv.src = `/video/${id}`; pv.play().catch(() => {});
  } else pararVideo(pv);
  const camadas = [$('#prot-a'), $('#prot-b')];
  const nova = camadas[protetorVez = 1 - protetorVez];
  nova.style.backgroundImage = `url("/capa/${id}")`;
  nova.classList.add('ativa'); camadas[1 - protetorVez].classList.remove('ativa');
  const d = new Date();
  $('#prot-relogio').textContent = d.toLocaleTimeString(window.TVC_LOCALE || 'pt-BR', { hour: '2-digit', minute: '2-digit' });
  $('#prot-data').textContent = $('#data').textContent;
}
// Qualquer tecla/mouse com o protetor ligado só desliga ele (não navega)
window.addEventListener('keydown', e => { if (protetorAtivo) { e.preventDefault(); e.stopImmediatePropagation(); desligarProtetor(); } else usou(); }, true);
window.addEventListener('mousemove', e => { if (Math.abs(e.movementX) + Math.abs(e.movementY) < 4) return; if (protetorAtivo) desligarProtetor(); else usou(); }, true);
window.addEventListener('focus', () => { if (protetorAtivo) desligarProtetor(); usou(); });
setInterval(() => {
  if (protetorAtivo) { $('#prot-relogio').textContent = new Date().toLocaleTimeString(window.TVC_LOCALE || 'pt-BR', { hour: '2-digit', minute: '2-digit' }); return; }
  if (protetorMin > 0 && !modal && Date.now() - ultimoUso > protetorMin * 60000 && document.hasFocus()) ligarProtetor();
}, 5000);

relogio();
setInterval(relogio, 5000);
linhas[2] = sistemaVisivel();
// Com a tela inicial na frente: atualiza a cada 1,5 s. Com um app aberto: a cada 6 s (economiza).
(function laco() {
  estado();
  setTimeout(laco, document.hasFocus() && !document.hidden ? 1500 : 6000);
})();
window.addEventListener('focus', () => mostrar('#abrindo', false));
