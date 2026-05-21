/* =============================================================
   Kora Editor — git-panel.js
   Git version control panel with mock data. Designed to be
   connected to a local HTTP bridge for real Git operations.
   ============================================================= */
'use strict';

class GitPanel {
  /**
   * @param {HTMLElement} container  – element that will host the panel UI
   * @param {object}      opts
   * @param {function}    opts.onLog – (msg, type) => void
   */
  constructor(container, opts = {}) {
    this._container = container;
    this._onLog     = opts.onLog || (() => {});

    // Internal state
    this._visible     = false;
    this._branch      = 'main';
    this._files       = [];
    this._commits     = [];
    this._diffContent = '';
    this._diffPath    = '';

    this._buildUI();
    this._initEvents();
  }

  // ----------------------------------------------------------------
  // Public API
  // ----------------------------------------------------------------

  async refresh() {
    this._log('Refreshing Git status...', 'info');
    await this._execGitCommand(['status', '--porcelain']);
    await this._execGitCommand(['branch', '--show-current']);
    await this._execGitCommand(['log', '--oneline', '-10']);
    this.render();
    this._log('Git status refreshed.', 'ok');
  }

  show() {
    this._visible = true;
    this._container.style.display = 'flex';
    this.refresh();
  }

  hide() {
    this._visible = false;
    this._container.style.display = 'none';
  }

  // ----------------------------------------------------------------
  // Git command abstraction
  // ----------------------------------------------------------------

  /**
   * Execute a Git command.
   * Currently returns mock data. In production, this should call a
   * local HTTP API that shells out to Git.
   *
   * @param {string[]} args - Git arguments (e.g. ['status', '--porcelain'])
   * @returns {Promise<string>} stdout
   */
  async _execGitCommand(args) {
    const cmd = args.join(' ');
    this._log(`$ git ${cmd}`, 'info');

    // Simulate network delay
    await new Promise(r => setTimeout(r, 80 + Math.random() * 80));

    // Mock responses based on command
    if (args[0] === 'status' && args[1] === '--porcelain') {
      this._files = this._mockStatus();
      return '';
    }
    if (args[0] === 'branch' && args[1] === '--show-current') {
      this._branch = this._mockBranch();
      return this._branch;
    }
    if (args[0] === 'log' && args[1] === '--oneline') {
      this._commits = this._mockLog();
      return this._commits.map(c => c.hash + ' ' + c.message).join('\n');
    }
    if (args[0] === 'diff') {
      const path = args[args.length - 1];
      this._diffPath   = path;
      this._diffContent = this._mockDiff(path);
      return this._diffContent;
    }
    if (args[0] === 'add') {
      const path = args[args.length - 1];
      this._mockStage(path);
      return '';
    }
    if (args[0] === 'restore' && args[1] === '--staged') {
      const path = args[args.length - 1];
      this._mockUnstage(path);
      return '';
    }
    if (args[0] === 'commit') {
      this._mockCommit(args[args.indexOf('-m') + 1]);
      return '';
    }

    return '';
  }

  // ----------------------------------------------------------------
  // Stage / Unstage / Commit / Diff
  // ----------------------------------------------------------------

  async stageFile(path) {
    await this._execGitCommand(['add', path]);
    this.render();
    this._log(`Staged: ${path}`, 'ok');
  }

  async unstageFile(path) {
    await this._execGitCommand(['restore', '--staged', path]);
    this.render();
    this._log(`Unstaged: ${path}`, 'warn');
  }

  async commit(message) {
    if (!message || !message.trim()) return;
    await this._execGitCommand(['commit', '-m', message.trim()]);
    this._log(`Committed: ${message.trim()}`, 'ok');
    this.render();
    // Refresh history and status
    await this._execGitCommand(['log', '--oneline', '-10']);
    this.render();
  }

  async showDiff(path) {
    await this._execGitCommand(['diff', path]);
    this._renderDiff();
  }

  // ----------------------------------------------------------------
  // UI Build
  // ----------------------------------------------------------------

