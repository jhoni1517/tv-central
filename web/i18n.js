// Idiomas do TV Central (português é o original; inglês e espanhol traduzidos aqui).
// Traduz o texto da tela, inclusive o que aparece depois (avisos, listas), sem mexer no resto do código.
(() => {
  // [português, inglês, espanhol]
  const T = [
    // ---------- TV ----------
    ['Bom dia', 'Good morning', 'Buenos días'], ['Boa tarde', 'Good afternoon', 'Buenas tardes'], ['Boa noite', 'Good evening', 'Buenas noches'],
    ['Escolha o que assistir', 'Choose what to watch', 'Elige qué ver'], ['Pressione OK para abrir', 'Press OK to open', 'Pulsa OK para abrir'],
    ['para abrir', 'to open', 'para abrir'], ['Assistir', 'Watch', 'Ver'], ['Assistir depois', 'Watch later', 'Ver más tarde'], ['Sistema', 'System', 'Sistema'],
    ['Trocar perfil', 'Switch profile', 'Cambiar perfil'], ['Fotos', 'Photos', 'Fotos'], ['Espelhar celular', 'Mirror phone', 'Duplicar móvil'],
    ['Conectar celular', 'Connect phone', 'Conectar móvil'], ['Reiniciar', 'Restart', 'Reiniciar'], ['Desligar', 'Shut down', 'Apagar'],
    ['Timer', 'Sleep timer', 'Temporizador'], ['desativado', 'off', 'desactivado'], ['Área de trabalho', 'Desktop', 'Escritorio'], ['Atualizar', 'Update', 'Actualizar'],
    ['Controle pelo celular', 'Phone remote', 'Control desde el móvil'], ['Na mesma Wi-Fi, aponte a câmera do celular ou abra:', 'On the same Wi-Fi, point your phone camera here or open:', 'En la misma Wi-Fi, apunta la cámara del móvil o abre:'],
    ['navegar', 'navigate', 'navegar'], ['abrir', 'open', 'abrir'], ['no celular volta aqui', 'on the phone comes back here', 'en el móvil vuelve aquí'], ['Início', 'Home', 'Inicio'],
    ['Digite no celular o código', 'Type this code on your phone', 'Escribe este código en el móvil'], ['Vale por 3 minutos', 'Valid for 3 minutes', 'Válido por 3 minutos'],
    ['Desligar o computador?', 'Shut down the computer?', '¿Apagar el ordenador?'], ['Reiniciar o computador?', 'Restart the computer?', '¿Reiniciar el ordenador?'],
    ['Sim', 'Yes', 'Sí'], ['Cancelar', 'Cancel', 'Cancelar'], ['Quem está assistindo?', "Who's watching?", '¿Quién está viendo?'],
    ['Abrindo…', 'Opening…', 'Abriendo…'], ['passar', 'next', 'pasar'], ['fechar', 'close', 'cerrar'], ['Voltar', 'Back', 'Atrás'],
    ['📲 Espelhar o celular', '📲 Mirror your phone', '📲 Duplicar el móvil'],
    ['No celular, abra o painel rápido (deslize de cima para baixo)', 'On the phone, open quick settings (swipe down from the top)', 'En el móvil, abre los ajustes rápidos (desliza hacia abajo)'],
    ['Toque em', 'Tap', 'Toca'], ['ou', 'or', 'o'], ['Transmitir', 'Cast', 'Enviar'], ['Espelhamento de tela', 'Screen mirroring', 'Duplicar pantalla'],
    ['Escolha este computador na lista', 'Pick this computer in the list', 'Elige este ordenador en la lista'],
    ['A janela do Windows "Conectar" abre por cima. Aperte Início no celular para voltar.', 'The Windows "Connect" window opens on top. Press Home on the phone to come back.', 'La ventana "Conectar" de Windows se abre encima. Pulsa Inicio en el móvil para volver.'],
    ['Sua TV, do seu jeito', 'Your TV, your way', 'Tu TV, a tu manera'],
    ['Para sair do modo criança, use o celular', 'To leave kids mode, use the phone', 'Para salir del modo infantil, usa el móvil'],
    ['🧸 Modo criança', '🧸 Kids mode', '🧸 Modo infantil'],
    ['Nenhuma foto: mande pelo celular em Apps → Fotos', 'No photos yet: send them from the phone in Apps → Photos', 'Sin fotos: envíalas desde el móvil en Apps → Fotos'],
    ['O computador vai desligar em 1 minuto', 'The computer will shut down in 1 minute', 'El ordenador se apagará en 1 minuto'],
    ['Desligamento programado cancelado', 'Sleep timer cancelled', 'Apagado programado cancelado'],
    ['Indo para a área de trabalho. Aperte Início no celular para voltar.', 'Going to the desktop. Press Home on the phone to come back.', 'Yendo al escritorio. Pulsa Inicio en el móvil para volver.'],
    ['Procurando versão nova…', 'Looking for a new version…', 'Buscando una versión nueva…'], ['Pronto', 'Done', 'Listo'], ['Sem conexão com o programa', 'No connection to the program', 'Sin conexión con el programa'],
    ['Instalando a versão nova…', 'Installing the new version…', 'Instalando la versión nueva…'],
    // descrições dos apps
    ['Séries, filmes e documentários', 'Series, movies and documentaries', 'Series, películas y documentales'],
    ['Séries e filmes da HBO e da Warner', 'HBO and Warner series and movies', 'Series y películas de HBO y Warner'],
    ['Disney, Pixar, Marvel, Star Wars e National Geographic', 'Disney, Pixar, Marvel, Star Wars and National Geographic', 'Disney, Pixar, Marvel, Star Wars y National Geographic'],
    ['Séries e filmes originais da Apple', 'Apple original series and movies', 'Series y películas originales de Apple'],
    ['Vídeos, música e canais, com tela de TV', 'Videos, music and channels, with the TV interface', 'Vídeos, música y canales, con interfaz de TV'],
    ['Vídeos para crianças', 'Videos for kids', 'Vídeos para niños'], ['Filmes, séries e esportes da Amazon', "Amazon's movies, series and sports", 'Películas, series y deportes de Amazon'],
    ['Novelas, séries e a Globo ao vivo', 'Soaps, series and Globo live', 'Telenovelas, series y Globo en vivo'],
    // ---------- celular ----------
    ['Conectar à TV', 'Connect to the TV', 'Conectar a la TV'], ['Toque no botão: um código vai aparecer na TV.', 'Tap the button: a code will show up on the TV.', 'Toca el botón: aparecerá un código en la TV.'],
    ['Mostrar código na TV', 'Show code on the TV', 'Mostrar código en la TV'], ['Conectar', 'Connect', 'Conectar'], ['Reconectando ao PC…', 'Reconnecting to the PC…', 'Reconectando al PC…'],
    ['Digitar na TV…', 'Type on the TV…', 'Escribir en la TV…'], ['Controle', 'Remote', 'Control'], ['Gestos', 'Gestures', 'Gestos'], ['Touchpad', 'Touchpad', 'Touchpad'], ['Apps', 'Apps', 'Apps'],
    ['Mudo', 'Mute', 'Silencio'], ['Play', 'Play', 'Play'], ['Vol', 'Vol', 'Vol'], ['Deslize', 'Swipe', 'Desliza'], ['para andar', 'to move', 'para moverte'], ['Toque', 'Tap', 'Toca'], ['Segure', 'Hold', 'Mantén'],
    ['Arraste para mover · toque para clicar · dois dedos para rolar', 'Drag to move · tap to click · two fingers to scroll', 'Arrastra para mover · toca para hacer clic · dos dedos para desplazar'],
    ['Clique', 'Click', 'Clic'], ['Botão direito', 'Right click', 'Clic derecho'], ['Desligar em', 'Shut down in', 'Apagar en'], ['Reiniciar PC', 'Restart PC', 'Reiniciar PC'], ['Desligar PC', 'Shut down PC', 'Apagar PC'],
    ['Diagnóstico', 'Diagnostics', 'Diagnóstico'], ['Carregando…', 'Loading…', 'Cargando…'], ['Ajustes', 'Settings', 'Ajustes'],
    ['Cole o link aqui', 'Paste the link here', 'Pega el enlace aquí'], ['Salvar', 'Save', 'Guardar'], ['Fotos na TV', 'Photos on the TV', 'Fotos en la TV'], ['Enviar fotos', 'Send photos', 'Enviar fotos'],
    ['As fotos viram o protetor de tela. Toque numa foto para mostrar na TV agora.', 'Photos become the screensaver. Tap a photo to show it on the TV now.', 'Las fotos se vuelven el salvapantallas. Toca una foto para mostrarla ahora en la TV.'],
    ['Cole o link de um filme, série ou vídeo. Dica: no app da Netflix/YouTube, toque em Compartilhar →', 'Paste the link of a movie, show or video. Tip: in the Netflix/YouTube app, tap Share →', 'Pega el enlace de una película, serie o vídeo. Consejo: en la app de Netflix/YouTube, toca Compartir →'],
    ['nenhuma ainda', 'none yet', 'ninguna todavía'], ['Mostrando na TV', 'Showing on the TV', 'Mostrando en la TV'], ['Apagar esta foto da TV?', 'Delete this photo from the TV?', '¿Borrar esta foto de la TV?'],
    ['Salvo em Assistir depois ✓', 'Saved to Watch later ✓', 'Guardado en Ver más tarde ✓'], ['Sem conexão', 'No connection', 'Sin conexión'], ['Não enviou', 'Upload failed', 'No se envió'],
    ['Quem está assistindo', "Who's watching", 'Quién está viendo'], ['Cada perfil tem seus apps e os próprios logins nos streamings. Toque no perfil atual para editar.', 'Each profile has its own apps and its own streaming logins. Tap the current profile to edit it.', 'Cada perfil tiene sus apps y sus propios inicios de sesión. Toca el perfil actual para editarlo.'],
    ['Perguntar ao ligar a TV', 'Ask when the TV turns on', 'Preguntar al encender la TV'], ['Novo', 'New', 'Nuevo'], ['Novo perfil', 'New profile', 'Nuevo perfil'], ['Editar perfil', 'Edit profile', 'Editar perfil'],
    ['Nome', 'Name', 'Nombre'], ['Apps que este perfil vê', 'Apps this profile sees', 'Apps que ve este perfil'], ['Apagar perfil', 'Delete profile', 'Borrar perfil'], ['Perfil salvo', 'Profile saved', 'Perfil guardado'], ['Perfil apagado', 'Profile deleted', 'Perfil borrado'],
    ['Apps da tela inicial', 'Home screen apps', 'Apps de la pantalla de inicio'], ['Mude a ordem, esconda apps e coloque sua imagem ou vídeo de fundo em cada um.', 'Reorder, hide apps and set your own image or background video for each one.', 'Cambia el orden, oculta apps y pon tu imagen o vídeo de fondo en cada una.'],
    ['arte padrão', 'default art', 'arte predeterminada'], ['imagem sua', 'your image', 'tu imagen'], ['Pôr imagem', 'Add image', 'Poner imagen'], ['Trocar imagem', 'Change image', 'Cambiar imagen'], ['Tirar imagem', 'Remove image', 'Quitar imagen'],
    ['Pôr vídeo', 'Add video', 'Poner vídeo'], ['Trocar vídeo', 'Change video', 'Cambiar vídeo'], ['Tirar vídeo', 'Remove video', 'Quitar vídeo'], ['Ordem salva', 'Order saved', 'Orden guardado'], ['App visível', 'App shown', 'App visible'], ['App escondido', 'App hidden', 'App oculta'],
    ['Imagem enviada', 'Image uploaded', 'Imagen enviada'], ['Vídeo enviado', 'Video uploaded', 'Vídeo enviado'], ['Removido', 'Removed', 'Quitado'],
    ['Adicionar um site', 'Add a website', 'Añadir un sitio'], ['Qualquer site vira um cartão na tela inicial (Crunchyroll, Twitch, Pluto TV, um canal…). Abre em tela cheia, com setas e bloqueio de anúncios.', 'Any website becomes a card on the home screen (Crunchyroll, Twitch, Pluto TV, a channel…). It opens full screen, with arrow navigation and ad blocking.', 'Cualquier sitio se vuelve una tarjeta en la pantalla de inicio (Crunchyroll, Twitch, Pluto TV, un canal…). Se abre a pantalla completa, con flechas y bloqueo de anuncios.'],
    ['Nome (ex.: Crunchyroll)', 'Name (e.g. Crunchyroll)', 'Nombre (ej.: Crunchyroll)'], ['Endereço (ex.: crunchyroll.com)', 'Address (e.g. crunchyroll.com)', 'Dirección (ej.: crunchyroll.com)'], ['Adicionar', 'Add', 'Añadir'], ['Remover', 'Remove', 'Quitar'],
    ['Site adicionado', 'Website added', 'Sitio añadido'], ['Site removido', 'Website removed', 'Sitio quitado'],
    ['Tela inicial', 'Home screen', 'Pantalla de inicio'], ['Clima', 'Weather', 'Clima'], ['Cidade para mostrar a temperatura', 'City for the temperature', 'Ciudad para la temperatura'], ['Ex.: Curitiba', 'E.g. London', 'Ej.: Madrid'],
    ['Cidade salva', 'City saved', 'Ciudad guardada'], ['Clima desligado', 'Weather off', 'Clima desactivado'],
    ['Bloquear anúncios', 'Block ads', 'Bloquear anuncios'], ['Anúncios e rastreadores nos sites, e anúncios do YouTube', 'Ads and trackers on websites, and YouTube ads', 'Anuncios y rastreadores en los sitios, y anuncios de YouTube'],
    ['Bloqueio ligado: vale ao abrir o próximo app', 'Blocking on: applies from the next app you open', 'Bloqueo activado: vale al abrir la próxima app'], ['Bloqueio desligado', 'Blocking off', 'Bloqueo desactivado'],
    ['Sons ao navegar', 'Navigation sounds', 'Sonidos al navegar'], ['Abertura com música ao ligar', 'Intro with music at start', 'Intro con música al encender'], ['Volume da abertura', 'Intro volume', 'Volumen de la intro'], ['Volume da abertura salvo', 'Intro volume saved', 'Volumen de la intro guardado'],
    ['Protetor de tela', 'Screensaver', 'Salvapantallas'], ['Desligado', 'Off', 'Desactivado'], ['Vídeos de fundo', 'Background videos', 'Vídeos de fondo'], ['Vídeo de fundo geral', 'General background video', 'Vídeo de fondo general'],
    ['Toca quando nenhum app está selecionado', 'Plays when no app is selected', 'Se reproduce cuando no hay app seleccionada'], ['Configurado', 'Set', 'Configurado'], ['Escolher', 'Choose', 'Elegir'], ['Trocar', 'Change', 'Cambiar'], ['Tirar', 'Remove', 'Quitar'],
    ['Desempenho', 'Performance', 'Rendimiento'], ['Leve tira animações e vídeos (PC fraco). Automático decide sozinho.', 'Light turns off animations and videos (slow PC). Automatic decides by itself.', 'Ligero quita animaciones y vídeos (PC lento). Automático decide solo.'],
    ['Automático', 'Automatic', 'Automático'], ['Leve', 'Light', 'Ligero'], ['Completo', 'Full', 'Completo'], ['Mostrar CPU, memória e disco', 'Show CPU, memory and disk', 'Mostrar CPU, memoria y disco'], ['Uma linha embaixo do relógio', 'One line under the clock', 'Una línea bajo el reloj'],
    ['Idioma', 'Language', 'Idioma'], ['Salvo', 'Saved', 'Guardado'],
    ['Modo criança', 'Kids mode', 'Modo infantil'], ['LIGADO', 'ON', 'ACTIVADO'], ['Ligar modo criança', 'Turn on kids mode', 'Activar modo infantil'], ['Modo criança ligado 🧸', 'Kids mode on 🧸', 'Modo infantil activado 🧸'], ['Modo criança desligado', 'Kids mode off', 'Modo infantil desactivado'],
    ['Mostra só os apps escolhidos, com logins separados, e esconde desligar, reiniciar e ajustes. Para sair, só com o PIN.', 'Shows only the chosen apps, with separate logins, and hides shut down, restart and settings. Leaving needs the PIN.', 'Muestra solo las apps elegidas, con sesiones separadas, y oculta apagar, reiniciar y ajustes. Para salir, solo con el PIN.'],
    ['A TV só mostra os apps liberados e nada pode ser mudado. Digite o PIN para sair.', 'The TV only shows the allowed apps and nothing can be changed. Enter the PIN to leave.', 'La TV solo muestra las apps permitidas y no se puede cambiar nada. Escribe el PIN para salir.'],
    ['Crie um PIN (4 a 8 números)', 'Create a PIN (4 to 8 digits)', 'Crea un PIN (4 a 8 números)'], ['PIN', 'PIN', 'PIN'], ['Sair', 'Leave', 'Salir'], ['Trocar PIN', 'Change PIN', 'Cambiar PIN'], ['PIN atual', 'Current PIN', 'PIN actual'], ['PIN novo', 'New PIN', 'PIN nuevo'], ['PIN trocado', 'PIN changed', 'PIN cambiado'],
    // mensagens do programa
    ['código errado', 'wrong code', 'código incorrecto'], ['código expirado: peça outro', 'code expired: ask for another', 'código vencido: pide otro'], ['muitas tentativas: peça outro código', 'too many tries: ask for another code', 'demasiados intentos: pide otro código'],
    ['muitas tentativas: espere 1 minuto', 'too many tries: wait 1 minute', 'demasiados intentos: espera 1 minuto'], ['celular não pareado', 'phone not paired', 'móvil no emparejado'], ['PIN errado', 'wrong PIN', 'PIN incorrecto'],
    ['modo criança ativo: digite o PIN para sair antes de mudar algo', 'kids mode is on: enter the PIN to leave before changing anything', 'modo infantil activo: escribe el PIN para salir antes de cambiar algo'],
    ['endereço inválido (ex.: crunchyroll.com)', 'invalid address (e.g. crunchyroll.com)', 'dirección no válida (ej.: crunchyroll.com)'], ['coloque o endereço do site', 'enter the website address', 'escribe la dirección del sitio'],
    ['não achei um link aí: copie o link do filme, série ou vídeo', "couldn't find a link there: copy the link of the movie, show or video", 'no encontré un enlace: copia el enlace de la película, serie o vídeo'],
    ['esse link já está na lista', 'that link is already in the list', 'ese enlace ya está en la lista'], ['crie um PIN de 4 a 8 números', 'create a PIN with 4 to 8 digits', 'crea un PIN de 4 a 8 números'],
    ['dê um nome ao perfil', 'give the profile a name', 'ponle un nombre al perfil'], ['escolha pelo menos um app para o modo criança', 'pick at least one app for kids mode', 'elige al menos una app para el modo infantil'],
  ];
  const P = [ // textos com partes que mudam: [regex em português, inglês, espanhol]
    [/^Abrindo (.+)…$/, 'Opening $1…', 'Abriendo $1…'], [/^Não abriu: (.+)$/, "Didn't open: $1", 'No se abrió: $1'], [/^Não deu: (.+)$/, "Couldn't: $1", 'No se pudo: $1'],
    [/^Assistir depois · (.+)$/, 'Watch later · $1', 'Ver más tarde · $1'], [/^desliga em (\d+) min$/, 'off in $1 min', 'apaga en $1 min'], [/^Desliga sozinho em (\d+) minutos$/, 'Shuts down by itself in $1 minutes', 'Se apaga solo en $1 minutos'],
    [/^(\d+) fotos?$/, '$1 photo(s)', '$1 foto(s)'], [/^versão (.+)$/, 'version $1', 'versión $1'], [/^(.+) Olá, (.+)!$/, '$1 Hi, $2!', '$1 ¡Hola, $2!'],
    [/^Atualizado para a versão (.+)$/, 'Updated to version $1', 'Actualizado a la versión $1'], [/^Baixando versão (.+)…$/, 'Downloading version $1…', 'Descargando la versión $1…'],
    [/^Versão (.+) pronta: instala quando nenhum app estiver aberto$/, 'Version $1 ready: installs when no app is open', 'Versión $1 lista: se instala cuando no haya apps abiertas'],
    [/^Você já tem a versão mais nova \((.+)\)$/, 'You already have the latest version ($1)', 'Ya tienes la versión más nueva ($1)'], [/^Não deu para verificar: (.+)$/, "Couldn't check: $1", 'No se pudo comprobar: $1'],
    [/^(\d+) na TV$/, '$1 on the TV', '$1 en la TV'], [/^(\d+) foto\(s\) enviada\(s\)(.*)$/, '$1 photo(s) sent$2', '$1 foto(s) enviada(s)$2'], [/^ · (\d+) com problema$/, ' · $1 with problems', ' · $1 con problemas'],
    [/^(\d+) min$/, '$1 min', '$1 min'], [/^(.+) (.+) assistindo$/, '$1 $2 watching', '$1 $2 viendo'], [/^Apagar o perfil (.+)\?$/, 'Delete the profile $1?', '¿Borrar el perfil $1?'],
    [/^Tirar (.+) da tela inicial\?$/, 'Remove $1 from the home screen?', '¿Quitar $1 de la pantalla de inicio?'], [/^Na TV agora: $/, 'On the TV now: ', 'En la TV ahora: '],
    [/^CPU (\d+)% · Memória livre (.+) de (.+) GB · Disco livre (\d+) GB$/, 'CPU $1% · Free memory $2 of $3 GB · Free disk $4 GB', 'CPU $1% · Memoria libre $2 de $3 GB · Disco libre $4 GB'],
  ];

  const LOCALES = { pt: 'pt-BR', en: 'en-US', es: 'es-ES' };
  function escolher() {
    let l = null;
    try { l = localStorage.getItem('idioma'); } catch (_) {}
    if (!l || l === 'auto') l = (navigator.language || 'pt').slice(0, 2).toLowerCase();
    return LOCALES[l] ? l : 'pt';
  }
  const idioma = escolher();
  window.TVC_IDIOMA = idioma;
  window.TVC_LOCALE = LOCALES[idioma];
  document.documentElement.lang = LOCALES[idioma];
  const col = { en: 1, es: 2 }[idioma];
  const mapa = new Map(T.map(l => [l[0], l]));

  function t(txt) {
    if (!col || !txt) return txt;
    const limpo = txt.trim();
    if (!limpo) return txt;
    const l = mapa.get(limpo);
    if (l) return txt.replace(limpo, l[col]);
    for (const p of P) if (p[0].test(limpo)) return txt.replace(limpo, limpo.replace(p[0], p[col]));
    return txt;
  }
  window.tvcT = t;

  // Muda o idioma (salva e recarrega a página)
  window.tvcIdioma = novo => {
    try { if (localStorage.getItem('idioma') === novo) return; localStorage.setItem('idioma', novo); } catch (_) { return; }
    if (escolher() !== idioma) location.reload();
  };

  if (!col) return; // português: nada a traduzir
  const ATRS = ['placeholder', 'aria-label', 'title'];
  function traduzir(raiz) {
    if (raiz.nodeType === 3) { const n = t(raiz.nodeValue); if (n !== raiz.nodeValue) raiz.nodeValue = n; return; }
    if (raiz.nodeType !== 1 || raiz.tagName === 'SCRIPT' || raiz.tagName === 'STYLE' || raiz.isContentEditable) return;
    for (const a of ATRS) if (raiz.hasAttribute && raiz.hasAttribute(a)) { const v = raiz.getAttribute(a), n = t(v); if (n !== v) raiz.setAttribute(a, n); }
    const w = document.createTreeWalker(raiz, NodeFilter.SHOW_TEXT | NodeFilter.SHOW_ELEMENT);
    let n;
    while ((n = w.nextNode())) {
      if (n.nodeType === 3) { const x = t(n.nodeValue); if (x !== n.nodeValue) n.nodeValue = x; }
      else for (const a of ATRS) if (n.hasAttribute(a)) { const v = n.getAttribute(a), x = t(v); if (x !== v) n.setAttribute(a, x); }
    }
  }
  const obs = new MutationObserver(lista => {
    for (const m of lista) {
      if (m.type === 'characterData') traduzir(m.target);
      else if (m.type === 'attributes') traduzir(m.target);
      else m.addedNodes.forEach(traduzir);
    }
  });
  const iniciar = () => {
    traduzir(document.body);
    obs.observe(document.body, { childList: true, subtree: true, characterData: true, attributes: true, attributeFilter: ATRS });
  };
  if (document.body) iniciar(); else document.addEventListener('DOMContentLoaded', iniciar);
  const conf = window.confirm.bind(window);
  window.confirm = m => conf(t(m));
})();
