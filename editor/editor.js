/* =============================================================
   Kora Visual Editor — editor.js  (v3: with preview panel)
============================================================= */
'use strict';

// ------------------------------------------------------------------
// State
// ------------------------------------------------------------------
const state = {
  entities:  [],
  selected:  null,
  tool:      'select',
  cam:       { x: 0, y: 0, zoom: 1 },
  drag:      null,
  idSeq:     1,
  meta:      { name: 'Untitled', version: 1, logicalW: 360, logicalH: 640 },
  dirty:     false,
};

const LOGICAL_W = 360;
const LOGICAL_H = 640;

// ------------------------------------------------------------------
// DOM refs
// ------------------------------------------------------------------
const canvas      = document.getElementById('scene-canvas');
const ctx         = canvas.getContext('2d');
const hierarchyEl = document.getElementById('hierarchy-list');
const inspectorEl = document.getElementById('inspector-body');
const consoleBody = document.getElementById('console-body');
const coordsEl    = document.getElementById('vp-coords');
const zoomEl      = document.getElementById('vp-zoom');

// ------------------------------------------------------------------
// Toast
// ------------------------------------------------------------------
let toastTimer;
function showToast(msg) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { t.hidden = true; }, 2200);
}

// ------------------------------------------------------------------
// Console
// ------------------------------------------------------------------
function log(msg, type = 'info') {
  const line = document.createElement('div');
  line.className = `log-${type}`;
  const ts = new Date().toLocaleTimeString('pt-BR', { hour12: false });
  line.textContent = `[${ts}] ${msg}`;
  consoleBody.appendChild(line);
  consoleBody.scrollTop = consoleBody.scrollHeight;
}
document.getElementById('btn-clear-console').addEventListener('click', () => { consoleBody.innerHTML = ''; });

// ------------------------------------------------------------------
// Preview Panel init
// ------------------------------------------------------------------
const previewPanel = new PreviewPanel({
  container: document.getElementById('preview-container'),
  getScene: () => ({ entities: state.entities, meta: state.meta }),
  onLog: log,
});

// ------------------------------------------------------------------
// Canvas sizing
// ------------------------------------------------------------------
function resizeCanvas() {
  const rect = canvas.parentElement.getBoundingClientRect();
  canvas.width  = rect.width;
  canvas.height = rect.height - 36;
  render();
}
window.addEventListener('resize', resizeCanvas);

// ------------------------------------------------------------------
// World ↔ screen
// ------------------------------------------------------------------
function worldToScreen(wx, wy) {
  const cx = canvas.width  / 2 + state.cam.x * state.cam.zoom;
  const cy = canvas.height / 2 + state.cam.y * state.cam.zoom;
  return [cx + wx * state.cam.zoom, cy + wy * state.cam.zoom];
}
function screenToWorld(sx, sy) {
  const cx = canvas.width  / 2 + state.cam.x * state.cam.zoom;
  const cy = canvas.height / 2 + state.cam.y * state.cam.zoom;
  return [(sx - cx) / state.cam.zoom, (sy - cy) / state.cam.zoom];
}

// ------------------------------------------------------------------
// Render
// ------------------------------------------------------------------
function render() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  drawGrid();
  const [ox, oy] = worldToScreen(-LOGICAL_W / 2, -LOGICAL_H / 2);
  const pw = LOGICAL_W * state.cam.zoom, ph = LOGICAL_H * state.cam.zoom;
  ctx.save();
  ctx.strokeStyle = 'rgba(0,229,160,0.25)'; ctx.lineWidth = 1;
  ctx.strokeRect(ox, oy, pw, ph);
  ctx.fillStyle = 'rgba(0,0,0,0.3)'; ctx.fillRect(ox, oy, pw, ph);
  ctx.restore();
  for (const e of state.entities) { if (e.visible) drawEntity(e); }
  if (state.selected) { const e = getEntity(state.selected); if (e) drawSelection(e); }
  zoomEl.textContent = Math.round(state.cam.zoom * 100) + '%';
  document.title = (state.dirty ? '● ' : '') + state.meta.name + ' — Kora Editor';
}