  _buildUI() {
    this._container.innerHTML = `
      <div class="git-panel">
        <!-- Toolbar -->
        <div class="git-toolbar">
          <span class="git-branch" id="git-branch">
            <span class="git-branch-icon">⎇</span>
            <span id="git-branch-name">main</span>
          </span>
          <button class="tb-btn git-refresh-btn" id="git-refresh" title="Refresh status">↻</button>
          <button class="tb-btn git-commit-btn" id="git-commit-btn" title="Commit staged changes">✔ Commit</button>
        </div>

        <!-- File list -->
        <div class="git-section-label">Changed files</div>
        <div class="git-file-list" id="git-file-list"></div>

        <!-- Commit message area -->
        <div class="git-commit-area" id="git-commit-area" style="display:none">
          <textarea class="git-commit-input" id="git-commit-input" placeholder="Commit message..." rows="3"></textarea>
          <div class="git-commit-actions">
            <button class="btn-secondary" id="git-commit-cancel">Cancel</button>
            <button class="btn-primary" id="git-commit-confirm">Commit</button>
          </div>
        </div>

        <!-- Diff view -->
        <div class="git-diff-area" id="git-diff-area" style="display:none">
          <div class="git-diff-header">
            <span class="git-diff-title" id="git-diff-title">Diff</span>
            <button class="icon-btn git-diff-close" id="git-diff-close">&times;</button>
          </div>
          <pre class="git-diff-content" id="git-diff-content"></pre>
        </div>

        <!-- History -->
        <div class="git-section-label">History</div>
        <div class="git-history" id="git-history"></div>
      </div>
    `;
  }

  _initEvents() {
    const refreshBtn     = this._container.querySelector('#git-refresh');
    const commitBtn      = this._container.querySelector('#git-commit-btn');
    const commitArea     = this._container.querySelector('#git-commit-area');
    const commitInput    = this._container.querySelector('#git-commit-input');
    const commitCancel   = this._container.querySelector('#git-commit-cancel');
    const commitConfirm  = this._container.querySelector('#git-commit-confirm');
    const diffClose      = this._container.querySelector('#git-diff-close');

    if (refreshBtn) {
      refreshBtn.addEventListener('click', () => this.refresh());
    }

    if (commitBtn) {
      commitBtn.addEventListener('click', () => {
        commitArea.style.display = 'block';
        commitInput.focus();
      });
    }

    if (commitCancel) {
      commitCancel.addEventListener('click', () => {
        commitArea.style.display = 'none';
        commitInput.value = '';
      });
    }

    if (commitConfirm) {
      commitConfirm.addEventListener('click', async () => {
        const msg = commitInput.value.trim();
        if (!msg) return;
        await this.commit(msg);
        commitArea.style.display = 'none';
        commitInput.value = '';
      });
    }

    if (commitInput) {
      commitInput.addEventListener('keydown', (ev) => {
        if (ev.key === 'Enter' && (ev.ctrlKey || ev.metaKey)) {
          ev.preventDefault();
          commitConfirm.click();
        }
        if (ev.key === 'Escape') {
          commitArea.style.display = 'none';
          commitInput.value = '';
        }
      });
    }

    if (diffClose) {
      diffClose.addEventListener('click', () => {
        this._container.querySelector('#git-diff-area').style.display = 'none';
      });
    }

    // Delegate click events for file list items
    const fileList = this._container.querySelector('#git-file-list');
    if (fileList) {
      fileList.addEventListener('click', (ev) => {
        const item = ev.target.closest('.git-file-item');
        if (!item) return;
        const path = item.dataset.path;
        if (!path) return;

        // Check if click was on the status badge
        const badge = ev.target.closest('.git-file-status');
        if (badge) {
          // Toggle stage/unstage
          if (item.dataset.staged === 'true') {
            this.unstageFile(path);
          } else {
            this.stageFile(path);
          }
          return;
        }

        // Single click on the item: also toggle stage/unstage
        if (item.dataset.staged === 'true') {
          this.unstageFile(path);
        } else {
          this.stageFile(path);
        }
      });

      fileList.addEventListener('dblclick', (ev) => {
        const item = ev.target.closest('.git-file-item');
        if (!item) return;
        const path = item.dataset.path;
        if (path) {
          this.showDiff(path);
        }
      });
    }
  }

