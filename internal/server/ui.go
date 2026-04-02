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
body{background:var(--bg);color:var(--cream);font-family:var(--font-mono);font-size:13px;line-height:1.6}
a{color:var(--rust-light);text-decoration:none}a:hover{color:var(--gold)}

.header{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.header h1{font-family:var(--font-serif);font-size:1.1rem;color:var(--cream)}
.header h1 span{color:var(--rust-light)}
.header-stats{display:flex;gap:1.5rem;font-size:.75rem;color:var(--leather)}
.header-stats .val{color:var(--cream);font-weight:600}

.controls{padding:.8rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;gap:.8rem;align-items:center;flex-wrap:wrap}
.search{background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .8rem;font-family:var(--font-mono);font-size:.8rem;flex:1;min-width:200px;outline:none}
.search:focus{border-color:var(--rust)}
.filter-btn{background:var(--bg2);border:1px solid var(--bg3);color:var(--cream-dim);padding:.3rem .7rem;font-family:var(--font-mono);font-size:.7rem;cursor:pointer;transition:all .15s}
.filter-btn:hover{border-color:var(--cream-muted);color:var(--cream)}
.filter-btn.active{border-color:var(--rust);color:var(--rust-light);background:var(--bg)}

.license-bar{padding:.6rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;align-items:center;gap:.8rem;font-size:.75rem}
.license-bar input{background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);padding:.3rem .6rem;font-family:var(--font-mono);font-size:.75rem;flex:1;max-width:400px;outline:none}
.license-bar input:focus{border-color:var(--gold)}
.license-bar button{background:var(--gold);color:var(--bg);border:none;padding:.3rem .8rem;font-family:var(--font-mono);font-size:.7rem;cursor:pointer}
.license-status{font-size:.7rem}
.license-status.set{color:var(--green)}
.license-status.unset{color:var(--amber)}

.tool-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(320px,1fr));gap:1px;background:var(--bg3);padding:0}
.tool-card{background:var(--bg);padding:1rem 1.2rem;display:flex;flex-direction:column;gap:.5rem}
.tool-top{display:flex;justify-content:space-between;align-items:flex-start}
.tool-name{font-size:.85rem;font-weight:600;color:var(--cream)}
.tool-name a{color:var(--cream)}
.tool-name a:hover{color:var(--rust-light)}
.tool-badge{font-size:.6rem;padding:.15rem .4rem;letter-spacing:1px;text-transform:uppercase;border-radius:2px}
.badge-healthy{background:rgba(74,158,92,.15);color:var(--green);border:1px solid rgba(74,158,92,.3)}
.badge-stopped{background:rgba(122,112,96,.1);color:var(--cream-muted);border:1px solid var(--bg3)}
.badge-unhealthy{background:rgba(196,64,64,.1);color:var(--red);border:1px solid rgba(196,64,64,.3)}
.badge-not-installed{background:transparent;color:var(--cream-muted);border:1px solid var(--bg3)}
.tool-tagline{font-size:.75rem;color:var(--cream-dim);font-style:italic;font-family:var(--font-serif)}
.tool-meta{font-size:.65rem;color:var(--leather);display:flex;gap:1rem}
.tool-actions{display:flex;gap:.4rem;margin-top:.3rem}
.tool-actions button{font-family:var(--font-mono);font-size:.65rem;padding:.2rem .5rem;cursor:pointer;border:1px solid;transition:all .15s}
.btn-install{background:transparent;border-color:var(--rust);color:var(--rust-light)}.btn-install:hover{background:var(--rust);color:var(--cream)}
.btn-start{background:transparent;border-color:var(--green);color:var(--green)}.btn-start:hover{background:var(--green);color:var(--bg)}
.btn-stop{background:transparent;border-color:var(--red);color:var(--red)}.btn-stop:hover{background:var(--red);color:var(--cream)}
.btn-open{background:transparent;border-color:var(--gold);color:var(--gold)}.btn-open:hover{background:var(--gold);color:var(--bg)}
.btn-remove{background:transparent;border-color:var(--bg3);color:var(--cream-muted)}.btn-remove:hover{border-color:var(--red);color:var(--red)}
.btn-disabled{opacity:.4;pointer-events:none}
.loading{color:var(--amber);font-size:.65rem}
.empty{padding:3rem;text-align:center;color:var(--cream-muted);font-style:italic;font-family:var(--font-serif)}
</style>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
</head>
<body>
<div class="header">
  <h1><span>Stockyard</span> Hub</h1>
  <div class="header-stats">
    <span>Catalog: <span class="val" id="stat-catalog">-</span></span>
    <span>Installed: <span class="val" id="stat-installed">-</span></span>
    <span>Running: <span class="val" id="stat-running">-</span></span>
  </div>
</div>

<div class="license-bar">
  <span style="color:var(--leather)">License:</span>
  <input type="text" id="license-input" placeholder="Paste your Complete license key here">
  <button onclick="saveLicense()">Save</button>
  <span id="license-status" class="license-status unset">not set</span>
</div>

<div class="controls">
  <input type="text" class="search" id="search" placeholder="Search tools..." oninput="filterTools()">
  <button class="filter-btn active" onclick="setFilter('all',this)">All</button>
  <button class="filter-btn" onclick="setFilter('installed',this)">Installed</button>
  <button class="filter-btn" onclick="setFilter('running',this)">Running</button>
  <button class="filter-btn" onclick="setFilter('developer',this)">Developer</button>
  <button class="filter-btn" onclick="setFilter('operations',this)">Operations</button>
  <button class="filter-btn" onclick="setFilter('creator',this)">Creator</button>
  <button class="filter-btn" onclick="setFilter('finance',this)">Finance</button>
  <button class="filter-btn" onclick="setFilter('personal',this)">Personal</button>
