// Abertura ao ligar: logo animado + música curta sintetizada na hora (composição própria).
// Volume contido e com limitador, para não estourar na TV. Qualquer tecla pula.
(() => {
  const el = document.getElementById('intro');
  if (sessionStorage.getItem('intro-vista')) { el.remove(); return; }
  sessionStorage.setItem('intro-vista', '1');
  document.body.classList.add('na-intro');

  let acabou = false;
  function terminar() {
    if (acabou) return;
    acabou = true;
    el.classList.add('saindo');
    document.body.classList.remove('na-intro');
    document.body.classList.add('entrando');
    setTimeout(() => { el.remove(); document.body.classList.remove('entrando'); }, 700);
    if (musica) musica.parar();
  }
  const pular = e => { if (!acabou) { e.preventDefault(); e.stopImmediatePropagation(); terminar(); } };
  window.addEventListener('keydown', pular, true);
  window.addEventListener('click', pular, true);

  let musica = null;
  fetch('api/estado').then(r => r.json()).then(s => {
    if (s.intro === false) { el.remove(); document.body.classList.remove('na-intro'); acabou = true; return; }
    el.classList.add('tocando');
    musica = tocar(typeof s.volumeIntro === 'number' ? s.volumeIntro : 0.5);
    setTimeout(terminar, 4800);
  }).catch(() => terminar());

  function tocar(volume) {
    const AC = window.AudioContext || window.webkitAudioContext;
    if (!AC || volume <= 0) return null;
    const ctx = new AC();
    ctx.resume().catch(() => {});
    const t0 = ctx.currentTime + 0.05;

    // Saída: volume geral -> limitador (nunca passa do teto) -> caixas de som
    const limitador = ctx.createDynamicsCompressor();
    limitador.threshold.value = -16; limitador.knee.value = 6; limitador.ratio.value = 12;
    limitador.attack.value = 0.003; limitador.release.value = 0.25;
    const geral = ctx.createGain();
    const teto = 0.32 * Math.min(1, volume);
    geral.gain.setValueAtTime(0, t0);
    geral.gain.linearRampToValueAtTime(teto, t0 + 0.6);
    geral.gain.setValueAtTime(teto, t0 + 3.6);
    geral.gain.linearRampToValueAtTime(0, t0 + 5.2);
    geral.connect(limitador).connect(ctx.destination);

    // Eco leve para dar espaço
    const eco = ctx.createDelay(); eco.delayTime.value = 0.28;
    const ecoVolta = ctx.createGain(); ecoVolta.gain.value = 0.28;
    const ecoSaida = ctx.createGain(); ecoSaida.gain.value = 0.35;
    eco.connect(ecoVolta).connect(eco); eco.connect(ecoSaida).connect(geral);

    const hz = n => 440 * Math.pow(2, (n - 69) / 12); // nota MIDI -> frequência

    // 1) Acorde de fundo (Fá maior com 7ª e 9ª) abrindo com filtro
    const filtro = ctx.createBiquadFilter();
    filtro.type = 'lowpass'; filtro.Q.value = 0.7;
    filtro.frequency.setValueAtTime(250, t0);
    filtro.frequency.exponentialRampToValueAtTime(2400, t0 + 2.4);
    filtro.frequency.exponentialRampToValueAtTime(900, t0 + 5);
    const pad = ctx.createGain();
    pad.gain.setValueAtTime(0, t0);
    pad.gain.linearRampToValueAtTime(0.16, t0 + 1.1);
    pad.gain.linearRampToValueAtTime(0, t0 + 5.1);
    filtro.connect(pad).connect(geral);
    for (const n of [53, 57, 60, 64, 67]) {
      for (const desafino of [-7, 7]) {
        const o = ctx.createOscillator();
        o.type = 'sawtooth'; o.frequency.value = hz(n); o.detune.value = desafino;
        o.connect(filtro); o.start(t0); o.stop(t0 + 5.3);
      }
    }

    // 2) Subida de "vento" (ruído filtrado) até a batida
    const ruido = ctx.createBuffer(1, ctx.sampleRate * 1.4, ctx.sampleRate);
    const dado = ruido.getChannelData(0);
    for (let i = 0; i < dado.length; i++) dado[i] = Math.random() * 2 - 1;
    const fonte = ctx.createBufferSource(); fonte.buffer = ruido;
    const banda = ctx.createBiquadFilter(); banda.type = 'bandpass'; banda.Q.value = 1.2;
    banda.frequency.setValueAtTime(300, t0); banda.frequency.exponentialRampToValueAtTime(5000, t0 + 1.15);
    const vento = ctx.createGain();
    vento.gain.setValueAtTime(0, t0); vento.gain.linearRampToValueAtTime(0.09, t0 + 1.05); vento.gain.linearRampToValueAtTime(0, t0 + 1.25);
    fonte.connect(banda).connect(vento).connect(geral); fonte.start(t0); fonte.stop(t0 + 1.4);

    // 3) Batida grave quando o logo aparece
    const bate = t0 + 1.2;
    const grave = ctx.createOscillator(); grave.type = 'sine';
    grave.frequency.setValueAtTime(110, bate); grave.frequency.exponentialRampToValueAtTime(42, bate + 0.5);
    const gGrave = ctx.createGain();
    gGrave.gain.setValueAtTime(0.0001, bate); gGrave.gain.exponentialRampToValueAtTime(0.5, bate + 0.02); gGrave.gain.exponentialRampToValueAtTime(0.0001, bate + 0.9);
    grave.connect(gGrave).connect(geral); grave.start(bate); grave.stop(bate + 1);

    // 4) Brilho de sino + arpejo subindo
    function nota(n, quando, dur, vol, tipo = 'sine') {
      const o = ctx.createOscillator(); o.type = tipo; o.frequency.value = hz(n);
      const g = ctx.createGain();
      g.gain.setValueAtTime(0.0001, quando);
      g.gain.exponentialRampToValueAtTime(vol, quando + 0.01);
      g.gain.exponentialRampToValueAtTime(0.0001, quando + dur);
      o.connect(g); g.connect(geral); g.connect(eco);
      o.start(quando); o.stop(quando + dur + 0.05);
    }
    nota(84, bate, 2.4, 0.07); nota(91, bate, 2.0, 0.035); // sino (Dó + Sol agudos)
    [72, 76, 79, 81, 84, 88].forEach((n, i) => nota(n, bate + 0.15 + i * 0.13, 0.9, 0.055, 'triangle'));

    return { parar() { const t = ctx.currentTime; geral.gain.cancelScheduledValues(t); geral.gain.setValueAtTime(geral.gain.value, t); geral.gain.linearRampToValueAtTime(0, t + 0.4); setTimeout(() => ctx.close && ctx.close(), 600); } };
  }
})();