function drawGrid() {
  const STEP = 32 * state.cam.zoom;
  const offX = (canvas.width  / 2 + state.cam.x * state.cam.zoom) % STEP;
  const offY = (canvas.height / 2 + state.cam.y * state.cam.zoom) % STEP;
  ctx.save(); ctx.strokeStyle = 'rgba(255,255,255,0.04)'; ctx.lineWidth = 1;
  for (let x = offX; x < canvas.width;  x += STEP) { ctx.beginPath(); ctx.moveTo(x,0); ctx.lineTo(x,canvas.height); ctx.stroke(); }
  for (let y = offY; y < canvas.height; y += STEP) { ctx.beginPath(); ctx.moveTo(0,y); ctx.lineTo(canvas.width,y);  ctx.stroke(); }
  ctx.restore();
}

function drawEntity(e) {
  const [sx, sy] = worldToScreen(e.x, e.y);
  const sw = e.w * state.cam.zoom, sh = e.h * state.cam.zoom;
  ctx.save(); ctx.translate(sx, sy); ctx.rotate(e.rotation * Math.PI / 180);
  ctx.fillStyle = e.color + '99'; ctx.fillRect(-sw/2,-sh/2,sw,sh);
  ctx.strokeStyle = e.color; ctx.lineWidth = 1.5; ctx.strokeRect(-sw/2,-sh/2,sw,sh);
  ctx.fillStyle = '#fff'; ctx.font = `${Math.max(9, 11*state.cam.zoom)}px Inter,sans-serif`;
  ctx.textAlign = 'center'; ctx.fillText(e.name, 0, -sh/2-4);
  ctx.restore();
}

function drawSelection(e) {
  const [sx, sy] = worldToScreen(e.x, e.y);
  const sw = e.w * state.cam.zoom, sh = e.h * state.cam.zoom;
  ctx.save(); ctx.translate(sx, sy); ctx.rotate(e.rotation * Math.PI / 180);
  ctx.strokeStyle = '#00e5a0'; ctx.lineWidth = 2; ctx.setLineDash([4,3]);
  ctx.strokeRect(-sw/2-2,-sh/2-2,sw+4,sh+4);
  ctx.setLineDash([]); ctx.fillStyle = '#00e5a0';
  for (const [hx,hy] of [[-sw/2-2,-sh/2-2],[0,-sh/2-2],[sw/2+2,-sh/2-2],[sw/2+2,0],[sw/2+2,sh/2+2],[0,sh/2+2],[-sw/2-2,sh/2+2],[-sw/2-2,0]]) {
    ctx.beginPath(); ctx.arc(hx,hy,3.5,0,Math.PI*2); ctx.fill();
  }
  ctx.restore();
}

// ------------------------------------------------------------------
// Entities
// ------------------------------------------------------------------
function getEntity(id) { return state.entities.find(e => e.id === id); }
function nextId() { return state.idSeq++; }
function createEntity(name, type) {
  return { id: nextId(), name, type, x:0, y:0, w:48, h:48, rotation:0, visible:true, locked:false, color: randomColor(), script:'' };
}
function randomColor() {
  return ['#00e5a0','#388bfd','#e3b341','#f85149','#bc8cff','#ff7b72','#79c0ff'][Math.floor(Math.random()*7)];
}

// ------------------------------------------------------------------
// Hierarchy
// ------------------------------------------------------------------
const ENTITY_ICONS = { sprite:'🟦', tilemap:'🔲', camera:'📷', audio:'🔊', custom:'⬡' };
function buildHierarchy() {
  hierarchyEl.innerHTML = '';
  for (const e of state.entities) {
    const li = document.createElement('li');
    li.className = 'hierarchy-item' + (e.id === state.selected ? ' selected' : '');
    li.dataset.id = e.id;
    li.innerHTML = `<span class="entity-icon">${ENTITY_ICONS[e.type]||'⬡'}</span><span class="entity-name">${e.name}</span><button class="entity-vis" title="Visibilidade">${e.visible?'👁':'🚫'}</button>`;
    li.addEventListener('click', ev => {
      if (ev.target.classList.contains('entity-vis')) { e.visible=!e.visible; markDirty(); buildHierarchy(); render(); return; }
      selectEntity(e.id);
    });
    hierarchyEl.appendChild(li);
  }
}

// ------------------------------------------------------------------
// Selection
// ------------------------------------------------------------------
function selectEntity(id) { state.selected=id; buildHierarchy(); buildInspector(getEntity(id)); render(); }
function deselectAll() { state.selected=null; buildHierarchy(); inspectorEl.innerHTML='<p class="inspector-empty">Selecione uma entidade</p>'; render(); }

// ------------------------------------------------------------------
// Dirty flag
// ------------------------------------------------------------------
function markDirty() { state.dirty = true; render(); }
function clearDirty() { state.dirty = false; render(); }

