package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Stockyard Hub</title>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rl:#e8753a;--leather:#a0845c;--ll:#c4a87a;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c44040;--mono:'JetBrains Mono',Consolas,monospace;--serif:'Libre Baskerville',Georgia,serif}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--cream);font-family:var(--mono);font-size:13px;line-height:1.6}
a{color:var(--rl);text-decoration:none}a:hover{color:var(--gold)}
.hdr{padding:.7rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.hdr h1{font-family:var(--serif);font-size:1.1rem}.hdr h1 span{color:var(--rl)}
.hdr-right{display:flex;align-items:center;gap:1rem;font-size:.7rem}
.badge{font-size:.55rem;padding:.15rem .4rem;border-radius:2px;text-transform:uppercase;letter-spacing:1px}
.badge-free{background:#2e261e;color:var(--cm)}.badge-pro{background:#1a3a1a;color:var(--green)}

.stats{display:flex;gap:1.5rem;padding:.8rem 1.5rem;border-bottom:1px solid var(--bg3);flex-wrap:wrap}
.stat{text-align:center}.stat-n{font-size:1.4rem;color:var(--rl);font-family:var(--serif)}.stat-l{font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px}

.tabs{display:flex;gap:0;padding:0 1.5rem;border-bottom:1px solid var(--bg3)}
.tab{padding:.5rem 1rem;cursor:pointer;font-size:.72rem;color:var(--cm);border-bottom:2px solid transparent;transition:.15s}
.tab:hover{color:var(--cream)}.tab.active{color:var(--rl);border-bottom-color:var(--rl)}

.pane{padding:1rem 1.5rem}
.controls{display:flex;gap:.5rem;flex-wrap:wrap;align-items:center;margin-bottom:1rem}
.search{background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);padding:.35rem .6rem;font-family:var(--mono);font-size:.78rem;outline:none;flex:1;min-width:200px}
.search:focus{border-color:var(--rust)}
.fbtn{font-family:var(--mono);font-size:.6rem;padding:.25rem .5rem;border:1px solid var(--bg3);color:var(--cm);background:transparent;cursor:pointer}
.fbtn:hover,.fbtn.active{border-color:var(--leather);color:var(--leather)}

.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:.5rem}
.tool{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem;transition:border-color .15s}
.tool:hover{border-color:var(--leather)}
.tool-top{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:.3rem}
.tool-name{font-size:.82rem;font-weight:bold}
.tool-st{font-size:.55rem;padding:.12rem .35rem;border-radius:2px;text-transform:uppercase;letter-spacing:1px}
.tool-st.healthy{background:#1a3a1a;color:var(--green)}
.tool-st.stopped{background:#2e261e;color:var(--cm)}
.tool-st.not_installed{background:var(--bg);color:var(--cm)}
.tool-st.unhealthy{background:#3a1a1a;color:var(--red)}
.tool-tag{font-size:.7rem;color:var(--cd);margin-bottom:.3rem}
.tool-meta{font-size:.58rem;color:var(--cm);display:flex;gap:.5rem}
.tool-acts{margin-top:.4rem;display:flex;gap:.3rem}
.act{font-family:var(--mono);font-size:.58rem;padding:.2rem .5rem;border:1px solid;cursor:pointer;background:transparent;transition:.15s}
.act-start{border-color:var(--green);color:var(--green)}.act-start:hover{background:var(--green);color:var(--bg)}
.act-stop{border-color:var(--red);color:var(--red)}.act-stop:hover{background:var(--red);color:var(--cream)}
.act-install{border-color:var(--gold);color:var(--gold)}.act-install:hover{background:var(--gold);color:var(--bg)}
.act-open{border-color:var(--leather);color:var(--leather)}.act-open:hover{background:var(--leather);color:var(--bg)}
.act:disabled{opacity:.3;cursor:default}

.act-item{font-size:.72rem;padding:.4rem 0;border-bottom:1px solid var(--bg3);color:var(--cd);display:flex;justify-content:space-between;align-items:center}
.act-left{display:flex;gap:.5rem;align-items:center}
.act-action{font-weight:600;color:var(--rl);text-transform:uppercase;font-size:.6rem;letter-spacing:.5px;min-width:65px}
.act-time{color:var(--cm);font-size:.6rem}
.act-slug{color:var(--leather);font-size:.65rem}

.license-box{background:var(--bg2);border:1px solid var(--bg3);padding:1rem;max-width:500px}
.license-box label{font-size:.65rem;color:var(--leather);text-transform:uppercase;letter-spacing:1px;display:block;margin-bottom:.3rem}
.license-box input{background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .6rem;font-family:var(--mono);font-size:.8rem;width:100%;outline:none;margin-bottom:.5rem}
.license-box input:focus{border-color:var(--rust)}
.btn{font-family:var(--mono);font-size:.7rem;padding:.35rem .8rem;border:1px solid var(--rl);color:var(--rl);background:transparent;cursor:pointer}
.btn:hover{background:var(--rl);color:var(--cream)}

.empty{text-align:center;padding:2rem;color:var(--cm);font-style:italic;font-family:var(--serif)}
.toast{position:fixed;bottom:1rem;right:1rem;background:var(--bg2);border:1px solid var(--green);color:var(--green);padding:.5rem 1rem;font-size:.75rem;display:none;z-index:100}
</style></head>
<body>
<div class="hdr">
  <h1><span>Stockyard</span> Hub</h1>
  <div class="hdr-right">
    <span id="tier-badge" class="badge badge-free">Free</span>
    <a href="https://stockyard.dev/cli/" style="font-size:.65rem;color:var(--cm)">CLI</a>
    <a href="https://stockyard.dev/tools/" style="font-size:.65rem;color:var(--cm)">Docs</a>
  </div>
</div>

<div class="stats" id="stats"></div>

<div class="tabs">
  <div class="tab active" data-tab="tools" onclick="switchTab('tools')">Tools</div>
  <div class="tab" data-tab="activity" onclick="switchTab('activity')">Activity</div>
  <div class="tab" data-tab="settings" onclick="switchTab('settings')">Settings</div>
</div>

<div id="pane-tools" class="pane">
  <div class="controls">
    <input class="search" id="search" placeholder="Search 150 tools..." oninput="debounceLoad()">
    <button class="fbtn active" data-f="" onclick="setFilter('')">All</button>
    <button class="fbtn" data-f="installed" onclick="setFilter('installed')">Installed</button>
    <button class="fbtn" data-f="running" onclick="setFilter('running')">Running</button>
    <span style="width:1px;height:18px;background:var(--bg3);margin:0 .2rem"></span>
    <button class="fbtn active" data-c="" onclick="setCat('')">Any</button>
    <button class="fbtn" data-c="developer" onclick="setCat('developer')">Dev</button>
    <button class="fbtn" data-c="operations" onclick="setCat('operations')">Ops</button>
    <button class="fbtn" data-c="finance" onclick="setCat('finance')">Finance</button>
    <button class="fbtn" data-c="creator" onclick="setCat('creator')">Creator</button>
    <button class="fbtn" data-c="personal" onclick="setCat('personal')">Personal</button>
  </div>
  <div class="grid" id="grid"></div>
</div>

<div id="pane-activity" class="pane" style="display:none">
  <div id="actList"></div>
</div>

<div id="pane-settings" class="pane" style="display:none">
  <h3 style="font-family:var(--serif);font-size:.9rem;margin-bottom:1rem">License</h3>
  <div class="license-box">
    <label>License key</label>
    <input type="text" id="licenseKey" placeholder="SY-xxxxx">
    <div style="display:flex;gap:.5rem;align-items:center">
      <button class="btn" onclick="saveLicense()">Save</button>
      <span id="licenseStatus" style="font-size:.65rem;color:var(--cm)"></span>
    </div>
    <p style="font-size:.65rem;color:var(--cm);margin-top:.8rem">Your license key unlocks Pro on all tools started through Hub. Get a key at <a href="https://stockyard.dev/pricing/">stockyard.dev/pricing</a></p>
  </div>
  <h3 style="font-family:var(--serif);font-size:.9rem;margin:1.5rem 0 .5rem">About</h3>
  <div style="font-size:.72rem;color:var(--cd)">
    <p>Stockyard Hub manages all 150 Stockyard tools from one dashboard.</p>
    <p style="margin-top:.3rem;color:var(--cm)">Install, start, stop, and monitor tools. Health checks run every 30 seconds.</p>
    <p style="margin-top:.5rem"><a href="https://stockyard.dev/hub/">stockyard.dev/hub</a> &middot; <a href="https://github.com/stockyard-dev/stockyard-hub">GitHub</a></p>
  </div>
</div>

<div class="toast" id="toast"></div>

<script>
let filter='',cat='',timer=null;

function debounceLoad(){clearTimeout(timer);timer=setTimeout(loadTools,200)}

function setFilter(f){
  filter=f;
  document.querySelectorAll('[data-f]').forEach(b=>b.classList.toggle('active',b.dataset.f===f));
  loadTools();
}
function setCat(c){
  cat=c;
  document.querySelectorAll('[data-c]').forEach(b=>b.classList.toggle('active',b.dataset.c===c));
  loadTools();
}

function timeAgo(d){
  const s=Math.floor((Date.now()-new Date(d))/1e3);
  if(s<60)return s+'s ago';if(s<3600)return Math.floor(s/60)+'m ago';
  if(s<86400)return Math.floor(s/3600)+'h ago';return Math.floor(s/86400)+'d ago';
}

async function loadStats(){
  const r=await fetch('/api/stats');const d=await r.json();
  document.getElementById('stats').innerHTML=[
    {n:d.total,l:'Total'},{n:d.installed,l:'Installed'},{n:d.running,l:'Running'},{n:d.healthy,l:'Healthy'}
  ].map(s=>'<div class="stat"><div class="stat-n">'+s.n+'</div><div class="stat-l">'+s.l+'</div></div>').join('');
}

async function loadTools(){
  let url='/api/tools?';
  const q=document.getElementById('search').value;
  if(q)url+='q='+encodeURIComponent(q)+'&';
  if(filter)url+='status='+filter+'&';
  if(cat)url+='category='+cat+'&';
  const r=await fetch(url);const d=await r.json();
  const tools=d.tools||[];
  if(!tools.length){document.getElementById('grid').innerHTML='<div class="empty">No tools match your filters.</div>';return}
  document.getElementById('grid').innerHTML=tools.map(t=>{
    const actions=t.health==='not_installed'
      ?'<button class="act act-install" onclick="install(event,\''+t.slug+'\')">Install</button>'
      :t.running
        ?'<button class="act act-stop" onclick="stop(event,\''+t.slug+'\')">Stop</button><button class="act act-open" onclick="window.open(\'http://localhost:'+t.port+'/ui\')">Open</button>'
        :'<button class="act act-start" onclick="start(event,\''+t.slug+'\')">Start</button>';
    return '<div class="tool"><div class="tool-top"><div class="tool-name">'+esc(t.name)+'</div><div class="tool-st '+t.health+'">'+t.health.replace('_',' ')+'</div></div><div class="tool-tag">'+esc(t.tagline)+'</div><div class="tool-meta"><span>:'+t.port+'</span><span>'+t.category+'</span></div><div class="tool-acts">'+actions+'</div></div>';
  }).join('');
  loadStats();
}

function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')}

async function start(e,slug){e.target.disabled=true;e.target.textContent='Starting...';await fetch('/api/tools/'+slug+'/start',{method:'POST'});toast(slug+' started');setTimeout(loadTools,1500)}
async function stop(e,slug){e.target.disabled=true;await fetch('/api/tools/'+slug+'/stop',{method:'POST'});toast(slug+' stopped');setTimeout(loadTools,500)}
async function install(e,slug){e.target.disabled=true;e.target.textContent='Installing...';await fetch('/api/tools/'+slug+'/install',{method:'POST'});toast(slug+' installed');setTimeout(loadTools,2000)}

// ── Activity ──
async function loadActivity(){
  const r=await fetch('/api/activity');const d=await r.json();
  const acts=d.activity||[];
  const el=document.getElementById('actList');
  if(!acts.length){el.innerHTML='<div class="empty">No activity yet. Install or start a tool to see events here.</div>';return}
  el.innerHTML=acts.map(a=>
    '<div class="act-item"><div class="act-left"><span class="act-action">'+esc(a.action)+'</span><span class="act-slug">'+esc(a.tool)+'</span><span>'+esc(a.detail)+'</span></div><span class="act-time">'+timeAgo(a.created_at)+'</span></div>'
  ).join('');
}

// ── License ──
async function loadLicense(){
  const r=await fetch('/api/license');const d=await r.json();
  if(d.license_key){document.getElementById('licenseKey').placeholder=d.license_key}
}

async function saveLicense(){
  const key=document.getElementById('licenseKey').value.trim();
  if(!key){return}
  await fetch('/api/license',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({key})});
  document.getElementById('licenseStatus').textContent='Saved. Restart tools to apply.';
  document.getElementById('licenseKey').value='';
  loadLicense();
  toast('License key saved');
}

// ── Tabs ──
function switchTab(tab){
  document.querySelectorAll('.tab').forEach(t=>t.classList.toggle('active',t.dataset.tab===tab));
  document.getElementById('pane-tools').style.display=tab==='tools'?'':'none';
  document.getElementById('pane-activity').style.display=tab==='activity'?'':'none';
  document.getElementById('pane-settings').style.display=tab==='settings'?'':'none';
  if(tab==='activity')loadActivity();
  if(tab==='settings')loadLicense();
}

function toast(msg){const t=document.getElementById('toast');t.textContent=msg;t.style.display='block';setTimeout(()=>t.style.display='none',3000)}

// ── Init ──
loadTools();
loadLicense();
fetch('/api/tier').then(r=>r.json()).then(j=>{
  const b=document.getElementById('tier-badge');
  if(j.tier==='pro'){b.className='badge badge-pro';b.textContent='Pro'}
}).catch(()=>{});

// Auto-refresh every 30s
setInterval(()=>{loadTools()},30000);
</script></body></html>`
