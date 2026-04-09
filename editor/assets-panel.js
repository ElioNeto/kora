/* =============================================================
   Kora Editor — assets-panel.js
   Manages the Assets tab: image/audio import, asset library,
   drag-and-drop from panel onto canvas to spawn entities.
============================================================= */
'use strict';

const ASSET_TYPES = {
  image: ['png','jpg','jpeg','gif','webp','svg'],
  audio: ['mp3','ogg','wav'],
  tileset: ['png','jpg','jpeg'],
};

const MIME_ICONS = {
  image:   '🖼️',
  audio:   '🔊',
  tileset: '🔲',
};

class AssetsPanel {
  /**
   * @param {object} opts
   * @param {HTMLElement} opts.container     – element for the panel UI
   * @param {HTMLElement} opts.canvas        – scene canvas (drop target)
   * @param {function}    opts.screenToWorld – (sx,sy) => [wx,wy]
   * @param {function}    opts.spawnEntity   – (name, type, wx, wy, assetId) => void
   * @param {function}    opts.onLog         – (msg, type) => void
   */
  constructor({ container, canvas, screenToWorld, spawnEntity, onLog }) {
    this._container     = container;
    this._canvas        = canvas;
    this._screenToWorld = screenToWorld;
    this._spawnEntity   = spawnEntity;
    this._log           = onLog || (() => {});

    /** @type {Map<string, Asset>} */
    this._assets = new Map();
    this._dragAssetId = null;
    this._dbInitPromise = null;

    this._buildUI();
    this._bindCanvasDrop();
    this._loadFromDB();
  }

  // ----------------------------------------------------------------
  // Public API
  // ----------------------------------------------------------------

  /** Returns a copy of all stored assets. */
  getAll() {
    return [...this._assets.values()];
  }

  /** Returns the asset object for a given id, or null. */
  get(id) {
    return this._assets.get(id) || null;
  }

  // ----------------------------------------------------------------
  // UI
  // ----------------------------------------------------------------