// ------------------------------------------------------------------
// Inspector
// ------------------------------------------------------------------
function buildInspector(e) {
  if (!e) { inspectorEl.innerHTML='<p class="inspector-empty">Selecione uma entidade</p>'; return; }
  inspectorEl.innerHTML = `
    <div class="inspector-section">
      <div class="inspector-section-title">Identidade</div>
      <div class="prop-row"><span class="prop-label">Nome</span><input class="prop-input" data-prop="name" value="${e.name}"></div>
      <div class="prop-row"><span class="prop-label">Tipo</span><input class="prop-input" value="${e.type}" disabled></div>
    </div>
    <div class="inspector-section">
      <div class="inspector-section-title">Transform</div>
      <div class="prop-row"><span class="prop-label">X</span><input class="prop-input" type="number" data-prop="x" value="${e.x}"></div>
      <div class="prop-row"><span class="prop-label">Y</span><input class="prop-input" type="number" data-prop="y" value="${e.y}"></div>
      <div class="prop-row"><span class="prop-label">Largura</span><input class="prop-input" type="number" data-prop="w" value="${e.w}"></div>
      <div class="prop-row"><span class="prop-label">Altura</span><input class="prop-input" type="number" data-prop="h" value="${e.h}"></div>
      <div class="prop-row"><span class="prop-label">Rotação</span><input class="prop-input" type="number" data-prop="rotation" value="${e.rotation}"></div>
    </div>
    <div class="inspector-section">
      <div class="inspector-section-title">Visual</div>
      <div class="prop-row"><span class="prop-label">Cor</span><input class="prop-color" type="color" data-prop="color" value="${e.color}"></div>
      <div class="prop-row"><span class="prop-label">Visível</span><input class="prop-checkbox" type="checkbox" data-prop="visible" ${e.visible?'checked':''}></div>
      <div class="prop-row"><span class="prop-label">Travado</span><input class="prop-checkbox" type="checkbox" data-prop="locked" ${e.locked?'checked':''}></div>
    </div>
    <div class="inspector-section">
      <div class="inspector-section-title">Script KScript</div>
      <textarea class="script-area" data-prop="script" placeholder="on Update(dt) {\n  // lógica aqui\n}">${e.script||''}</textarea>
    </div>
    <div class="inspector-section">
      <div class="inspector-section-title">Ações</div>
      <button class="btn-secondary" id="btn-duplicate" style="width:100%;margin-bottom:4px">Duplicar</button>
      <button class="btn-secondary" id="btn-delete" style="width:100%;color:#f85149;border-color:#f85149">Excluir</button>
    </div>
  `;
  inspectorEl.querySelectorAll('[data-prop]').forEach(input => {
    input.addEventListener('input', () => {
      const prop = input.dataset.prop;
      if (input.type==='checkbox') e[prop]=input.checked;
      else if (input.type==='number') e[prop]=parseFloat(input.value)||0;
      else e[prop]=input.value;
      if (prop==='name') buildHierarchy();
      markDirty(); render();
    });
  });
  inspectorEl.querySelector('#btn-duplicate')?.addEventListener('click', () => {
    const copy={...e,id:nextId(),name:e.name+'_copy',x:e.x+16,y:e.y+16};
    state.entities.push(copy); selectEntity(copy.id); buildHierarchy(); markDirty();
    log(`Duplicado: ${copy.name}`,'ok');
  });
  inspectorEl.querySelector('#btn-delete')?.addEventListener('click', () => {
    const idx=state.entities.findIndex(en=>en.id===e.id);
    if (idx>=0) state.entities.splice(idx,1);
    deselectAll(); markDirty(); log(`Excluído: ${e.name}`,'warn');
  });
}