</div>

<div class="tool-grid" id="tool-grid"></div>

<script>
let allTools = [];
let currentFilter = 'all';

async function loadTools() {
  try {
    const r = await fetch('/api/tools');
    const d = await r.json();
    allTools = d.tools || [];
    renderTools();
    updateStats();
  } catch(e) { console.error('load:', e); }
}

function updateStats() {
  const installed = allTools.filter(t => t.installed).length;
  const running = allTools.filter(t => t.running).length;
  document.getElementById('stat-catalog').textContent = allTools.length;
  document.getElementById('stat-installed').textContent = installed;
  document.getElementById('stat-running').textContent = running;
}

function filterTools() {
  renderTools();
}

function setFilter(f, btn) {
  currentFilter = f;
  document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');
  renderTools();
}

function renderTools() {
  const grid = document.getElementById('tool-grid');
  const search = document.getElementById('search').value.toLowerCase();

  let filtered = allTools.filter(t => {
    if (search && !(t.slug+t.name+t.tagline).toLowerCase().includes(search)) return false;
    if (currentFilter === 'installed') return t.installed;
    if (currentFilter === 'running') return t.running;
    if (['developer','operations','creator','finance','personal'].includes(currentFilter)) return t.category === currentFilter;
    return true;
  });

  if (filtered.length === 0) {
    grid.innerHTML = '<div class="empty">No tools match your filter.</div>';
    return;
  }

  grid.innerHTML = filtered.map(t => {
    const badge = t.health === 'healthy' ? '<span class="tool-badge badge-healthy">healthy</span>'
      : t.health === 'unhealthy' ? '<span class="tool-badge badge-unhealthy">unhealthy</span>'
      : t.health === 'stopped' ? '<span class="tool-badge badge-stopped">stopped</span>'
      : '<span class="tool-badge badge-not-installed">not installed</span>';

    let actions = '';
    if (!t.installed) {
      actions = '<button class="btn-install" onclick="installTool(\''+t.slug+'\')">Install</button>';
    } else if (t.running) {
      actions = '<button class="btn-open" onclick="openDashboard('+t.port+')">Open :'+t.port+'</button>' +
        '<button class="btn-stop" onclick="stopTool(\''+t.slug+'\')">Stop</button>';
    } else {
      actions = '<button class="btn-start" onclick="startTool(\''+t.slug+'\')">Start</button>' +
        '<button class="btn-remove" onclick="uninstallTool(\''+t.slug+'\')">Remove</button>';
    }

    return '<div class="tool-card" id="card-'+t.slug+'">' +
      '<div class="tool-top"><span class="tool-name"><a href="https://stockyard.dev/'+t.slug+'/" target="_blank">'+t.name+'</a></span>'+badge+'</div>' +
      '<div class="tool-tagline">'+t.tagline+'</div>' +
      '<div class="tool-meta"><span>:'+t.port+'</span><span>'+t.category+'</span></div>' +
      '<div class="tool-actions" id="actions-'+t.slug+'">'+actions+'</div>' +
      '</div>';
  }).join('');
}

async function installTool(slug) {
  setLoading(slug, 'Installing...');
  try {
    const r = await fetch('/api/tools/'+slug+'/install', {method:'POST'});
    const d = await r.json();
    if (d.error) { alert('Install error: '+d.error); }
  } catch(e) { alert('Install failed: '+e); }
  await loadTools();
}

async function startTool(slug) {
  setLoading(slug, 'Starting...');
  try {
    const r = await fetch('/api/tools/'+slug+'/start', {method:'POST'});
    const d = await r.json();
    if (d.error) { alert('Start error: '+d.error); }
  } catch(e) { alert('Start failed: '+e); }
  setTimeout(loadTools, 1000); // give it a second to bind port
}

async function stopTool(slug) {
  setLoading(slug, 'Stopping...');
  try {
    await fetch('/api/tools/'+slug+'/stop', {method:'POST'});
  } catch(e) { alert('Stop failed: '+e); }
  await loadTools();
}

async function uninstallTool(slug) {
  if (!confirm('Remove '+slug+'? Data in the tool data directory will be preserved.')) return;
  setLoading(slug, 'Removing...');
  try {
    await fetch('/api/tools/'+slug+'/uninstall', {method:'POST'});
  } catch(e) { alert('Remove failed: '+e); }
  await loadTools();
}

function openDashboard(port) {
  window.open('http://localhost:'+port+'/ui', '_blank');
}

function setLoading(slug, msg) {
  const el = document.getElementById('actions-'+slug);
  if (el) el.innerHTML = '<span class="loading">'+msg+'</span>';
}

async function saveLicense() {
  const key = document.getElementById('license-input').value.trim();
  if (!key) return;
  try {
    await fetch('/api/config/license', {
      method: 'POST',
      headers: {'Content-Type':'application/json'},
      body: JSON.stringify({key: key})
    });
    document.getElementById('license-status').textContent = 'saved';
    document.getElementById('license-status').className = 'license-status set';
    document.getElementById('license-input').value = '';
  } catch(e) { alert('Save failed: '+e); }
}

async function checkLicense() {
  try {
    const r = await fetch('/api/config');
    const d = await r.json();
    const el = document.getElementById('license-status');
    if (d.license_key_set) {
      el.textContent = 'active';
      el.className = 'license-status set';
    }
  } catch(e) {}
}

// Init
loadTools();
checkLicense();
setInterval(loadTools, 15000); // refresh every 15s
</script>
</body>
</html>`