  // ----------------------------------------------------------------
  // Render
  // ----------------------------------------------------------------

  render() {
    // Update branch name
    const branchNameEl = this._container.querySelector('#git-branch-name');
    if (branchNameEl) branchNameEl.textContent = this._branch;

    // Render file list
    const fileList = this._container.querySelector('#git-file-list');
    if (fileList) {
      fileList.innerHTML = '';
      if (this._files.length === 0) {
        fileList.innerHTML = '<div class="git-empty">Working tree clean</div>';
      } else {
        for (const f of this._files) {
          fileList.appendChild(this._createFileItem(f));
        }
      }
    }

    // Render history
    const historyEl = this._container.querySelector('#git-history');
    if (historyEl) {
      historyEl.innerHTML = '';
      if (this._commits.length === 0) {
        historyEl.innerHTML = '<div class="git-empty">No commits yet</div>';
      } else {
        for (const c of this._commits) {
          historyEl.appendChild(this._createCommitItem(c));
        }
      }
    }
  }

  _createFileItem(file) {
    const item = document.createElement('div');
    item.className    = 'git-file-item';
    item.dataset.path = file.path;
    item.dataset.staged = file.staged ? 'true' : 'false';

    // Status badge
    const badge = document.createElement('span');
    badge.className = 'git-file-status';
    if (file.staged) {
      badge.textContent = '✓';
      badge.classList.add('staged');
    } else {
      badge.textContent = file.porcelain;
      badge.classList.add(this._statusClass(file.porcelain));
    }
    item.appendChild(badge);

    // File path
    const label = document.createElement('span');
    label.className = 'git-file-label';
    label.textContent = file.path;
    item.appendChild(label);

    // Action indicator (stage/unstage hint)
    const hint = document.createElement('span');
    hint.className = 'git-file-hint';
    hint.textContent = file.staged ? 'click to unstage' : 'click to stage';
    item.appendChild(hint);

    return item;
  }

  _createCommitItem(commit) {
    const item = document.createElement('div');
    item.className = 'git-commit-item';

    const hash = document.createElement('span');
    hash.className = 'git-commit-hash';
    hash.textContent = commit.hash;
    item.appendChild(hash);

    const msg = document.createElement('span');
    msg.className = 'git-commit-msg';
    msg.textContent = commit.message;
    item.appendChild(msg);

    const meta = document.createElement('span');
    meta.className = 'git-commit-meta';
    meta.textContent = commit.date + (commit.author ? ` · ${commit.author}` : '');
    item.appendChild(meta);

    return item;
  }

  _renderDiff() {
    const diffArea   = this._container.querySelector('#git-diff-area');
    const diffTitle  = this._container.querySelector('#git-diff-title');
    const diffContent = this._container.querySelector('#git-diff-content');

    if (!diffArea || !diffTitle || !diffContent) return;

    diffTitle.textContent = `Diff — ${this._diffPath}`;
    diffContent.textContent = this._diffContent || '(no diff content)';
    diffArea.style.display = 'flex';

    // Scroll to the diff area
    diffArea.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  }

  // ----------------------------------------------------------------
  // Mock Data
  // ----------------------------------------------------------------

  _mockStatus() {
    return [
      { path: 'src/scenes/title.kora',      porcelain: ' M', staged: false },
      { path: 'src/scenes/level1.kora',      porcelain: 'M ', staged: true  },
      { path: 'src/entities/player.kora',    porcelain: ' M', staged: false },
      { path: 'src/entities/enemy.kora',     porcelain: ' M', staged: false },
      { path: 'src/systems/physics.kora',    porcelain: 'A ', staged: true  },
      { path: 'src/ui/hud.kora',             porcelain: '??', staged: false },
      { path: 'assets/sprites/player.png',   porcelain: '??', staged: false },
      { path: 'assets/sprites/enemy.png',    porcelain: ' M', staged: true  },
      { path: 'package.json',                porcelain: ' D', staged: false },
      { path: 'src/config.kora',             porcelain: 'M ', staged: true  },
    ];
  }