// ------------------------------------------------------------------
// Serialization actions
// ------------------------------------------------------------------
function actionNew() {
  if (state.dirty && !confirm('Há alterações não salvas. Criar nova cena?')) return;
  state.entities=[]; state.selected=null; state.idSeq=1;
  state.meta={ name:'Untitled', version:1, logicalW:360, logicalH:640 };
  buildHierarchy(); deselectAll(); clearDirty();
  log('Nova cena criada.','ok');
}
function actionSave() {
  const doc = Serializer.sceneToJSON(state.entities, state.meta);
  Serializer.downloadJSON(doc, `${state.meta.name}.kora.json`);
  clearDirty(); showToast('✅ Cena salva como JSON');
  log(`Cena salva: ${state.meta.name}.kora.json`,'ok');
}
async function actionOpen() {
  try {
    const doc = await Serializer.openSceneFile();
    const { entities, meta } = Serializer.jsonToScene(doc, nextId);
    if (state.dirty && !confirm('Há alterações não salvas. Abrir mesmo assim?')) return;
    state.entities = entities; state.meta = meta; state.selected = null;
    buildHierarchy(); deselectAll(); clearDirty();
    log(`Cena carregada: ${meta.name} (${entities.length} entidades)`,'ok');
    showToast(`📂 ${meta.name} carregado`);
  } catch (err) { log(err.message, 'error'); }
}
function actionExportKScript() {
  const src = Serializer.sceneToKScript(state.entities, state.meta.name);
  Serializer.downloadKScript(src, `${state.meta.name}.ks`);
  showToast('📄 KScript exportado');
  log(`KScript exportado: ${state.meta.name}.ks`,'ok');
}

document.getElementById('btn-new').addEventListener('click', actionNew);
document.getElementById('btn-save').addEventListener('click', actionSave);
document.getElementById('btn-open').addEventListener('click', actionOpen);
document.getElementById('btn-export-ks').addEventListener('click', actionExportKScript);

// Play button → switch to preview tab AND run
document.getElementById('btn-play').addEventListener('click', () => {
  // Switch to preview tab
  document.querySelectorAll('.tb-btn[data-tab]').forEach(b => b.classList.remove('active'));
  document.querySelector('[data-tab="preview"]').classList.add('active');
  switchTab('preview');
  setTimeout(() => previewPanel.play(), 80);
});

document.getElementById('btn-export').addEventListener('click', () => {
  log('Gerando APK... requer Android SDK + gomobile','warn');
  setTimeout(()=>log('Execute: ./android/build.sh debug','info'),500);
});

// ------------------------------------------------------------------
// Tab system
// ------------------------------------------------------------------
function switchTab(tab) {
  document.getElementById('tab-scene').style.display   = tab === 'scene'   ? 'flex' : 'none';
  document.getElementById('tab-preview').style.display = tab === 'preview' ? 'flex' : 'none';
  if (tab === 'scene') resizeCanvas();
}

document.querySelectorAll('.tb-btn[data-tab]').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.tb-btn[data-tab]').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    switchTab(btn.dataset.tab);
  });
});

// ------------------------------------------------------------------
// Canvas interaction
// ------------------------------------------------------------------
canvas.addEventListener('mousemove', ev => {
  const r=canvas.getBoundingClientRect();
  const [wx,wy]=screenToWorld(ev.clientX-r.left, ev.clientY-r.top);
  coordsEl.textContent=`x: ${Math.round(wx)}  y: ${Math.round(wy)}`;
  if (state.drag) {
    const e=getEntity(state.drag.entityId);
    if (e&&!e.locked) { e.x=Math.round(wx-state.drag.ox); e.y=Math.round(wy-state.drag.oy); buildInspector(e); markDirty(); render(); }
  }
});
canvas.addEventListener('mousedown', ev => {
  const r=canvas.getBoundingClientRect();
  const [wx,wy]=screenToWorld(ev.clientX-r.left, ev.clientY-r.top);
  const hit=[...state.entities].reverse().find(e=>{
    if (!e.visible||e.locked) return false;
    const hw=e.w/2,hh=e.h/2;
    return wx>=e.x-hw&&wx<=e.x+hw&&wy>=e.y-hh&&wy<=e.y+hh;
  });
  if (hit) { selectEntity(hit.id); if (state.tool==='select'||state.tool==='move') state.drag={entityId:hit.id,ox:wx-hit.x,oy:wy-hit.y}; }
  else deselectAll();
});
canvas.addEventListener('mouseup',    () => { state.drag=null; });
canvas.addEventListener('mouseleave', () => { state.drag=null; });
canvas.addEventListener('wheel', ev => {
  ev.preventDefault();
  state.cam.zoom=Math.min(4,Math.max(0.1,state.cam.zoom*(ev.deltaY>0?0.9:1.1)));
  render();
}, { passive:false });

