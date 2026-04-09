/* =============================================================
   Kora Editor — assets-panel.js
   Manages the Assets panel: import files, thumbnail preview,
   drag-and-drop onto the scene canvas to spawn entities.
============================================================= */
'use strict';

const ASSET_TYPES = {
  image: { exts: ['png','jpg','jpeg','gif','webp','svg'], icon: '🖼', entityType: 'sprite' },
  audio: { exts: ['mp3','ogg','wav','flac'],             icon: '🔊', entityType: 'audio'  },
  tilemap: { exts: ['tmj','tmx','json'],                 icon: '🗺',  entityType: 'tilemap'},
  script: { exts: ['ks'],                               icon: '📜', entityType: 'custom' },
};

function extOf(name) { return name.split('.').pop().toLowerCase(); }
function typeOf(name) {
  const e = extOf(name);
  for (const [t, info] of Object.entries(ASSET_TYPES)) {
    if (info.exts.includes(e)) return t;
  }
  return 'other';
}

class AssetsPanel {
  /**
   * @param {object} opts
   * @param {HTMLElement} opts.container
   * @param {function}    opts.onSpawn   – (asset) => void  called when user drops asset onto canvas
   * @param {function}    opts.onLog     – (msg, type) => void
   */
  constructor({ container, onSpawn, onLog }) {
    this._container = container;
    this._onSpawn   = onSpawn || (() => {});
    this._log       = onLog   || (() => {});
    this._assets    = [];   // { id, name, type, url, size }
    this._idSeq     = 1;
    this._filter    = 'all';
    this._search    = '';
    this._build();
  }

