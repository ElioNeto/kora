/* =============================================================
   Kora Editor — preview-panel.js
   Manages the embedded preview iframe inside the editor.
   Communicates with preview.html via postMessage.
============================================================= */
'use strict';

class PreviewPanel {
  /**
   * @param {object} opts
   * @param {HTMLElement} opts.container   – element that will host the iframe
   * @param {function}    opts.getScene    – () => { entities, meta }
   * @param {function}    opts.onLog       – (msg, type) => void
   */
  constructor({ container, getScene, onLog }) {
    this._container = container;
    this._getScene  = getScene;
    this._log       = onLog || (() => {});
    this._running   = false;
    this._iframe    = null;

    this._build();
  }

  // ---- Private ------------------------------------------------

  _build() {
    this._container.innerHTML = '';

    const wrap = document.createElement('div');
    wrap.className = 'preview-wrap';

    // Toolbar
    const toolbar = document.createElement('div');
    toolbar.className = 'preview-toolbar';
    toolbar.innerHTML = `
      <span class="preview-title">&#9654; Preview</span>
      <div class="preview-toolbar-actions">
        <button class="tb-btn" id="preview-play">&#9654; Rodar</button>
        <button class="tb-btn" id="preview-stop" disabled>&#9646;&#9646; Parar</button>
        <button class="tb-btn" id="preview-reload" title="Recarregar cena">&#8635;</button>
        <select class="tb-btn" id="preview-device" title="Tamanho do dispositivo">
          <option value="360x640">360 × 640 (Android)</option>
          <option value="390x844">390 × 844 (iPhone 14)</option>
          <option value="412x915">412 × 915 (Pixel 7)</option>
          <option value="768x1024">768 × 1024 (Tablet)</option>
        </select>
      </div>
    `;
    wrap.appendChild(toolbar);

    // Iframe shell
    const shell = document.createElement('div');
    shell.className = 'preview-shell';

    const frame = document.createElement('iframe');
    frame.src        = './preview.html';
    frame.className  = 'preview-iframe';
    frame.title      = 'Kora Game Preview';
    frame.setAttribute('sandbox', 'allow-scripts allow-same-origin');
    shell.appendChild(frame);
    wrap.appendChild(shell);

    this._container.appendChild(wrap);
    this._iframe = frame;

    // Wait for iframe to be ready
    frame.addEventListener('load', () => {
      this._log('Preview carregado. Clique em Rodar para iniciar.', 'info');
    });

    // Buttons
    toolbar.querySelector('#preview-play').addEventListener('click',   () => this.play());
    toolbar.querySelector('#preview-stop').addEventListener('click',   () => this.stop());
    toolbar.querySelector('#preview-reload').addEventListener('click', () => this.reload());
    toolbar.querySelector('#preview-device').addEventListener('change', ev => {
      const [w, h] = ev.target.value.split('x').map(Number);
      this._resizeShell(w, h);
    });

    this._resizeShell(360, 640);
  }

  _resizeShell(w, h) {
    const shell = this._container.querySelector('.preview-shell');
    if (!shell) return;
    // Scale to fit available height
    const maxH = this._container.clientHeight - 48;
    const scale = Math.min(1, maxH / h, (this._container.clientWidth - 24) / w);
    shell.style.width  = w + 'px';
    shell.style.height = h + 'px';
    shell.style.transform = `scale(${scale})`;
    shell.style.transformOrigin = 'top center';
  }

  _post(msg) {
    if (!this._iframe?.contentWindow) return;
    this._iframe.contentWindow.postMessage(msg, '*');
  }

  _buildSceneDoc() {
    const { entities, meta } = this._getScene();
    return {
      kora:     '1.0',
      name:     meta.name,
      logicalW: meta.logicalW,
      logicalH: meta.logicalH,
      entities: entities.map(e => ({ ...e })),
    };
  }

  // ---- Public API -------------------------------------------

  play() {
    const doc = this._buildSceneDoc();
    this._post({ type: 'KORA_LOAD_SCENE', scene: doc });
    this._running = true;
    this._log(`Cena "${doc.name}" enviada ao preview (${doc.entities.length} entidades).`, 'ok');
    this._updateButtons(true);
  }

  stop() {
    this._post({ type: 'KORA_STOP' });
    this._running = false;
    this._log('Preview parado.', 'warn');
    this._updateButtons(false);
  }

  reload() {
    this.stop();
    setTimeout(() => this.play(), 150);
    this._log('Preview recarregado.', 'info');
  }

  isRunning() { return this._running; }

  _updateButtons(playing) {
    const p = this._container.querySelector('#preview-play');
    const s = this._container.querySelector('#preview-stop');
    if (p) p.disabled = playing;
    if (s) s.disabled = !playing;
  }
}

window.PreviewPanel = PreviewPanel;
