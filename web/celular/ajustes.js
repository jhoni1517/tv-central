// Painel de ajustes do celular: perfis, apps (ordem, ocultar, imagem e vídeo), tela inicial e modo criança.
(() => {
  const corpo = $('#ajustes-corpo');
  const EMOJIS = ['😎', '🦊', '🐼', '🦁', '🐸', '🦄', '🐱', '🐶', '🌸', '⚡', '🎮', '🎬', '👩', '👨', '👧', '👦', '👵', '👴'];
  let dados = null;

  function toast(txt, ms = 2200) {
    const el = document.createElement('div');
    el.className = 'toast'; el.textContent = txt;
    document.body.append(el);
    setTimeout(() => el.remove(), ms);
  }

  // Cria elementos sem innerHTML com dados (nomes vêm da pessoa)
  function h(tag, props = {}, ...filhos) {
    const el = document.createElement(tag);
    for (const [k, v] of Object.entries(props)) {
      if (k === 'class') el.className = v;
      else if (k.startsWith('on')) el.addEventListener(k.slice(2), v);
      else if (k === 'style') el.style.cssText = v;
      else if (v !== false && v != null) el.setAttribute(k, v === true ? '' : v);
    }
    for (const f of filhos.flat()) if (f != null && f !== false) el.append(f.nodeType ? f : document.createTextNode(f));
    return el;
  }

  function chave(ligada, aoMudar) {
    const inp = h('input', { type: 'checkbox' });
    inp.checked = !!ligada;
    inp.onchange = () => aoMudar(inp.checked);
    return h('label', { class: 'chave' }, inp, h('span'));
  }

  async function salvar(campos, msg = 'Salvo') {
    try { await api('api/ajustes', { method: 'POST', body: JSON.stringify(campos) }); toast(msg); await carregar(); }
    catch (e) { toast(e.message, 3500); }
  }

  async function carregar() {
    try { dados = await api('api/ajustes'); desenhar(); }
    catch (e) { corpo.replaceChildren(h('p', { class: 'erro' }, 'Não deu para carregar: ' + e.message)); }
  }

  // ---------- envio de arquivo com barra de progresso ----------
  function escolherArquivo(aceita, aoEscolher) {
    const inp = h('input', { type: 'file', accept: aceita, style: 'display:none' });
    inp.onchange = () => { if (inp.files[0]) aoEscolher(inp.files[0]); inp.remove(); };
    document.body.append(inp);
    inp.click();
  }
  function enviar(id, tipo, arquivo, barra) {
    return new Promise((ok, falha) => {
      const fd = new FormData();
      fd.append('arquivo', arquivo, arquivo.name);
      const x = new XMLHttpRequest();
      x.open('POST', `api/midia/${id}?tipo=${tipo}`);
      x.setRequestHeader('X-Token', token);
      x.upload.onprogress = e => { if (e.lengthComputable && barra) barra.style.width = (100 * e.loaded / e.total) + '%'; };
      x.onload = () => { let j = {}; try { j = JSON.parse(x.responseText); } catch (_) {} x.status === 200 ? ok(j) : falha(new Error(j.erro || 'falhou')); };
      x.onerror = () => falha(new Error('sem conexão'));
      x.send(fd);
    });
  }
  function botaoEnvio(rotulo, id, tipo) {
    const barra = h('i');
    const prog = h('div', { class: 'progresso', style: 'display:none' }, barra);
    const b = h('button', { class: 'aj-btn', onclick: () => escolherArquivo(tipo === 'video' ? 'video/mp4,video/webm,.mp4,.webm,.m4v' : 'image/*', async f => {
      b.disabled = true; prog.style.display = ''; barra.style.width = '0';
      try { await enviar(id, tipo, f, barra); toast(tipo === 'video' ? 'Vídeo enviado' : 'Imagem enviada'); await carregar(); }
      catch (e) { toast(e.message, 3500); b.disabled = false; prog.style.display = 'none'; }
    }) }, rotulo);
    return [b, prog];
  }
  async function apagarMidia(id, tipo) {
    try { await api(`api/midia/${id}?tipo=${tipo}`, { method: 'DELETE' }); toast('Removido'); await carregar(); }
    catch (e) { toast(e.message, 3500); }
  }

  // ---------- seções ----------
  function secaoPerfis(aj) {
    const lista = h('div', { class: 'perfis' }, aj.perfis.map(p => h('button', {
      class: 'perfil-chip' + (p.id === aj.perfilAtual ? ' atual' : ''),
      onclick: async () => {
        if (p.id === aj.perfilAtual) { editarPerfil(p); return; }
        try { await api('api/perfil/atual', { method: 'POST', body: JSON.stringify({ id: p.id }) }); toast(`${p.emoji} ${p.nome} assistindo`); vibrar(20); await carregar(); }
        catch (e) { toast(e.message, 3500); }
      },
    }, h('span', { class: 'av', style: `background:${p.cor}` }, p.emoji), p.nome)),
    h('button', { class: 'perfil-chip', onclick: () => editarPerfil(null) }, h('span', { class: 'av', style: 'background:#2a2e3a' }, '+'), 'Novo'));
    return h('div', { class: 'cartao-aj' },
      h('h2', {}, 'Quem está assistindo'),
      h('p', { class: 'ajuda' }, 'Cada perfil tem seus apps e os próprios logins nos streamings. Toque no perfil atual para editar.'),
      lista,
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Perguntar ao ligar a TV'), chave(aj.perguntarPerfil, v => salvar({ PerguntarPerfil: v }))));
  }

  function editarPerfil(p) {
    const novo = !p;
    p = p || { nome: '', emoji: EMOJIS[dados.ajustes.perfis.length % EMOJIS.length], ocultos: [] };
    let emoji = p.emoji;
    const nome = h('input', { class: 'aj-input', maxlength: 20, placeholder: 'Nome', value: p.nome });
    const emojis = h('div', { class: 'emojis' }, EMOJIS.map(e => {
      const b = h('button', { class: e === emoji ? 'sel' : '', onclick: () => { emoji = e; emojis.querySelectorAll('button').forEach(x => x.classList.toggle('sel', x === b)); } }, e);
      return b;
    }));
    const apps = h('div', { class: 'marcar' }, dados.apps.map(a => {
      const inp = h('input', { type: 'checkbox', value: a.ID });
      inp.checked = !(p.ocultos || []).includes(a.ID);
      return h('label', {}, inp, a.Nome);
    }));
    const salvarPerfil = async () => {
      const ocultos = [...apps.querySelectorAll('input')].filter(i => !i.checked).map(i => i.value);
      try {
        await api('api/perfis', { method: 'POST', body: JSON.stringify({ acao: novo ? 'criar' : 'editar', perfil: { id: p.id, nome: nome.value, emoji, ocultos } }) });
        toast('Perfil salvo'); await carregar();
      } catch (e) { toast(e.message, 3500); }
    };
    const remover = !novo && p.id !== 'principal' ? h('button', { class: 'aj-btn perigo', onclick: async () => {
      if (!confirm(`Apagar o perfil ${p.nome}?`)) return;
      try { await api('api/perfis', { method: 'POST', body: JSON.stringify({ acao: 'remover', perfil: { id: p.id } }) }); toast('Perfil apagado'); await carregar(); }
      catch (e) { toast(e.message, 3500); }
    } }, 'Apagar perfil') : null;
    corpo.replaceChildren(h('div', { class: 'cartao-aj' },
      h('h2', {}, novo ? 'Novo perfil' : 'Editar perfil'), nome, emojis,
      h('h2', {}, 'Apps que este perfil vê'), apps,
      h('div', { class: 'linha-aj' }, h('button', { class: 'aj-btn', onclick: () => desenhar() }, 'Cancelar'), h('span', { class: 'rot' }), remover, h('button', { class: 'aj-btn primario', onclick: salvarPerfil }, 'Salvar'))));
    nome.focus();
  }

  function secaoApps(aj) {
    const visiveis = new Set(dados.apps.map(a => a.ID).filter(id => !(aj.ocultos || []).includes(id)));
    // ordem atual: a salva primeiro, depois o resto
    const ordem = [...(aj.ordem || []).filter(id => dados.apps.some(a => a.ID === id)), ...dados.apps.map(a => a.ID).filter(id => !(aj.ordem || []).includes(id))];
    const porId = Object.fromEntries(dados.apps.map(a => [a.ID, a]));
    const mover = (i, d) => { const o = [...ordem]; const j = i + d; if (j < 0 || j >= o.length) return; [o[i], o[j]] = [o[j], o[i]]; salvar({ Ordem: o }, 'Ordem salva'); };
    const agora = Date.now();
    return h('div', { class: 'cartao-aj' },
      h('h2', {}, 'Apps da tela inicial'),
      h('p', { class: 'ajuda' }, 'Mude a ordem, esconda apps e coloque sua imagem ou vídeo de fundo em cada um.'),
      ordem.map((id, i) => {
        const a = porId[id];
        return h('div', { class: 'app-aj' },
          h('div', { class: 'mini', style: `background-image:url("/capa/${id}?p=1&v=${agora}")` }),
          h('div', {}, h('div', { class: 'nome-aj' }, a.Nome), h('small', { style: 'color:var(--suave);font-size:11px' }, [a.TemCapa ? 'imagem sua' : 'arte padrão', a.TemVideo ? ' · vídeo' : ''].join(''))),
          h('div', { class: 'linha-aj' },
            h('div', { class: 'setas-aj' }, h('button', { onclick: () => mover(i, -1), disabled: i === 0 }, '↑'), h('button', { onclick: () => mover(i, 1), disabled: i === ordem.length - 1 }, '↓')),
            chave(visiveis.has(id), v => { const oc = new Set(aj.ocultos || []); v ? oc.delete(id) : oc.add(id); salvar({ Ocultos: [...oc] }, v ? 'App visível' : 'App escondido'); })),
          h('div', { class: 'acoes-aj' },
            ...botaoEnvio(a.TemCapa ? 'Trocar imagem' : 'Pôr imagem', id, 'capa'),
            a.TemCapa ? h('button', { class: 'aj-btn perigo', onclick: () => apagarMidia(id, 'capa') }, 'Tirar imagem') : null,
            ...botaoEnvio(a.TemVideo ? 'Trocar vídeo' : 'Pôr vídeo', id, 'video'),
            a.TemVideo ? h('button', { class: 'aj-btn perigo', onclick: () => apagarMidia(id, 'video') }, 'Tirar vídeo') : null));
      }));
  }

  function secaoSites(aj) {
    const nome = h('input', { class: 'aj-input', placeholder: 'Nome (ex.: Crunchyroll)', maxlength: 24 });
    const end = h('input', { class: 'aj-input', placeholder: 'Endereço (ex.: crunchyroll.com)', inputmode: 'url', autocapitalize: 'off' });
    const CORES = ['#e5484d', '#f5a524', '#30a46c', '#12a594', '#3d7bff', '#8e4ec6', '#e93d82', '#5c5c66'];
    let cor = CORES[0];
    const cores = h('div', { class: 'emojis' }, CORES.map(c => {
      const b = h('button', { class: c === cor ? 'sel' : '', style: `background:${c}`, onclick: () => { cor = c; cores.querySelectorAll('button').forEach(x => x.classList.toggle('sel', x === b)); } });
      return b;
    }));
    const extras = dados.apps.filter(a => a.Extra);
    return h('div', { class: 'cartao-aj' },
      h('h2', {}, 'Adicionar um site'),
      h('p', { class: 'ajuda' }, 'Qualquer site vira um cartão na tela inicial (Crunchyroll, Twitch, Pluto TV, um canal…). Abre em tela cheia, com setas e bloqueio de anúncios.'),
      nome, end, cores,
      h('button', { class: 'aj-btn primario', onclick: async () => {
        try { await api('api/sites', { method: 'POST', body: JSON.stringify({ acao: 'criar', nome: nome.value, destino: end.value, cor }) }); toast('Site adicionado'); await carregar(); }
        catch (e) { toast(e.message, 3500); }
      } }, 'Adicionar'),
      extras.length ? h('div', {}, extras.map(a => h('div', { class: 'linha-aj' },
        h('span', { class: 'rot' }, a.Nome, h('small', {}, a.Destino)),
        h('button', { class: 'aj-btn perigo', onclick: async () => {
          if (!confirm(`Tirar ${a.Nome} da tela inicial?`)) return;
          try { await api('api/sites', { method: 'POST', body: JSON.stringify({ acao: 'remover', id: a.ID }) }); toast('Site removido'); await carregar(); }
          catch (e) { toast(e.message, 3500); }
        } }, 'Remover')))) : null);
  }

  function secaoTela(ef) {
    const cidade = h('input', { class: 'aj-input', placeholder: 'Ex.: Curitiba', value: ef.cidade || '', maxlength: 60 });
    const vol = h('input', { type: 'range', min: 0, max: 100, value: Math.round(ef.volumeIntro * 100), style: 'flex:1' });
    vol.onchange = () => salvar({ VolumeIntro: vol.value / 100 }, 'Volume da abertura salvo');
    const prot = h('select', { class: 'aj-input', style: 'flex:0 0 auto' }, [0, 5, 10, 20, 30, 60].map(m => {
      const o = h('option', { value: m }, m ? `${m} min` : 'Desligado'); if (m === ef.protetorMin) o.selected = true; return o;
    }));
    prot.onchange = () => salvar({ ProtetorMin: +prot.value });
    return h('div', { class: 'cartao-aj' },
      h('h2', {}, 'Tela inicial'),
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Clima', h('small', {}, 'Cidade para mostrar a temperatura')),),
      h('div', { class: 'linha-aj' }, cidade, h('button', { class: 'aj-btn', onclick: () => salvar({ Cidade: cidade.value }, cidade.value ? 'Cidade salva' : 'Clima desligado') }, 'Salvar')),
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Bloquear anúncios', h('small', {}, 'Anúncios e rastreadores nos sites, e anúncios do YouTube')), chave(ef.bloquearAnuncios, v => salvar({ BloquearAnuncios: v }, v ? 'Bloqueio ligado: vale ao abrir o próximo app' : 'Bloqueio desligado'))),
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Sons ao navegar'), chave(ef.sons, v => salvar({ Sons: v }))),
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Abertura com música ao ligar'), chave(ef.intro, v => salvar({ Intro: v }))),
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot', style: 'flex:0 0 auto' }, 'Volume da abertura'), vol),
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Protetor de tela'), prot),
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Vídeos de fundo'), chave(ef.videosFundo, v => salvar({ VideosFundo: v }))),
      (() => {
        const d = h('select', { class: 'aj-input', style: 'flex:0 0 auto' }, [['auto', 'Automático'], ['leve', 'Leve'], ['completo', 'Completo']].map(([v, t]) => {
          const o = h('option', { value: v }, t); if (v === ef.desempenho) o.selected = true; return o;
        }));
        d.onchange = () => salvar({ Desempenho: d.value });
        return h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Desempenho', h('small', {}, 'Leve tira animações e vídeos (PC fraco). Automático decide sozinho.')), d);
      })(),
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Mostrar CPU, memória e disco', h('small', {}, 'Uma linha embaixo do relógio')), chave(ef.mostrarStats, v => salvar({ MostrarStats: v }))),
      h('div', { class: 'linha-aj' }, h('span', { class: 'rot' }, 'Vídeo de fundo geral', h('small', {}, dados.fundo ? 'Configurado' : 'Toca quando nenhum app está selecionado')),
        ...botaoEnvio(dados.fundo ? 'Trocar' : 'Escolher', 'fundo', 'video'),
        dados.fundo ? h('button', { class: 'aj-btn perigo', onclick: () => apagarMidia('fundo', 'video') }, 'Tirar') : null));
  }

  function secaoCrianca(aj, temPin) {
    const pin = h('input', { class: 'aj-input', type: 'password', inputmode: 'numeric', maxlength: 8, placeholder: temPin ? 'PIN' : 'Crie um PIN (4 a 8 números)', autocomplete: 'off' });
    if (aj.crianca.ativo) {
      return h('div', { class: 'cartao-aj' },
        h('h2', {}, 'Modo criança ', h('span', { class: 'selo-crianca' }, 'LIGADO')),
        h('p', { class: 'ajuda' }, 'A TV só mostra os apps liberados e nada pode ser mudado. Digite o PIN para sair.'),
        h('div', { class: 'linha-aj' }, pin, h('button', { class: 'aj-btn primario', onclick: async () => {
          try { await api('api/crianca/sair', { method: 'POST', body: JSON.stringify({ pin: pin.value }) }); toast('Modo criança desligado'); vibrar(30); await carregar(); }
          catch (e) { toast(e.message, 3500); vibrar([30, 40, 30]); pin.value = ''; }
        } }, 'Sair')));
    }
    const apps = h('div', { class: 'marcar' }, dados.apps.map(a => {
      const inp = h('input', { type: 'checkbox', value: a.ID });
      inp.checked = aj.crianca.apps.includes(a.ID);
      return h('label', {}, inp, a.Nome);
    }));
    const escolhidos = () => [...apps.querySelectorAll('input')].filter(i => i.checked).map(i => i.value);
    const atualPin = h('input', { class: 'aj-input', type: 'password', inputmode: 'numeric', maxlength: 8, placeholder: 'PIN atual', autocomplete: 'off' });
    const novoPin = h('input', { class: 'aj-input', type: 'password', inputmode: 'numeric', maxlength: 8, placeholder: 'PIN novo', autocomplete: 'off' });
    return h('div', { class: 'cartao-aj' },
      h('h2', {}, 'Modo criança'),
      h('p', { class: 'ajuda' }, 'Mostra só os apps escolhidos, com logins separados, e esconde desligar, reiniciar e ajustes. Para sair, só com o PIN.'),
      apps,
      temPin ? null : pin,
      h('button', { class: 'aj-btn primario', onclick: async () => {
        try { await api('api/crianca/entrar', { method: 'POST', body: JSON.stringify({ pin: pin.value, apps: escolhidos() }) }); toast('Modo criança ligado 🧸'); vibrar(30); await carregar(); }
        catch (e) { toast(e.message, 3500); }
      } }, 'Ligar modo criança'),
      temPin ? h('details', {}, h('summary', { style: 'color:var(--suave);font-size:13px' }, 'Trocar PIN'),
        h('div', { class: 'linha-aj' }, atualPin, novoPin, h('button', { class: 'aj-btn', onclick: async () => {
          try { await api('api/crianca/pin', { method: 'POST', body: JSON.stringify({ atual: atualPin.value, novo: novoPin.value }) }); toast('PIN trocado'); atualPin.value = novoPin.value = ''; }
          catch (e) { toast(e.message, 3500); }
        } }, 'Trocar'))) : null);
  }

  function desenhar() {
    const aj = dados.ajustes;
    if (aj.crianca.ativo) { corpo.replaceChildren(secaoCrianca(aj, dados.ajustes.temPin)); return; }
    corpo.replaceChildren(secaoPerfis(aj), secaoApps(aj), secaoSites(aj), secaoTela(dados.efetivo), secaoCrianca(aj, dados.ajustes.temPin));
  }

  $('#btn-ajustes').onclick = () => { tela('ajustes'); corpo.replaceChildren(h('p', { style: 'color:var(--suave)' }, 'Carregando…')); carregar(); };
  $('#ajustes-voltar').onclick = () => { tela('controle'); carregarApps(); };
})();
