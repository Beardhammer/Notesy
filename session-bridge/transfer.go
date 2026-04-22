package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"
)

const transferTTL = 10 * time.Minute

type TransferNewHandler struct {
	store *Store
	authz *AuthzHandler
	tmpl  *template.Template
}

func NewTransferNewHandler(s *Store, authz *AuthzHandler) *TransferNewHandler {
	return &TransferNewHandler{
		store: s,
		authz: authz,
		tmpl:  template.Must(template.New("new").Parse(newCodeHTML)),
	}
}

func (h *TransferNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Validate independently. Do NOT trust any X-Authentik-Username header
	// the client may have supplied — authenticate against the transfer
	// cookie or oauth2-proxy directly.
	res, err := h.authz.Authenticate(r)
	if err != nil {
		http.Error(w, "auth check failed", http.StatusBadGateway)
		return
	}
	if res == nil {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	user := res.Username
	code, err := h.store.Issue(user, transferTTL)
	if err != nil {
		http.Error(w, "issue failed", 500)
		return
	}

	// API clients (curl with no Accept header, or explicit application/json)
	// still get the JSON contract. Browsers get HTML.
	if strings.Contains(r.Header.Get("Accept"), "application/json") ||
		!strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":             code,
			"expiresInSeconds": int(transferTTL.Seconds()),
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.Execute(w, map[string]any{
		"User":             user,
		"Code":             code,
		"ExpiresInSeconds": int(transferTTL.Seconds()),
	})
}

type TransferClaimHandler struct {
	store  *Store
	signer *Signer
	tmpl   *template.Template
}

func NewTransferClaimHandler(s *Store, sg *Signer) *TransferClaimHandler {
	return &TransferClaimHandler{
		store:  s,
		signer: sg,
		tmpl:   template.Must(template.New("claim").Parse(claimFormHTML)),
	}
}

func (h *TransferClaimHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = h.tmpl.Execute(w, map[string]any{"Error": ""})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", 400)
		return
	}
	code := r.PostFormValue("code")

	sub, err := h.store.Claim(code)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = h.tmpl.Execute(w, map[string]any{"Error": "Invalid or expired code."})
		return
	}

	tok, err := h.signer.Sign(sub, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "sign failed", 500)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "notesy-transfer",
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

const pageCSS = `
<style>
*,*::before,*::after{box-sizing:border-box}
body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;
font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;
background:#0f1115;color:#e6e8ec}
.card{background:#1a1d24;border:1px solid #2a2f3a;border-radius:12px;
padding:2.5rem;max-width:480px;width:90%;
box-shadow:0 8px 32px rgba(0,0,0,.4)}
h1{margin:0 0 .5rem;font-size:1.25rem;font-weight:600;letter-spacing:-.01em}
p.sub{margin:0 0 2rem;color:#8a92a5;font-size:.95rem;line-height:1.5}
.code{display:block;font-family:ui-monospace,"SF Mono",Menlo,monospace;
font-size:3.5rem;font-weight:700;letter-spacing:.25em;
text-align:center;color:#7ee7c7;padding:1.25rem .5rem;
background:#0a0c10;border-radius:8px;border:1px solid #2a2f3a;
user-select:all;cursor:copy;margin:0 0 1rem}
.code:hover{background:#0c0f14;border-color:#3a4050}
.meta{display:flex;justify-content:space-between;align-items:center;
font-size:.85rem;color:#8a92a5;margin-bottom:1.25rem}
.timer{font-family:ui-monospace,monospace;color:#e6e8ec}
.timer.warn{color:#f5a623}
.timer.expired{color:#e5484d}
.progress{height:3px;background:#2a2f3a;border-radius:2px;overflow:hidden;margin-bottom:1rem}
.progress-bar{height:100%;background:#7ee7c7;transition:width 1s linear,background .3s}
.progress-bar.warn{background:#f5a623}
.progress-bar.expired{background:#e5484d}
.btn{display:inline-block;width:100%;padding:.75rem 1rem;border:0;border-radius:6px;
background:#3b82f6;color:#fff;font-size:1rem;font-weight:500;cursor:pointer;
font-family:inherit;transition:background .15s}
.btn:hover{background:#2563eb}
.btn-ghost{background:transparent;border:1px solid #2a2f3a;color:#8a92a5;margin-top:.5rem}
.btn-ghost:hover{background:#2a2f3a;color:#e6e8ec}
.input{width:100%;padding:.875rem 1rem;font-size:1.5rem;font-family:ui-monospace,monospace;
letter-spacing:.15em;text-align:center;background:#0a0c10;border:1px solid #2a2f3a;
border-radius:6px;color:#e6e8ec;margin-bottom:1rem;outline:none;transition:border .15s}
.input:focus{border-color:#3b82f6}
.error{color:#e5484d;font-size:.9rem;margin:0 0 1rem;padding:.75rem;
background:rgba(229,72,77,.1);border-radius:6px;text-align:center}
.user{color:#7ee7c7;font-weight:500}
.footer{margin-top:1.5rem;font-size:.8rem;color:#5a6170;text-align:center;line-height:1.5}
</style>
`

const newCodeHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Session Transfer Code</title>
` + pageCSS + `
</head>
<body>
<div class="card">
  <h1>Session Transfer Code</h1>
  <p class="sub">Signed in as <span class="user">{{.User}}</span>. Enter this code on the public link within the time limit to carry your session over.</p>

  <div class="code" id="code" title="Click to copy">{{.Code}}</div>

  <div class="progress"><div class="progress-bar" id="bar" style="width:100%"></div></div>

  <div class="meta">
    <span>Expires in</span>
    <span class="timer" id="timer">{{.ExpiresInSeconds}}s</span>
  </div>

  <button class="btn" id="copyBtn" type="button">Copy code</button>
  <button class="btn btn-ghost" onclick="location.reload()">Issue a new code</button>

  <p class="footer">Single-use. 10-minute expiry.<br>Anyone with this code becomes you for 7 days.</p>
</div>

<script>
(function(){
  var total = {{.ExpiresInSeconds}};
  var remaining = total;
  var timer = document.getElementById('timer');
  var bar = document.getElementById('bar');
  var code = document.getElementById('code');
  var copyBtn = document.getElementById('copyBtn');

  function fmt(s){
    if (s <= 0) return 'expired';
    var m = Math.floor(s/60), r = s%60;
    return m + ':' + (r<10?'0':'') + r;
  }
  function tick(){
    timer.textContent = fmt(remaining);
    var pct = Math.max(0, (remaining/total)*100);
    bar.style.width = pct + '%';
    var warn = remaining <= 60, expired = remaining <= 0;
    timer.classList.toggle('warn', warn && !expired);
    timer.classList.toggle('expired', expired);
    bar.classList.toggle('warn', warn && !expired);
    bar.classList.toggle('expired', expired);
    if (expired){ code.style.opacity = '0.3'; copyBtn.disabled = true; copyBtn.textContent = 'Code expired'; return; }
    remaining -= 1;
    setTimeout(tick, 1000);
  }
  tick();

  function copy(){
    var t = code.textContent.trim();
    if (navigator.clipboard){
      navigator.clipboard.writeText(t).then(function(){
        copyBtn.textContent = 'Copied!';
        setTimeout(function(){ copyBtn.textContent = 'Copy code'; }, 1500);
      });
    } else {
      var r = document.createRange(); r.selectNode(code);
      window.getSelection().removeAllRanges();
      window.getSelection().addRange(r);
      document.execCommand('copy');
      copyBtn.textContent = 'Copied!';
      setTimeout(function(){ copyBtn.textContent = 'Copy code'; }, 1500);
    }
  }
  copyBtn.addEventListener('click', copy);
  code.addEventListener('click', copy);
})();
</script>
</body>
</html>`

const claimFormHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Redeem Session Transfer Code</title>
` + pageCSS + `
</head>
<body>
<div class="card">
  <h1>Redeem transfer code</h1>
  <p class="sub">Enter the 8-digit code shown on the internal network. You'll be signed in for 7 days without needing to reach the identity provider.</p>

  {{if .Error}}<p class="error">{{.Error}}</p>{{end}}

  <form method="POST" action="/transfer/claim" autocomplete="off">
    <input class="input" name="code" id="code" type="text"
      inputmode="numeric" pattern="[0-9]{8}" maxlength="8"
      autocomplete="off" autofocus required
      placeholder="00000000">
    <button class="btn" type="submit">Redeem</button>
  </form>

  <p class="footer">Codes are single-use and valid for 10 minutes after they're issued.</p>
</div>
<script>
// digits only, auto-submit when 8 entered
var i = document.getElementById('code');
i.addEventListener('input', function(){
  i.value = i.value.replace(/\D/g,'').slice(0,8);
  if (i.value.length === 8) i.form.submit();
});
</script>
</body>
</html>`
