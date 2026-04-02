package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Stockyard Hub</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rl:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c45d2c;--mono:'JetBrains Mono',Consolas,monospace;--serif:'Libre Baskerville',Georgia,serif}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--cream);font-family:var(--mono);font-size:13px;line-height:1.6}
.hdr{padding:.8rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.hdr h1{font-family:var(--serif);font-size:1.1rem}.hdr h1 span{color:var(--rl)}
.stats{display:flex;gap:1.5rem;padding:.8rem 1.5rem;border-bottom:1px solid var(--bg3);flex-wrap:wrap}
.stat{text-align:center}.stat-n{font-size:1.4rem;color:var(--rl);font-family:var(--serif)}.stat-l{font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px}
.controls{padding:.8rem 1.5rem;display:flex;gap:.5rem;flex-wrap:wrap;align-items:center}
.search{background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);padding:.35rem .6rem;font-family:var(--mono);font-size:.78rem;outline:none;flex:1;min-width:200px}
.search:focus{border-color:var(--rust)}
.fbtn{font-family:var(--mono);font-size:.6rem;padding:.25rem .5rem;border:1px solid var(--bg3);color:var(--cm);background:transparent;cursor:pointer}
.fbtn:hover,.fbtn.active{border-color:var(--leather);color:var(--leather)}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:.5rem;padding:1rem 1.5rem}
.tool{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem;transition:border-color .15s}
.tool:hover{border-color:var(--leather)}
.tool-top{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:.3rem}
.tool-name{font-size:.85rem;font-weight:bold}
.tool-status{font-size:.6rem;padding:.15rem .4rem;border-radius:2px;text-transform:uppercase;letter-spacing:1px}
.tool-status.healthy{background:#1a3a1a;color:var(--green)}
.tool-status.stopped{background:#2e261e;color:var(--cm)}
.tool-status.not_installed{background:var(--bg);color:var(--cm)}
.tool-status.unhealthy{background:#3a1a1a;color:var(--red)}
.tool-tagline{font-size:.72rem;color:var(--cd);margin-bottom:.4rem}
.tool-meta{font-size:.6rem;color:var(--cm);display:flex;gap:.5rem}
.tool-actions{margin-top:.4rem;display:flex;gap:.3rem}
.act{font-family:var(--mono);font-size:.6rem;padding:.2rem .5rem;border:1px solid;cursor:pointer;background:transparent}
.act-start{border-color:var(--green);color:var(--green)}.act-start:hover{background:var(--green);color:var(--bg)}
.act-stop{border-color:var(--red);color:var(--red)}.act-stop:hover{background:var(--red);color:var(--cream)}
.act-install{border-color:var(--gold);color:var(--gold)}.act-install:hover{background:var(--gold);color:var(--bg)}
.act-open{border-color:var(--leather);color:var(--leather)}.act-open:hover{background:var(--leather);color:var(--bg)}
.act:disabled{opacity:.3;cursor:default}
.toast{position:fixed;bottom:1rem;right:1rem;background:var(--bg2);border:1px solid var(--green);color:var(--green);padding:.5rem 1rem;font-size:.75rem;display:none;z-index:100}
</style></head>
<body>
<div class="hdr"><h1><span>Stockyard</span> Hub</h1><div style="font-size:.7rem;color:var(--cm)">Tool Management Dashboard</div></div>
<div class="stats" id="stats"></div>
<div class="controls">
  <input class="search" id="search" placeholder="Search tools..." oninput="debounceLoad()">
  <button class="fbtn active" onclick="setFilter('')">All</button>
  <button class="fbtn" onclick="setFilter('installed')">Installed</button>
  <button class="fbtn" onclick="setFilter('running')">Running</button>
  <button class="fbtn" onclick="setCat('')">Any category</button>
  <button class="fbtn" onclick="setCat('developer')">Dev</button>
  <button class="fbtn" onclick="setCat('operations')">Ops</button>
  <button class="fbtn" onclick="setCat('finance')">Finance</button>
  <button class="fbtn" onclick="setCat('creator')">Creator</button>
  <button class="fbtn" onclick="setCat('personal')">Personal</button>
</div>
<div class="grid" id="grid"></div>
<div class="toast" id="toast"></div>

<script>
let filter='',cat='',timer=null;
function debounceLoad(){clearTimeout(timer);timer=setTimeout(load,200)}
function setFilter(f){filter=f;document.querySelectorAll('.fbtn').forEach((b,i)=>{if(i<3)b.classList.toggle('active',(!f&&i===0)||(f==='installed'&&i===1)||(f==='running'&&i===2))});load()}
function setCat(c){cat=c;document.querySelectorAll('.fbtn').forEach((b,i)=>{if(i>=3)b.classList.toggle('active',(!c&&i===3)||(c==='developer'&&i===4)||(c==='operations'&&i===5)||(c==='finance'&&i===6)||(c==='creator'&&i===7)||(c==='personal'&&i===8))});load()}

async function loadStats(){
  const r=await fetch('/api/stats');const d=await r.json();
  document.getElementById('stats').innerHTML=[
    {n:d.total,l:'Total tools'},{n:d.installed,l:'Installed'},{n:d.running,l:'Running'},{n:d.healthy,l:'Healthy'}
  ].map(s=>'<div class="stat"><div class="stat-n">'+s.n+'</div><div class="stat-l">'+s.l+'</div></div>').join('');
}

async function load(){
  let url='/api/tools?';
  const q=document.getElementById('search').value;
  if(q)url+='q='+encodeURIComponent(q)+'&';
  if(filter)url+='status='+filter+'&';
  if(cat)url+='category='+cat+'&';
  const r=await fetch(url);const d=await r.json();
  const tools=d.tools||[];
  document.getElementById('grid').innerHTML=tools.map(t=>{
    const actions=t.health==='not_installed'
      ?'<button class="act act-install" onclick="install(event,\''+t.slug+'\')">Install</button>'
      :t.running
        ?'<button class="act act-stop" onclick="stop(event,\''+t.slug+'\')">Stop</button><button class="act act-open" onclick="window.open(\'http://localhost:'+t.port+'/ui\')">Open :'+t.port+'</button>'
        :'<button class="act act-start" onclick="start(event,\''+t.slug+'\')">Start</button>';
    return '<div class="tool"><div class="tool-top"><div class="tool-name">'+t.name+'</div><div class="tool-status '+t.health+'">'+t.health.replace('_',' ')+'</div></div><div class="tool-tagline">'+t.tagline+'</div><div class="tool-meta"><span>:'+t.port+'</span><span>'+t.category+'</span><span>stockyard-'+t.slug+'</span></div><div class="tool-actions">'+actions+'</div></div>';
  }).join('');
  loadStats();
}

async function start(e,slug){e.target.disabled=true;e.target.textContent='Starting...';await fetch('/api/tools/'+slug+'/start',{method:'POST'});toast(slug+' started');setTimeout(load,1500)}
async function stop(e,slug){e.target.disabled=true;await fetch('/api/tools/'+slug+'/stop',{method:'POST'});toast(slug+' stopped');setTimeout(load,500)}
async function install(e,slug){e.target.disabled=true;e.target.textContent='Installing...';await fetch('/api/tools/'+slug+'/install',{method:'POST'});toast(slug+' installed');setTimeout(load,2000)}

function toast(msg){const t=document.getElementById('toast');t.textContent=msg;t.style.display='block';setTimeout(()=>t.style.display='none',3000)}

load()
</script></body></html>` 