  _mockBranch() {
    return 'feature/git-integration';
  }

  _mockLog() {
    return [
      { hash: 'a1b2c3d', message: 'Add physics system',        author: 'dev',   date: '2 hours ago' },
      { hash: 'e4f5g6h', message: 'Fix player collision',      author: 'dev',   date: '5 hours ago' },
      { hash: 'i7j8k9l', message: 'Refactor scene loader',     author: 'dev',   date: '1 day ago' },
      { hash: 'm0n1o2p', message: 'Add tilemap support',       author: 'dev',   date: '2 days ago' },
      { hash: 'q3r4s5t', message: 'Initial HUD implementation', author: 'dev',   date: '3 days ago' },
      { hash: 'u6v7w8x', message: 'Setup build pipeline',       author: 'dev',   date: '5 days ago' },
      { hash: 'y9z0a1b', message: 'Add asset manager',         author: 'dev',   date: '1 week ago' },
      { hash: 'c2d3e4f', message: 'Implement save/load JSON',  author: 'dev',   date: '1 week ago' },
      { hash: 'g5h6i7j', message: 'Add canvas renderer',       author: 'dev',   date: '2 weeks ago' },
      { hash: 'k8l9m0n', message: 'Initial project setup',     author: 'dev',   date: '3 weeks ago' },
    ];
  }

  _mockDiff(path) {
    const ext = path.split('.').pop();
    if (ext === 'png') {
      return 'diff --git a/' + path + ' b/' + path + '\n'
        + 'Binary files differ\n';
    }
    return 'diff --git a/' + path + ' b/' + path + '\n'
      + 'index abc123..def456 100644\n'
      + '--- a/' + path + '\n'
      + '+++ b/' + path + '\n'
      + '@@ -10,6 +10,8 @@\n'
      + '   // previous line\n'
      + '-\n'
      + '+  // added line\n'
      + '+  // another addition\n'
      + '   // unchanged context\n'
      + '   // more context\n'
      + ' \n'
      + '+// new function\n'
      + '+function setup() {\n'
      + '+  init();\n'
      + '+}\n';
  }

  _mockStage(path) {
    const file = this._files.find(f => f.path === path);
    if (file) {
      file.staged = true;
      file.porcelain = file.porcelain.replace(/[ MAD?!]/, 'M');
      file.porcelain = 'M ';
    }
  }

  _mockUnstage(path) {
    const file = this._files.find(f => f.path === path);
    if (file) {
      file.staged = false;
      if (file.porcelain.startsWith('M')) {
        file.porcelain = ' M';
      } else if (file.porcelain.startsWith('A')) {
        file.porcelain = ' A';
      }
    }
  }

  _mockCommit(message) {
    const now   = new Date();
    const hash  = Array.from({ length: 7 }, () =>
      '0123456789abcdef'[Math.floor(Math.random() * 16)]
    ).join('');
    const dateStr = 'just now';
    this._commits.unshift({
      hash,
      message: message,
      author: 'you',
      date: dateStr,
    });
    // Remove staged files
    this._files = this._files.filter(f => !f.staged);
  }

  // ----------------------------------------------------------------
  // Helpers
  // ----------------------------------------------------------------

  _statusClass(porcelain) {
    if (porcelain.startsWith('M')) return 'status-modified';
    if (porcelain.startsWith('A')) return 'status-added';
    if (porcelain.startsWith('D')) return 'status-deleted';
    if (porcelain.startsWith('?')) return 'status-untracked';
    return '';
  }

  _log(msg, type) {
    this._onLog(msg, type);
  }
}

window.GitPanel = GitPanel;
