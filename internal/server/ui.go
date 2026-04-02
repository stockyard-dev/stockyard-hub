package server

import "net/http"

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Stockyard Hub</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rust-light:#e8753a;--leather:#a0845c;--leather-light:#c4a87a;--cream:#f0e6d3;--cream-dim:#bfb5a3;--cream-muted:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c44040;--amber:#c9892a;--font-mono:'JetBrains Mono',monospace;--font-serif:'Libre Baskerville',Georgia,serif}
*{margin:0;padding:0;box-sizing:border-box}
html,body{height:100%;overflow:hidden}
body{background:var(--bg);color:var(--cream);font-family:var(--font-mono);font-size:13px;line-height:1.6}
a{color:var(--rust-light);text-decoration:none}a:hover{color:var(--gold)}
.app{display:flex;flex-direction:column;height:100vh}
.header{padding:.6rem 1rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center;flex-shrink:0}
.header h1{font-family:var(--font-serif);font-size:1rem;color:var(--cream)}.header h1 span{color:var(--rust-light)}
.header-stats{display:flex;gap:1.2rem;font-size:.7rem;color:var(--leather)}.header-stats .val{color:var(--cream);font-weight:600}
.license-bar{padding:.4rem 1rem;border-bottom:1px solid var(--bg3);display:flex;align-items:center;gap:.5rem;font-size:.7rem;flex-shrink:0}
.license-bar input{background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);padding:.2rem .5rem;font-family:var(--font-mono);font-size:.7rem;flex:1;max-width:300px;outline:none}
.license-bar input:focus{border-color:var(--gold)}
.license-bar button{background:var(--gold);color:var(--bg);border:none;padding:.2rem .6rem;font-family:var(--font-mono);font-size:.65rem;cursor:pointer}
.ls{font-size:.65rem}.ls.set{color:var(--green)}.ls.unset{color:var(--amber)}
.main{display:flex;flex:1;overflow:hidden}
.sidebar{width:300px;min-width:300px;border-right:1px solid var(--bg3);display:flex;flex-direction:column;overflow:hidden;flex-shrink:0}
.topbar{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);display:flex;gap:.4rem;align-items:center;flex-shrink:0}
.search{background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);padding:.25rem .5rem;font-family:var(--font-mono);font-size:.7rem;flex:1;min-width:80px;outline:none}
.search:focus{border-color:var(--rust)}
.fb{background:var(--bg2);border:1px solid var(--bg3);color:var(--cream-dim);padding:.15rem .4rem;font-family:var(--font-mono);font-size:.6rem;cursor:pointer}
.fb:hover{border-color:var(--cream-muted);color:var(--cream)}.fb.active{border-color:var(--rust);color:var(--rust-light);background:var(--bg)}
.sidebar-list{flex:1;overflow-y:auto}
.ti{padding:.5rem .8rem;border-bottom:1px solid var(--bg3);transition:background .1s}
.ti:hover{background:var(--bg2)}.ti.active{background:var(--bg2);border-left:2px solid var(--gold)}
.ti-top{display:flex;justify-content:space-between;align-items:center}
.ti-name{font-size:.75rem;font-weight:600}.ti-sub{font-size:.63rem;color:var(--cream-dim);font-style:italic;font-family:var(--font-serif);margin-top:1px}
.ti-meta{font-size:.58rem;color:var(--leather);margin-top:1px}
.badge{font-size:.5rem;padding:1px 3px;letter-spacing:.5px;text-transform:uppercase}
.b-h{background:rgba(74,158,92,.15);color:var(--green);border:1px solid rgba(74,158,92,.3)}
.b-s{color:var(--cream-muted);border:1px solid var(--bg3)}.b-u{background:rgba(196,64,64,.1);color:var(--red);border:1px solid rgba(196,64,64,.3)}
.b-n{color:var(--cream-muted);border:1px solid var(--bg3)}
.ti-btns{display:flex;gap:3px;margin-top:3px}
.ti-btns button{font-family:var(--font-mono);font-size:.58rem;padding:1px 4px;cursor:pointer;border:1px solid;background:transparent}
.bi{border-color:var(--rust);color:var(--rust-light)}.bi:hover{background:var(--rust);color:var(--cream)}
.bs{border-color:var(--green);color:var(--green)}.bs:hover{background:var(--green);color:var(--bg)}
.bx{border-color:var(--red);color:var(--red)}.bx:hover{background:var(--red);color:var(--cream)}
.bo{border-color:var(--gold);color:var(--gold)}.bo:hover{background:var(--gold);color:var(--bg)}
.br{border-color:var(--bg3);color:var(--cream-muted)}.br:hover{border-color:var(--red);color:var(--red)}
.ld{color:var(--amber);font-size:.58rem}
.panel{flex:1;display:flex;flex-direction:column;background:var(--bg)}
.panel-empty{flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;color:var(--cream-muted);gap:10px}
.panel-empty-t{font-family:var(--font-serif);font-size:1rem;font-style:italic;color:var(--leather)}
.panel-empty-s{font-size:.7rem;max-width:260px;text-align:center;line-height:1.7}
.panel-hdr{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center;flex-shrink:0;background:var(--bg2)}
.panel-hdr .name{font-size:.78rem;font-weight:600}.panel-hdr .port{color:var(--gold);font-size:.68rem;margin-left:6px;font-weight:400}
.panel-hdr button{background:none;border:1px solid var(--bg3);color:var(--cream-muted);padding:1px 6px;font-family:var(--font-mono);font-size:.6rem;cursor:pointer;margin-left:4px}
.panel-hdr button:hover{border-color:var(--cream);color:var(--cream)}
.panel-frame{flex:1;border:none;width:100%;height:100%}
.empty{padding:1.5rem;text-align:center;color:var(--cream-muted);font-style:italic;font-family:var(--font-serif);font-size:.75rem}
</style>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
</head>
<body>
<div class="app">
<div class="header">
  <h1><span>Stockyard</span> Hub</h1>
  <div class="header-stats">
    <span>Catalog: <span class="val" id="sc">-</span></span>
    <span>Installed: <span class="val" id="si">-</span></span>
    <span>Running: <span class="val" id="sr">-</span></span>
  </div>