// ------------------------------------------------------------------
// Toolbar
// ------------------------------------------------------------------
document.querySelectorAll('.vp-tool[data-tool]').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.vp-tool[data-tool]').forEach(b=>b.classList.remove('active'));
    btn.classList.add('active'); state.tool=btn.dataset.tool;
  });
});
document.getElementById('btn-zoom-fit').addEventListener('click', () => {
  state.cam={x:0,y:0,zoom:Math.min(canvas.width/LOGICAL_W,canvas.height/LOGICAL_H)*0.85};
  render();
});

// ------------------------------------------------------------------
// Add entity modal
// ------------------------------------------------------------------
const backdrop=document.getElementById('modal-backdrop');
document.getElementById('btn-add-entity').addEventListener('click',()=>{ backdrop.hidden=false; document.getElementById('new-entity-name').focus(); });
document.querySelectorAll('.modal-close').forEach(btn=>btn.addEventListener('click',()=>{ backdrop.hidden=true; }));
backdrop.addEventListener('click',ev=>{ if(ev.target===backdrop) backdrop.hidden=true; });
document.getElementById('confirm-add-entity').addEventListener('click',()=>{
  const name=document.getElementById('new-entity-name').value.trim()||'Entity';
  const type=document.getElementById('new-entity-type').value;
  const e=createEntity(name,type);
  state.entities.push(e); selectEntity(e.id); buildHierarchy(); backdrop.hidden=true;
  document.getElementById('new-entity-name').value='';
  markDirty(); log(`Adicionado: ${name} (${type})`,'ok');
});
document.getElementById('new-entity-name').addEventListener('keydown',ev=>{ if(ev.key==='Enter') document.getElementById('confirm-add-entity').click(); });

// ------------------------------------------------------------------
// Theme toggle
// ------------------------------------------------------------------
(function(){
  const btn=document.querySelector('[data-theme-toggle]'), root=document.documentElement;
  let dark=root.getAttribute('data-theme')!=='light';
  function apply(){ root.setAttribute('data-theme',dark?'dark':'light'); btn.innerHTML=dark?'<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>':'<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>'; }
  btn.addEventListener('click',()=>{ dark=!dark; apply(); }); apply();
})();

// ------------------------------------------------------------------
// Keyboard shortcuts
// ------------------------------------------------------------------
window.addEventListener('keydown', ev => {
  const inInput = ev.target.tagName==='INPUT'||ev.target.tagName==='TEXTAREA'||ev.target.tagName==='SELECT';
  if (ev.ctrlKey || ev.metaKey) {
    if (ev.key==='s') { ev.preventDefault(); actionSave(); return; }
    if (ev.key==='o') { ev.preventDefault(); actionOpen(); return; }
    if (ev.key==='n') { ev.preventDefault(); actionNew();  return; }
  }
  if (inInput) return;
  switch(ev.key) {
    case 'v': case 'V': document.querySelector('[data-tool="select"]').click(); break;
    case 'g': case 'G': document.querySelector('[data-tool="move"]').click();   break;
    case 's': case 'S': document.querySelector('[data-tool="scale"]').click();  break;
    case 'f': case 'F': document.getElementById('btn-zoom-fit').click();        break;
    case 'Delete': case 'Backspace':
      if (state.selected) {
        const e=getEntity(state.selected), idx=state.entities.findIndex(en=>en.id===state.selected);
        if (idx>=0){ state.entities.splice(idx,1); deselectAll(); markDirty(); log(`Excluído: ${e?.name}`,'warn'); }
      }
      break;
    case 'F5': ev.preventDefault(); document.getElementById('btn-play').click(); break;
  }
});

window.addEventListener('beforeunload', ev => { if (state.dirty) { ev.preventDefault(); ev.returnValue=''; } });

// ------------------------------------------------------------------
// Seed scene
// ------------------------------------------------------------------
function seedScene() {
  const player=createEntity('Player','sprite');   player.x=0;  player.y=60;  player.color='#00e5a0';
  const ground=createEntity('Ground','tilemap');  ground.x=0;  ground.y=240; ground.w=340; ground.h=24; ground.color='#388bfd';
  const cam=createEntity('MainCamera','camera');  cam.x=0; cam.y=0; cam.w=32; cam.h=32; cam.color='#e3b341';
  state.entities.push(player,ground,cam);
  buildHierarchy(); log('Cena de exemplo carregada.','ok');
}

// ------------------------------------------------------------------
// Init
// ------------------------------------------------------------------
resizeCanvas();
state.cam.zoom=0.75;
seedScene();
render();
log('Kora Editor v3 iniciado. F5 = Rodar Preview · Ctrl+S salvar · Ctrl+O abrir.','ok');
