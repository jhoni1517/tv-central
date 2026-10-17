// Necessário para instalar como app. Não guarda cache: o controle precisa da rede local.
self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', e => e.waitUntil(self.clients.claim()));
self.addEventListener('fetch', () => {});