</div>
<div class="license-bar">
  <span style="color:var(--leather)">License:</span>
  <input type="text" id="li" placeholder="Paste Complete license key">
  <button onclick="saveLicense()">Save</button>
  <span id="lst" class="ls unset">not set</span>
</div>
<div class="main">
<div class="sidebar">
  <div class="topbar">
    <input type="text" class="search" id="search" placeholder="Search..." oninput="renderTools()">
    <button class="fb active" onclick="setFilter('all',this)">All</button>
    <button class="fb" onclick="setFilter('installed',this)">Inst</button>
    <button class="fb" onclick="setFilter('running',this)">Run</button>
  </div>
  <div class="sidebar-list" id="tl"></div>
</div>
<div class="panel" id="panel">
  <div class="panel-empty"><div class="panel-empty-t">Select a tool</div><div class="panel-empty-s">Click Open on a running tool to view its dashboard here.</div></div>
</div>
</div>
</div>
<script>
let allTools=[],currentFilter='all',activeSlug=null;
async function loadTools(){try{const r=await fetch('/api/tools');const d=await r.json();allTools=d.tools||[];renderTools();document.getElementById('sc').textContent=allTools.length;document.getElementById('si').textContent=allTools.filter(t=>t.installed).length;document.getElementById('sr').textContent=allTools.filter(t=>t.running).length}catch(e){}}
function setFilter(f,btn){currentFilter=f;document.querySelectorAll('.fb').forEach(b=>b.classList.remove('active'));btn.classList.add('active');renderTools()}
function renderTools(){const list=document.getElementById('tl');const q=document.getElementById('search').value.toLowerCase();let items=allTools.filter(t=>{if(q&&!(t.slug+t.name+t.tagline).toLowerCase().includes(q))return false;if(currentFilter==='installed')return t.installed;if(currentFilter==='running')return t.running;return true});if(!items.length){list.innerHTML='<div class="empty">No tools match.</div>';return}
list.innerHTML=items.map(t=>{const bc=t.health==='healthy'?'b-h':t.health==='unhealthy'?'b-u':t.health==='stopped'?'b-s':'b-n';const bl=t.health==='healthy'?'healthy':t.health==='unhealthy'?'unhealthy':t.health==='stopped'?'stopped':'not installed';const ac=activeSlug===t.slug?' active':'';let btns='';if(!t.installed){btns='<button class="bi" onclick="installTool(\''+t.slug+'\')">Install</button>'}else if(t.running){btns='<button class="bo" onclick="openTool(\''+t.slug+'\','+t.port+')">Open</button><button class="bx" onclick="stopTool(\''+t.slug+'\')">Stop</button>'}else{btns='<button class="bs" onclick="startTool(\''+t.slug+'\')">Start</button><button class="br" onclick="uninstallTool(\''+t.slug+'\')">Remove</button>'}
return'<div class="ti'+ac+'" id="ti-'+t.slug+'"><div class="ti-top"><span class="ti-name">'+t.name+'</span><span class="badge '+bc+'">'+bl+'</span></div><div class="ti-sub">'+t.tagline+'</div><div class="ti-meta">:'+t.port+' · '+t.category+'</div><div class="ti-btns" id="tb-'+t.slug+'">'+btns+'</div></div>'}).join('')}
function openTool(slug,port){activeSlug=slug;renderTools();document.getElementById('panel').innerHTML='<div class="panel-hdr"><div><span class="name">'+allTools.find(t=>t.slug===slug).name+'</span><span class="port">:'+port+'</span></div><div><button onclick="window.open(\'http://localhost:'+port+'/ui\',\'_blank\')">New tab &#8599;</button><button onclick="closePanel()">Close &#10005;</button></div></div><iframe class="panel-frame" src="http://localhost:'+port+'/ui"></iframe>'}
function closePanel(){activeSlug=null;renderTools();document.getElementById('panel').innerHTML='<div class="panel-empty"><div class="panel-empty-t">Select a tool</div><div class="panel-empty-s">Click Open on a running tool to view its dashboard here.</div></div>'}
async function installTool(slug){setLoading(slug,'Installing...');try{const r=await fetch('/api/tools/'+slug+'/install',{method:'POST'});const d=await r.json();if(d.error)alert('Error: '+d.error)}catch(e){alert('Failed: '+e)}await loadTools()}
async function startTool(slug){setLoading(slug,'Starting...');try{const r=await fetch('/api/tools/'+slug+'/start',{method:'POST'});const d=await r.json();if(d.error)alert('Error: '+d.error)}catch(e){alert('Failed: '+e)}setTimeout(loadTools,1000)}
async function stopTool(slug){setLoading(slug,'Stopping...');try{await fetch('/api/tools/'+slug+'/stop',{method:'POST'})}catch(e){alert('Failed: '+e)}if(activeSlug===slug)closePanel();await loadTools()}
async function uninstallTool(slug){if(!confirm('Remove '+slug+'? Data preserved.'))return;setLoading(slug,'Removing...');try{await fetch('/api/tools/'+slug+'/uninstall',{method:'POST'})}catch(e){alert('Failed: '+e)}if(activeSlug===slug)closePanel();await loadTools()}
function setLoading(slug,msg){const el=document.getElementById('tb-'+slug);if(el)el.innerHTML='<span class="ld">'+msg+'</span>'}
async function saveLicense(){const key=document.getElementById('li').value.trim();if(!key)return;try{await fetch('/api/config/license',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({key:key})});document.getElementById('lst').textContent='active';document.getElementById('lst').className='ls set';document.getElementById('li').value=''}catch(e){alert('Failed: '+e)}}
async function checkLicense(){try{const r=await fetch('/api/config');const d=await r.json();if(d.license_key_set){document.getElementById('lst').textContent='active';document.getElementById('lst').className='ls set'}}catch(e){}}
loadTools();checkLicense();setInterval(loadTools,15000)
</script>
</body>
</html>` + "`"

// closing backtick above ends the raw string literal