  _buildUI() {
    this._container.innerHTML = `
      <div class="assets-panel">
        <div class="assets-toolbar">
          <button class="tb-btn" id="assets-import-btn" title="Importar arquivo">
            &#43; Importar
          </button>
          <input type="file" id="assets-file-input"
            accept="image/*,audio/*"
            multiple hidden>
          <div class="assets-filter">
            <button class="af-btn active" data-filter="all">Todos</button>
            <button class="af-btn" data-filter="image">🖼️</button>
            <button class="af-btn" data-filter="audio">🔊</button>
          </div>
        </div>

        <div class="assets-drop-zone" id="assets-drop-zone">
          <div class="assets-drop-hint">
            <span style="font-size:28px">📂</span>
            <span>Solte arquivos aqui ou clique em Importar</span>
          </div>
        </div>

        <div class="assets-grid" id="assets-grid"></div>

        <div class="assets-status" id="assets-status">0 assets</div>
      </div>
    `;

    // Import button
    this._container.querySelector('#assets-import-btn')
      .addEventListener('click', () =>
        this._container.querySelector('#assets-file-input').click()
      );

    // File picker
    this._container.querySelector('#assets-file-input')
      .addEventListener('change', ev => {
        this._importFiles(ev.target.files);
        ev.target.value = '';
      });

    // Drop zone
    const dz = this._container.querySelector('#assets-drop-zone');
    dz.addEventListener('dragover',  ev => { ev.preventDefault(); dz.classList.add('drag-over'); });
    dz.addEventListener('dragleave', ()  => dz.classList.remove('drag-over'));
    dz.addEventListener('drop', ev => {
      ev.preventDefault();
      dz.classList.remove('drag-over');
      this._importFiles(ev.dataTransfer.files);
    });

    // Filter buttons
    this._container.querySelectorAll('.af-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        this._container.querySelectorAll('.af-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        this._renderGrid(btn.dataset.filter);
      });
    });
  }

  // ----------------------------------------------------------------
  // IndexedDB
  // ----------------------------------------------------------------

  async _loadFromDB() {
    this._dbInitPromise = this._dbInitPromise || AssetDB.init();
    try {
      const assets = await AssetDB.getAll();
      for (const a of assets) {
        // Recreate blob URL for each loaded asset
        if (a.blob) {
          a.url = URL.createObjectURL(a.blob);
        }
        this._assets.set(a.id, a);
      }
      this._log(`Recuperado ${assets.length} asset(s) do armazenamento.`, 'ok');
      this._renderGrid();
    } catch(err) {
      console.error('Erro ao carregar assets do IndexedDB:', err);
    }
  }

  async _saveToDB(asset) {
    if (!this._dbInitPromise) this._dbInitPromise = AssetDB.init();
    await this._dbInitPromise;
    await AssetDB.add(asset);
  }

  async _deleteFromDB(id) {
    if (!this._dbInitPromise) this._dbInitPromise = AssetDB.init();
    await this._dbInitPromise;
    await AssetDB.delete(id);
  }

  // ----------------------------------------------------------------
  // Import
  // ----------------------------------------------------------------

  async _importFiles(fileList) {
    if (!fileList?.length) return;
    let imported = 0;
    for (const file of fileList) {
      const ext  = file.name.split('.').pop().toLowerCase();
      const kind = this._detectKind(ext, file.type);
      if (!kind) {
        this._log(`Formato não suportado: ${file.name}`, 'warn');
        continue;
      }
      // Para imagens e áudio, usar Blobs para salvar no IndexedDB
      const data = await file.arrayBuffer();
      const blob  = new Blob([data], { type: file.type || 'application/octet-stream' });
      const url   = URL.createObjectURL(blob);
      const id    = `asset_${Date.now()}_${Math.random().toString(36).slice(2,7)}`;
      const asset = { id, name: file.name, kind, url, size: file.size, ext, blob, contentType: file.type };
      this._assets.set(id, asset);
      imported++;
      await this._saveToDB(asset);
      this._log(`Importado: ${file.name} (${kind})`, 'ok');
    }
    if (imported > 0) this._renderGrid();
  }

  _detectKind(ext, mime) {
    if (mime.startsWith('image/') || ASSET_TYPES.image.includes(ext)) return 'image';
    if (mime.startsWith('audio/') || ASSET_TYPES.audio.includes(ext)) return 'audio';
    return null;
  }

  // ----------------------------------------------------------------
  // Grid
  // ----------------------------------------------------------------

  _renderGrid(filter = 'all') {
    const grid = this._container.querySelector('#assets-grid');
    const status = this._container.querySelector('#assets-status');
    const dz   = this._container.querySelector('#assets-drop-zone');
    const assets = [...this._assets.values()]
      .filter(a => filter === 'all' || a.kind === filter);

    dz.style.display  = this._assets.size === 0 ? 'flex' : 'none';
    grid.style.display = this._assets.size > 0  ? 'grid' : 'none';

    grid.innerHTML = '';
    for (const asset of assets) {
      grid.appendChild(this._buildCard(asset));
    }
    status.textContent = `${this._assets.size} asset${this._assets.size !== 1 ? 's' : ''}`;
  }

  _buildCard(asset) {
    const card = document.createElement('div');
    card.className  = 'asset-card';
    card.draggable  = true;
    card.dataset.id = asset.id;
    card.title      = `${asset.name}\n${this._fmtSize(asset.size)}\nArraste para a cena`;

    // Thumbnail
    const thumb = document.createElement('div');
    thumb.className = 'asset-thumb';
    if (asset.kind === 'image') {
      const img = document.createElement('img');
      img.src    = asset.url;
      img.alt    = asset.name;
      img.width  = 64;
      img.height = 64;
      img.loading = 'lazy';
      img.style.objectFit = 'contain';
      thumb.appendChild(img);
    } else {
      thumb.innerHTML = `<span style="font-size:28px">${MIME_ICONS[asset.kind] || '📄'}</span>`;
    }
    card.appendChild(thumb);

    // Label
    const label = document.createElement('div');
    label.className   = 'asset-label';
    label.textContent = this._truncate(asset.name, 14);
    card.appendChild(label);

    // Delete button
    const del = document.createElement('button');
    del.className = 'asset-delete';
    del.innerHTML = '×';
    del.title     = 'Remover asset';
    del.addEventListener('click', async ev => {
      ev.stopPropagation();
      URL.revokeObjectURL(asset.url);
      asset.blob?.slice(); // Garbage collection hint
      await this._deleteFromDB(asset.id);
      this._assets.delete(asset.id);
      this._log(`Removido: ${asset.name}`, 'warn');
      this._renderGrid();
    });
    card.appendChild(del);

    // Drag start: mark which asset is being dragged
    card.addEventListener('dragstart', ev => {
      this._dragAssetId = asset.id;
      ev.dataTransfer.setData('text/plain', asset.id);
      ev.dataTransfer.effectAllowed = 'copy';
      card.classList.add('dragging');
    });
    card.addEventListener('dragend', () => {
      card.classList.remove('dragging');
      this._dragAssetId = null;
    });

    return card;
  }

  // ----------------------------------------------------------------
  // Canvas drop
  // ----------------------------------------------------------------

  _bindCanvasDrop() {
    const c = this._canvas;

    c.addEventListener('dragover', ev => {
      if (!this._dragAssetId) return;
      ev.preventDefault();
      ev.dataTransfer.dropEffect = 'copy';
      c.style.outline = '2px solid #00e5a0';
    });
    c.addEventListener('dragleave', () => {
      c.style.outline = '';
    });
    c.addEventListener('drop', ev => {
      ev.preventDefault();
      c.style.outline = '';
      const id = this._dragAssetId || ev.dataTransfer.getData('text/plain');
      if (!id) return;
      const asset = this._assets.get(id);
      if (!asset) return;

      const rect   = c.getBoundingClientRect();
      const [wx, wy] = this._screenToWorld(ev.clientX - rect.left, ev.clientY - rect.top);
      const entityType = asset.kind === 'audio' ? 'audio' : 'sprite';
      this._spawnEntity(asset.name.replace(/\.[^.]+$/, ''), entityType, wx, wy, id);
      this._log(`Entidade criada a partir de: ${asset.name}`, 'ok');
    });
  }

  // ----------------------------------------------------------------
  // Helpers
  // ----------------------------------------------------------------

  _truncate(str, max) {
    return str.length > max ? str.slice(0, max - 1) + '…' : str;
  }

  _fmtSize(bytes) {
    if (bytes < 1024)       return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }
}

window.AssetsPanel = AssetsPanel;
