package main

// Domínios de anúncio e rastreamento bloqueados nas janelas do TV Central.
// Conservador de propósito: nada que os streamings usem para tocar o vídeo.
var dominiosBloqueados = []string{
	"doubleclick.net", "googlesyndication.com", "googleadservices.com", "adservice.google.com",
	"pagead2.googlesyndication.com", "google-analytics.com", "googletagmanager.com", "googletagservices.com",
	"amazon-adsystem.com", "adnxs.com", "adsrvr.org", "taboola.com", "outbrain.com", "criteo.com", "criteo.net",
	"pubmatic.com", "rubiconproject.com", "openx.net", "casalemedia.com", "moatads.com", "scorecardresearch.com",
	"quantserve.com", "advertising.com", "adform.net", "smartadserver.com", "teads.tv", "yieldmo.com",
	"facebook.net/tr", "connect.facebook.net", "hotjar.com", "mouseflow.com", "clarity.ms", "bat.bing.com",
	"ads.linkedin.com", "analytics.tiktok.com", "ads-twitter.com", "static.ads-twitter.com", "branch.io",
	"app-measurement.com", "doubleverify.com", "adsafeprotected.com", "serving-sys.com", "zedo.com",
	"youtube.com/api/stats/ads", "youtube.com/pagead", "youtube.com/ptracking", "youtube.com/get_midroll_info",
}

// padroesBloqueio vira a lista no formato do Network.setBlockedURLs (curinga *).
func padroesBloqueio() []string {
	l := make([]string, 0, len(dominiosBloqueados))
	for _, d := range dominiosBloqueados {
		l = append(l, "*"+d+"*")
	}
	return l
}
