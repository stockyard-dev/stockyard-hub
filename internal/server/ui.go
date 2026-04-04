package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Stockyard Hub</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rl:#e8753a;--leather:#a0845c;--ll:#c4a87a;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c44040;--mono:'JetBrains Mono',Consolas,monospace;--serif:'Libre Baskerville',Georgia,serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);font-size:13px;line-height:1.6}
a{color:var(--rl);text-decoration:none}a:hover{color:var(--gold)}
.hdr{padding:.7rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.hdr h1{font-family:var(--serif);font-size:1.1rem;display:flex;align-items:center;gap:.6rem}.hdr h1 span{color:var(--rl)}
.badge{font-size:.55rem;padding:.15rem .4rem;border-radius:2px;text-transform:uppercase;letter-spacing:1px}
.badge-free{background:#2e261e;color:var(--cm)}.badge-pro{background:#1a3a1a;color:var(--green)}
.stats{display:flex;gap:1.2rem;padding:.7rem 1.5rem;border-bottom:1px solid var(--bg3);flex-wrap:wrap}
.stat{text-align:center;min-width:55px}.stat-n{font-size:1.3rem;color:var(--cream);font-family:var(--serif)}.stat-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px}
.stat-n.green{color:var(--green)}.stat-n.gold{color:var(--gold)}
.tabs{display:flex;gap:0;border-bottom:1px solid var(--bg3);padding:0 1.5rem}
.tab{padding:.5rem 1.2rem;cursor:pointer;font-size:.72rem;color:var(--cm);border-bottom:2px solid transparent;transition:.15s}
.tab:hover{color:var(--cream)}.tab.active{color:var(--rl);border-bottom-color:var(--rl)}
.pane{padding:1rem 1.5rem;display:none}.pane.active{display:block}
.controls{display:flex;gap:.5rem;flex-wrap:wrap;align-items:center;margin-bottom:.8rem}
.search{background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);padding:.35rem .6rem;font-family:var(--mono);font-size:.75rem;outline:none;flex:1;min-width:180px}.search:focus{border-color:var(--rust)}
.fbtn{font-family:var(--mono);font-size:.58rem;padding:.25rem .5rem;border:1px solid var(--bg3);color:var(--cm);background:transparent;cursor:pointer;transition:.15s}
.fbtn:hover,.fbtn.active{border-color:var(--rl);color:var(--rl)}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(265px,1fr));gap:.5rem}
.tool{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;transition:border-color .15s}.tool:hover{border-color:var(--leather)}
.tool-top{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:.2rem}
.tool-name{font-size:.8rem;font-weight:bold}
.tool-st{font-size:.5rem;padding:.12rem .35rem;border-radius:2px;text-transform:uppercase;letter-spacing:.5px}
.tool-st.healthy{background:#1a3a1a;color:var(--green)}.tool-st.stopped{background:var(--bg);color:var(--cm)}.tool-st.not_installed{background:var(--bg);color:var(--cm)}.tool-st.unhealthy{background:#3a1a1a;color:var(--red)}
.tool-tag{font-size:.68rem;color:var(--cd);margin-bottom:.3rem;font-style:italic;font-family:var(--serif)}
.tool-meta{font-size:.55rem;color:var(--cm);display:flex;gap:.5rem}
.tool-actions{margin-top:.35rem;display:flex;gap:.3rem}
.act{font-family:var(--mono);font-size:.58rem;padding:.2rem .5rem;border:1px solid;cursor:pointer;background:transparent;transition:.15s}
.act-start{border-color:var(--green);color:var(--green)}.act-start:hover{background:var(--green);color:var(--bg)}
.act-stop{border-color:var(--red);color:var(--red)}.act-stop:hover{background:var(--red);color:var(--cream)}
.act-install{border-color:var(--gold);color:var(--gold)}.act-install:hover{background:var(--gold);color:var(--bg)}
.act-open{border-color:var(--leather);color:var(--leather)}.act-open:hover{background:var(--leather);color:var(--bg)}
.act:disabled{opacity:.3;cursor:default}
.act-item{font-size:.72rem;padding:.4rem .5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.act-left{display:flex;gap:.5rem;align-items:center}
.act-badge{font-size:.55rem;padding:.1rem .3rem;border-radius:2px;text-transform:uppercase;letter-spacing:.5px;min-width:55px;text-align:center}
.act-badge.started{background:#1a3a1a;color:var(--green)}.act-badge.stopped{background:#3a1a1a;color:var(--red)}.act-badge.installed{background:#1a2a3a;color:#4a7ec4}.act-badge.license_set{background:#2e261e;color:var(--gold)}
.act-time{font-size:.6rem;color:var(--cm)}.act-tool{color:var(--rl);font-weight:600}
.hg{display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:.4rem}
.hc{background:var(--bg2);border:1px solid var(--bg3);padding:.5rem .6rem}
.hc-top{display:flex;justify-content:space-between;align-items:center;margin-bottom:.3rem}
.hc-name{font-size:.75rem;font-weight:600}
.hc-st{font-size:.55rem;padding:.1rem .3rem;border-radius:2px}
.hc-st.healthy{background:#1a3a1a;color:var(--green)}.hc-st.unhealthy{background:#3a1a1a;color:var(--red)}.hc-st.stopped{color:var(--cm)}
.hbar{display:flex;gap:1px;height:14px;margin-top:.3rem}
.hp{flex:1;min-width:3px}.hp.healthy{background:var(--green)}.hp.unhealthy{background:var(--red)}.hp.stopped{background:var(--bg3)}
.cfg{max-width:500px}
.cfg-row{display:flex;gap:.5rem;align-items:center;margin-bottom:.5rem}
.cfg-row label{font-size:.65rem;color:var(--leather);text-transform:uppercase;letter-spacing:1px;min-width:80px}
.cfg-row input{background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.35rem .5rem;font-family:var(--mono);font-size:.75rem;flex:1;outline:none}.cfg-row input:focus{border-color:var(--rust)}
.btn{font-family:var(--mono);font-size:.7rem;padding:.35rem .8rem;border:1px solid var(--rl);color:var(--rl);background:transparent;cursor:pointer}.btn:hover{background:var(--rust);color:var(--cream)}
.empty{text-align:center;padding:2rem;color:var(--cm);font-style:italic;font-family:var(--serif)}
.toast{position:fixed;bottom:1rem;right:1rem;background:var(--bg2);border:1px solid var(--green);color:var(--green);padding:.5rem 1rem;font-size:.72rem;display:none;z-index:100}
</style>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
</head><body>
<div class="hdr">
  <h1><svg viewBox="0 0 64 64" width="20" height="20" fill="none"><rect x="8" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="28" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="48" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="8" y="27" width="48" height="7" rx="2.5" fill="#c4a87a"/></svg><span>Stockyard</span> Hub</h1>
  <div style="display:flex;gap:1rem;align-items:center;font-size:.7rem"><span id="tbadge" class="badge badge-free">Free</span></div>
</div>
<div class="stats" id="stats"></div>
<div class="tabs">
  <div class="tab active" data-tab="tools" onclick="sw('tools')">Tools</div>
  <div class="tab" data-tab="activity" onclick="sw('activity')">Activity</div>
  <div class="tab" data-tab="health" onclick="sw('health')">Health</div>
  <div class="tab" data-tab="settings" onclick="sw('settings')">Settings</div>
</div>
<div id="pane-tools" class="pane active">
  <div class="controls">
    <input class="search" id="search" placeholder="Search 150 tools..." oninput="dload()">
    <button class="fbtn active" data-f="" onclick="sf('')">All</button>
    <button class="fbtn" data-f="installed" onclick="sf('installed')">Installed</button>
    <button class="fbtn" data-f="running" onclick="sf('running')">Running</button>
    <span style="color:var(--bg3)">|</span>
    <button class="fbtn active" data-c="" onclick="sc('')">Any</button>
    <button class="fbtn" data-c="developer" onclick="sc('developer')">Dev</button>
    <button class="fbtn" data-c="operations" onclick="sc('operations')">Ops</button>
    <button class="fbtn" data-c="finance" onclick="sc('finance')">Fin</button>
    <button class="fbtn" data-c="creator" onclick="sc('creator')">Create</button>
    <button class="fbtn" data-c="personal" onclick="sc('personal')">Personal</button>
  </div>
  <div class="grid" id="grid"></div>
</div>
<div id="pane-activity" class="pane"><div id="aList"></div></div>
<div id="pane-health" class="pane">
  <p style="font-size:.7rem;color:var(--leather);margin-bottom:.8rem">Health checks every 30s for installed tools. Last 20 checks shown.</p>
  <div class="hg" id="hGrid"></div>
</div>
<div id="pane-settings" class="pane">
  <h3 style="font-family:var(--serif);font-size:.9rem;margin-bottom:1rem">License</h3>
  <div class="cfg">
    <div class="cfg-row"><label>Key</label><input type="text" id="lkIn" placeholder="SY-xxxxx"><button class="btn" onclick="sLic()">Save</button></div>
    <p style="font-size:.6rem;color:var(--cm);margin-top:.3rem">Set once. Applies Pro to all tools. <a href="https://stockyard.dev/pricing/" target="_blank">Get a key</a></p>
    <div id="lkSt" style="font-size:.65rem;color:var(--leather);margin-top:.5rem"></div>
  </div>
</div>
<div class="toast" id="toast"></div>
<script>
let flt='',cat='',tmr=null;
function dload(){clearTimeout(tmr);tmr=setTimeout(lt,200)}
function sf(f){flt=f;document.querySelectorAll('[data-f]').forEach(b=>b.classList.toggle('active',b.dataset.f===f));lt()}
function sc(c){cat=c;document.querySelectorAll('[data-c]').forEach(b=>b.classList.toggle('active',b.dataset.c===c));lt()}
function sw(t){document.querySelectorAll('.tab').forEach(x=>x.classList.toggle('active',x.dataset.tab===t));document.querySelectorAll('.pane').forEach(p=>p.classList.toggle('active',p.id==='pane-'+t));if(t==='activity')la();if(t==='health')lh();if(t==='settings')ll()}
function ta(d){const s=Math.floor((Date.now()-new Date(d))/1e3);if(s<60)return s+'s ago';if(s<3600)return Math.floor(s/60)+'m ago';if(s<86400)return Math.floor(s/3600)+'h ago';return Math.floor(s/86400)+'d ago'}
function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')}
function toast(m){const t=document.getElementById('toast');t.textContent=m;t.style.display='block';setTimeout(()=>t.style.display='none',3000)}
async function ls(){const d=await(await fetch('/api/stats')).json();document.getElementById('stats').innerHTML=[{n:d.total,l:'Total',c:''},{n:d.installed,l:'Installed',c:'gold'},{n:d.running,l:'Running',c:'green'},{n:d.healthy,l:'Healthy',c:'green'}].map(s=>'<div class="stat"><div class="stat-n '+s.c+'">'+s.n+'</div><div class="stat-l">'+s.l+'</div></div>').join('')}
async function lt(){
  let u='/api/tools?';const q=document.getElementById('search').value;
  if(q)u+='q='+encodeURIComponent(q)+'&';if(flt)u+='status='+flt+'&';if(cat)u+='category='+cat+'&';
  const tools=(await(await fetch(u)).json()).tools||[];
  if(!tools.length){document.getElementById('grid').innerHTML='<div class="empty">No tools match.</div>';ls();return}
  document.getElementById('grid').innerHTML=tools.map(t=>{
    const a=t.health==='not_installed'?'<button class="act act-install" onclick="da(event,\'install\',\''+t.slug+'\')">Install</button>':t.running?'<button class="act act-stop" onclick="da(event,\'stop\',\''+t.slug+'\')">Stop</button><button class="act act-open" onclick="window.open(\'http://localhost:'+t.port+'/ui\')">Open :'+t.port+'</button>':'<button class="act act-start" onclick="da(event,\'start\',\''+t.slug+'\')">Start</button>';
    return '<div class="tool"><div class="tool-top"><div class="tool-name">'+esc(t.name)+'</div><div class="tool-st '+t.health+'">'+t.health.replace('_',' ')+'</div></div><div class="tool-tag">'+esc(t.tagline)+'</div><div class="tool-meta"><span>:'+t.port+'</span><span>'+t.category+'</span></div><div class="tool-actions">'+a+'</div></div>'}).join('');ls()}
async function da(e,action,slug){e.target.disabled=true;e.target.textContent=action==='install'?'Installing...':action==='start'?'Starting...':'Stopping...';await fetch('/api/tools/'+slug+'/'+action,{method:'POST'});toast(slug+' '+action+(action==='stop'?'ped':'ed'));setTimeout(lt,action==='install'?2000:action==='start'?1500:500)}
async function la(){const acts=(await(await fetch('/api/activity')).json()).activity||[];const el=document.getElementById('aList');if(!acts.length){el.innerHTML='<div class="empty">No activity yet.</div>';return}el.innerHTML=acts.map(a=>'<div class="act-item"><div class="act-left"><span class="act-badge '+a.action+'">'+a.action+'</span><span class="act-tool">'+esc(a.tool)+'</span><span style="color:var(--cd)">'+esc(a.detail)+'</span></div><span class="act-time">'+ta(a.created_at)+'</span></div>').join('')}
async function lh(){const recs=(await(await fetch('/api/health-history?limit=500')).json()).records||[];const bt={};recs.forEach(r=>{if(!bt[r.tool])bt[r.tool]=[];bt[r.tool].push(r)});const el=document.getElementById('hGrid');const slugs=Object.keys(bt).sort();if(!slugs.length){el.innerHTML='<div class="empty">No health data yet.</div>';return}el.innerHTML=slugs.map(s=>{const ch=bt[s].slice(0,20).reverse();const lat=bt[s][0];const avg=Math.round(ch.reduce((a,c)=>a+c.response_ms,0)/ch.length);const pips=ch.map(c=>'<div class="hp '+c.status+'" title="'+c.status+' '+c.response_ms+'ms"></div>').join('');return '<div class="hc"><div class="hc-top"><div class="hc-name">'+esc(s)+'</div><div class="hc-st '+lat.status+'">'+lat.status+' '+avg+'ms</div></div><div class="hbar">'+pips+'</div></div>'}).join('')}
async function ll(){const d=await(await fetch('/api/license')).json();const el=document.getElementById('lkSt');if(d.license_key)el.innerHTML='Current: <span style="color:var(--cream)">'+d.license_key+'</span> &middot; <span style="color:'+(d.tier==='pro'?'var(--green)':'var(--cm)')+'">'+d.tier+'</span>';else el.textContent='No license key set.'}
async function sLic(){const k=document.getElementById('lkIn').value.trim();if(!k)return;await fetch('/api/license',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({key:k})});toast('License saved');document.getElementById('lkIn').value='';ll()}
lt();fetch('/api/tier').then(r=>r.json()).then(j=>{const b=document.getElementById('tbadge');if(j.tier==='pro'){b.className='badge badge-pro';b.textContent='Pro'}}).catch(()=>{});
setInterval(()=>{const a=document.querySelector('.tab.active');if(a&&a.dataset.tab==='tools')lt()},30000);
</script></body></html>` + "\n"