  // ----------------------------------------------------------------
  // Build UI
  // ----------------------------------------------------------------
  _build() {
    this._container.innerHTML = `
      <div class="assets-toolbar">
        <button class="tb-btn" id="ap-import">&#8682; Importar</button>
        <input type="file" id="ap-file-input" multiple accept="image/*,audio/*,.ks,.tmj,.tmx,.json" style="display:none">
        <input class="prop-input ap-search" id="ap-search" placeholder="Buscar asset..." type="search">
      </div>
      <div class="assets-filters" id="ap-filters">
        <button class="af-btn active" data-filter="all">Todos</button>
        <button class="af-btn" data-filter="image">🖼 Imagens</button>
        <button class="af-btn" data-filter="audio">🔊 Áudio</button>
        <button class="af-btn" data-filter="tilemap">🗺 Tilemaps</button>
        <button class="af-btn" data-filter="script">📜 Scripts</button>
      </div>
      <div class="assets-drop-zone" id="ap-drop-zone">
        <div class="assets-drop-hint" id="ap-drop-hint">
          <div style="font-size:28px">📂</div>
          <div>Arraste arquivos aqui ou clique em Importar</div>
          <div style="font-size:11px;margin-top:4px;opacity:.5">PNG · JPG · SVG · MP3 · OGG · WAV · KS · TMJ</div>
        </div>
        <div class="assets-grid" id="ap-grid"></div>
      </div>
      <div class="assets-status" id="ap-status">0 assets</div>
    `;

    // Import button
    const fileInput = this._container.querySelector('#ap-file-input');
    this._container.querySelector('#ap-import').addEventListener('click', () => fileInput.click());
    fileInput.addEventListener('change', ev => this._importFiles(ev.target.files));

    // Search
    this._container.querySelector('#ap-search').addEventListener('input', ev => {
      this._search = ev.target.value.toLowerCase();
      this._renderGrid();
    });

    // Filters
    this._container.querySelectorAll('.af-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        this._container.querySelectorAll('.af-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        this._filter = btn.dataset.filter;
        this._renderGrid();
      });
    });

    // Drop zone (file drop from OS)
    const dropZone = this._container.querySelector('#ap-drop-zone');
    dropZone.addEventListener('dragover',  ev => { ev.preventDefault(); dropZone.classList.add('drag-over'); });
    dropZone.addEventListener('dragleave', ()  => dropZone.classList.remove('drag-over'));
    dropZone.addEventListener('drop', ev => {
      ev.preventDefault();
      dropZone.classList.remove('drag-over');
      this._importFiles(ev.dataTransfer.files);
    });
  }

  // ----------------------------------------------------------------
  // Import
  // ----------------------------------------------------------------
  _importFiles(fileList) {
    if (!fileList || fileList.length === 0) return;
    let count = 0;
    for (const file of fileList) {
      const reader = new FileReader();
      const assetType = typeOf(file.name);
      reader.onload = ev => {
        const asset = {
          id:   this._idSeq++,
          name: file.name,
          type: assetType,
          url:  ev.target.result,
          size: file.size,
        };
        this._assets.push(asset);
        count++;
        this._renderGrid();
        this._updateStatus();
      };
      if (assetType === 'image') reader.readAsDataURL(file);
      else if (assetType === 'audio') reader.readAsDataURL(file);
      else reader.readAsText(file);
    }
    this._log(`${fileList.length} arquivo(s) importado(s).`, 'ok');
  }

  // ----------------------------------------------------------------
  // Render grid
  // ----------------------------------------------------------------
  _renderGrid() {
    const grid = this._container.querySelector('#ap-grid');
    const hint = this._container.querySelector('#ap-drop-hint');
    if (!grid) return;

    const visible = this._assets.filter(a => {
      if (this._filter !== 'all' && a.type !== this._filter) return false;
      if (this._search && !a.name.toLowerCase().includes(this._search)) return false;
      return true;
    });

    hint.style.display = this._assets.length === 0 ? 'flex' : 'none';
    grid.innerHTML = '';

    for (const asset of visible) {
      const card = document.createElement('div');
      card.className = 'asset-card';
      card.dataset.assetId = asset.id;
      card.title = `${asset.name}\n${formatSize(asset.size)}`;

      // Thumbnail
      const thumb = document.createElement('div');
      thumb.className = 'asset-thumb';
      if (asset.type === 'image') {
        const img = document.createElement('img');
        img.src = asset.url;
        img.alt = asset.name;
        img.style.cssText = 'width:100%;height:100%;object-fit:contain;border-radius:3px;';
        thumb.appendChild(img);
      } else {
        thumb.textContent = ASSET_TYPES[asset.type]?.icon || '📄';
        thumb.style.fontSize = '28px';
      }

      // Name
      const label = document.createElement('div');
      label.className = 'asset-label';
      label.textContent = asset.name.length > 16 ? asset.name.slice(0, 14) + '…' : asset.name;

      card.appendChild(thumb);
      card.appendChild(label);

      // Drag-from-panel to canvas
      card.draggable = true;
      card.addEventListener('dragstart', ev => {
        ev.dataTransfer.setData('kora/asset-id', asset.id);
        ev.dataTransfer.effectAllowed = 'copy';
      });

      // Double-click → spawn at scene center
      card.addEventListener('dblclick', () => this._spawn(asset, 0, 0));

      // Context menu
      card.addEventListener('contextmenu', ev => {
        ev.preventDefault();
        this._showContextMenu(ev, asset);
      });

      grid.appendChild(card);
    }
  }

  // ----------------------------------------------------------------
  // Spawn
  // ----------------------------------------------------------------
  _spawn(asset, worldX, worldY) {
    this._onSpawn({ asset, worldX, worldY });
    this._log(`Spawned "${asset.name}" na cena.`, 'ok');
  }

  // ----------------------------------------------------------------
  // Context menu
  // ----------------------------------------------------------------
  _showContextMenu(ev, asset) {
    document.querySelector('#ap-ctx-menu')?.remove();
    const menu = document.createElement('div');
    menu.id = 'ap-ctx-menu';
    menu.className = 'ctx-menu';
    menu.style.cssText = `left:${ev.clientX}px;top:${ev.clientY}px`;
    menu.innerHTML = `
      <button class="ctx-item" data-action="spawn">➕ Adicionar à cena</button>
      <button class="ctx-item" data-action="rename">✏️ Renomear</button>
      <div class="ctx-sep"></div>
      <button class="ctx-item ctx-danger" data-action="delete">🗑 Remover</button>
    `;
    document.body.appendChild(menu);
    const close = () => menu.remove();
    setTimeout(() => document.addEventListener('click', close, { once: true }), 0);

    menu.querySelector('[data-action="spawn"]').addEventListener('click', () => {
      this._spawn(asset, 0, 0); close();
    });
    menu.querySelector('[data-action="rename"]').addEventListener('click', () => {
      const n = prompt('Novo nome:', asset.name);
      if (n && n.trim()) { asset.name = n.trim(); this._renderGrid(); } close();
    });
    menu.querySelector('[data-action="delete"]').addEventListener('click', () => {
      this._assets = this._assets.filter(a => a.id !== asset.id);
      this._renderGrid(); this._updateStatus();
      this._log(`Asset "${asset.name}" removido.`, 'warn'); close();
    });
  }

  // ----------------------------------------------------------------
  // Canvas drop target registration
  // ----------------------------------------------------------------
  /**
   * Call this once with the scene canvas element.
   * When an asset card is dropped onto the canvas, it calls onSpawn
   * with the world coordinates of the drop point.
   */
  registerCanvasDrop(canvasEl, screenToWorld) {
    canvasEl.addEventListener('dragover', ev => {
      if (ev.dataTransfer.types.includes('kora/asset-id')) {
        ev.preventDefault();
        ev.dataTransfer.dropEffect = 'copy';
        canvasEl.classList.add('canvas-drop-target');
      }
    });
    canvasEl.addEventListener('dragleave', () => canvasEl.classList.remove('canvas-drop-target'));
    canvasEl.addEventListener('drop', ev => {
      ev.preventDefault();
      canvasEl.classList.remove('canvas-drop-target');
      const id = parseInt(ev.dataTransfer.getData('kora/asset-id'));
      const asset = this._assets.find(a => a.id === id);
      if (!asset) return;
      const r = canvasEl.getBoundingClientRect();
      const [wx, wy] = screenToWorld(ev.clientX - r.left, ev.clientY - r.top);
      this._spawn(asset, Math.round(wx), Math.round(wy));
    });
  }

  // ----------------------------------------------------------------
  // Helpers
  // ----------------------------------------------------------------
  _updateStatus() {
    const el = this._container.querySelector('#ap-status');
    if (el) el.textContent = `${this._assets.length} asset${this._assets.length !== 1 ? 's' : ''}`;
  }

  getAsset(id) { return this._assets.find(a => a.id === id); }
  getAll()     { return [...this._assets]; }
}

function formatSize(bytes) {
  if (!bytes) return '';
  if (bytes < 1024)       return bytes + ' B';
  if (bytes < 1024*1024)  return (bytes/1024).toFixed(1) + ' KB';
  return (bytes/(1024*1024)).toFixed(1) + ' MB';
}

window.AssetsPanel = AssetsPanel;
