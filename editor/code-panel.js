/* =============================================================
   Kora Editor — code-panel.js
   CodeMirror 6 integration with KScript syntax highlight,
   autocomplete, error display, and entity binding.
============================================================= */
'use strict';

const KSCRIPT_KEYWORDS = new Set([
  'object', 'func', 'fn', 'async', 'await', 'return', 'if', 'else',
  'for', 'while', 'break', 'continue', 'const', 'var',
  'true', 'false', 'null', 'import', 'from', 'this', 'emit',
  'entity', 'scene', 'spawn', 'signal', 'when', 'new',
]);

const KSCRIPT_TYPES = new Set([
  'int', 'float', 'bool', 'string', 'void',
  'Vec2', 'Vector2', 'Entity', 'Sprite', 'Sound', 'Task', 'Signal', 'Color', 'Rect',
  'Dict', 'TweenProps',
]);

const LIFECYCLE_HOOKS = [
  'onCreate', 'onUpdate', 'onDraw', 'onDestroy',
  'onCollision', 'onInput', 'onTouch',
  'create', 'update', 'draw', 'destroy',
];

const BUILTIN_MODULES = {
  'Input': ['pressed', 'held', 'released', 'axisX', 'axisY', 'touchPos'],
  'Audio': ['play', 'stop', 'setVolume'],
  'Camera': ['setZoom', 'setPos', 'shake', 'follow'],
  'Physics': ['setGravity', 'raycast', 'overlapRect'],
  'Scene': ['spawn', 'destroy', 'find', 'load', 'reload'],
  'Entity': ['get', 'find', 'create', 'destroy'],
  'System': ['log', 'warn', 'error', 'exit', 'time'],
  'Asset': ['load', 'loadAsync', 'unload', 'get'],
};

const ASYNC_PRIMITIVES = ['wait', 'waitFrames', 'waitSignal', 'tween', 'race', 'all', 'cancel'];

const ENTITY_PROPERTIES = ['x', 'y', 'velX', 'velY', 'w', 'h', 'alpha', 'visible', 'tag', 'rotation',
  'onGround', 'destroy', 'emit', 'getNode', 'setVelocity'];

class KScriptLanguage {
  static define() {
    const keywords = KSCRIPT_KEYWORDS;
    const types = KSCRIPT_TYPES;

    function token(stream, state) {
      if (stream.eatSpace()) return null;

      if (stream.match('//')) {
        stream.skipToEnd();
        return 'comment';
      }

      if (stream.match('/*')) {
        stream.skipTo('*/');
        if (!stream.next()) { state.inBlockComment = true; }
        return 'comment';
      }
      if (state.inBlockComment) {
        const found = stream.skipTo('*/');
        if (found) { stream.next(); stream.next(); state.inBlockComment = false; }
        else { stream.skipToEnd(); }
        return 'comment';
      }

      if (stream.peek() === '"') {
        stream.next();
        while (!stream.eol()) {
          const ch = stream.peek();
          if (ch === '"') {
            stream.next();
            return 'string';
          }
          if (ch === '$') {
            stream.next();
            if (stream.peek() === '{') {
              stream.next();
              state.interpNesting++;
            } else {
              stream.match(/^[a-zA-Z_]\w*/);
            }
            return 'stringSpecial';
          }
          if (ch === '}' && state.interpNesting > 0) {
            stream.next();
            state.interpNesting--;
            return 'stringSpecial';
          }
          if (ch === '\\') { stream.next(); if (!stream.eol()) stream.next(); }
          else { stream.next(); }
        }
        return 'string';
      }

      if (stream.match(/^0[xX][0-9a-fA-F]+/)) return 'number';
      if (stream.match(/^(\d+(\.\d+)?([eE][+-]?\d+)?)/)) return 'number';

      if (/[a-zA-Z_]/.test(stream.peek())) {
        const word = stream.match(/^[a-zA-Z_]\w*/);
        if (word) {
          const w = word[0];
          if (keywords.has(w)) return 'keyword';
          if (types.has(w)) return 'typeName';
          if (LIFECYCLE_HOOKS.includes(w) || ASYNC_PRIMITIVES.includes(w)) return 'functionName';
          if (w.startsWith('on') && /^on[A-Z]/.test(w)) return 'functionName';
          return 'variable';
        }
      }

      if (stream.match(/^[=!<>]=|&&|\|\||[+\-*/%&|^~]=|\.\./)) return 'operator';
      if (stream.match(/^[+\-*/%<>=!&|^~?:]/)) return 'operator';

      if (stream.match(/^[{}()\[\];,.:]/)) return 'punctuation';

      stream.next();
      return null;
    }

    return {
      token,
      startState: () => ({ inBlockComment: false, interpNesting: 0 }),
      languageData: {
        commentTokens: { line: '//', block: { open: '/*', close: '*/' } },
        indentOnInput: /^\s*[\}\]]$/,
        closeBrackets: { brackets: ['(', '[', '{', "'", '"'] },
      },
    };
  }
}

class CodePanel {
  constructor(container, opts = {}) {
    this._container = container;
    this._onApply = opts.onApply || (() => {});
    this._changeCallback = opts.onChange || (() => {});
    this._entityNameEl = opts.nameEl || null;
    this._errorBadgeEl = opts.errorBadgeEl || null;
    this._applyBtnEl = opts.applyBtnEl || null;
    this._entityId = null;
    this._errors = [];
    this._dirty = false;
    this._pendingLoad = null;
    this._ready = false;

    this._initEditor();
    this._bindButtons();
  }

  async _whenReady() {
    while (!this._ready) await new Promise(r => setTimeout(r, 10));
  }

  async _initEditor() {
    const { EditorState } = await import('@codemirror/state');
    const {
      EditorView, keymap, lineNumbers,
      highlightActiveLineGutter, highlightSpecialChars,
      drawSelection, rectangularSelection, crosshairCursor,
    } = await import('@codemirror/view');
    const { defaultKeymap, history, historyKeymap, indentWithTab } = await import('@codemirror/commands');
    const {
      StreamLanguage, indentUnit, syntaxHighlighting, HighlightStyle,
    } = await import('@codemirror/language');
    const { autocompletion, completionKeymap, closeBrackets, closeBracketsKeymap } = await import('@codemirror/autocomplete');
    const { tags } = await import('@lezer/highlight');

    const kscriptLang = StreamLanguage.define(KScriptLanguage.define());

    const darkTheme = EditorView.theme({
      '&': {
        backgroundColor: '#0f1117',
        color: '#e6edf3',
        height: '100%',
        fontSize: '12px',
        fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
      },
      '.cm-gutters': {
        backgroundColor: '#161b22',
        color: '#484f58',
        border: 'none',
        borderRight: '1px solid #21262d',
      },
      '.cm-activeLineGutter': {
        backgroundColor: 'rgba(0,229,160,0.08)',
      },
      '.cm-activeLine': {
        backgroundColor: 'rgba(0,229,160,0.04)',
      },
      '.cm-cursor': {
        borderLeftColor: '#00e5a0',
      },
      '.cm-selectionBackground': {
        backgroundColor: 'rgba(56,139,253,0.3)',
      },
      '.cm-selectionMatch': {
        backgroundColor: 'rgba(56,139,253,0.15)',
      },
      '.cm-matchingBracket': {
        backgroundColor: 'rgba(0,229,160,0.2)',
        outline: '1px solid rgba(0,229,160,0.4)',
      },
      '.cm-scroller': {
        fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
      },
      '.cm-foldPlaceholder': {
        backgroundColor: '#1c2128',
        color: '#7d8590',
      },
    }, { dark: true });

    const kscriptHighlight = syntaxHighlighting(HighlightStyle.define([
      { tag: tags.keyword, color: '#ff7b72', fontWeight: '500' },
      { tag: tags.typeName, color: '#79c0ff' },
      { tag: tags.comment, color: '#7d8590', fontStyle: 'italic' },
      { tag: tags.string, color: '#a5d6ff' },
      { tag: tags.number, color: '#79c0ff' },
      { tag: tags.operator, color: '#e6edf3' },
      { tag: tags.punctuation, color: '#484f58' },
      { tag: tags.function(tags.variableName), color: '#d2a8ff' },
      { tag: tags.function(tags.propertyName), color: '#d2a8ff' },
      { tag: tags.function(tags.definition), color: '#00e5a0' },
      { tag: tags.variableName, color: '#e6edf3' },
      { tag: tags.propertyName, color: '#79c0ff' },
      { tag: tags.definition(tags.variableName), color: '#ffa657' },
      { tag: tags.bool, color: '#ff7b72' },
      { tag: tags.null, color: '#7d8590' },
      { tag: tags.special(tags.string), color: '#ffa657' },
      { tag: tags.self, color: '#ff7b72' },
      { tag: tags.moduleKeyword, color: '#ff7b72' },
    ]));

    const kscriptCompletions = this._createCompletions();

    const kscriptAutocomplete = autocompletion({
      override: [
        (context) => {
          const word = context.matchBefore(/\w+\.\w*/);
          if (word) {
            const parts = word.text.split('.');
            const moduleName = parts[0];
            const partial = parts[1] || '';
            if (BUILTIN_MODULES[moduleName]) {
              const options = BUILTIN_MODULES[moduleName]
                .filter(m => m.startsWith(partial))
                .map(m => ({ label: `${moduleName}.${m}`, type: 'function', detail: m }));
              return { from: word.from, options, validFor: /^\w+\.\w*$/ };
            }
          }

          const wordMatch = context.matchBefore(/\w+/);
          if (!wordMatch && !context.explicit) return null;
          const from = wordMatch ? wordMatch.from : context.pos;
          const partial = wordMatch ? wordMatch.text : '';

          const options = [];
          for (const kw of KSCRIPT_KEYWORDS) {
            if (kw.startsWith(partial)) options.push({ label: kw, type: 'keyword' });
          }
          for (const t of KSCRIPT_TYPES) {
            if (t.startsWith(partial)) options.push({ label: t, type: 'type' });
          }
          for (const hook of LIFECYCLE_HOOKS) {
            if (hook.startsWith(partial)) options.push({ label: hook, type: 'function' });
          }
          for (const prim of ASYNC_PRIMITIVES) {
            if (prim.startsWith(partial)) options.push({ label: prim, type: 'function' });
          }
          for (const [modName, methods] of Object.entries(BUILTIN_MODULES)) {
            if (modName.startsWith(partial)) {
              options.push({ label: modName, type: 'namespace', detail: `${methods.length} members` });
            }
          }

          const sceneEntities = window.__sceneEntities ? window.__sceneEntities() : [];
          for (const ent of sceneEntities) {
            const sanitized = ent.replace(/[^a-zA-Z0-9_]/g, '_');
            if (sanitized.startsWith(partial)) {
              options.push({ label: sanitized, type: 'class', detail: 'entity' });
            }
          }

          return { from, options, validFor: /^\w*$/ };
        },
      ],
      closeOnBlur: true,
      defaultKeymap: true,
      icons: false,
    });

    const state = EditorState.create({
      doc: '',
      extensions: [
        lineNumbers(),
        highlightActiveLineGutter(),
        highlightSpecialChars(),
        drawSelection(),
        rectangularSelection(),
        crosshairCursor(),
        history(),
        closeBrackets(),
        keymap.of([
          ...defaultKeymap,
          ...historyKeymap,
          ...completionKeymap,
          ...closeBracketsKeymap,
          indentWithTab,
          { key: 'Ctrl-Enter', run: () => { this._apply(); return true; } },
          { key: 'Shift-Enter', run: () => { this._apply(); return true; } },
        ]),
        kscriptLang,
        kscriptHighlight,
        darkTheme,
        kscriptAutocomplete,
        indentUnit.of('  '),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            this._dirty = true;
            this._changeCallback(this._dirty);
          }
        }),
        EditorView.lineWrapping,
      ],
    });

    this._view = new EditorView({
      state,
      parent: this._container,
    });
    this._ready = true;
    if (this._pendingLoad) {
      this._doLoad(this._pendingLoad.id, this._pendingLoad.name, this._pendingLoad.script);
      this._pendingLoad = null;
    }
  }

  _doLoad(id, name, script) {
    this._entityId = id;
    this._dirty = false;
    if (this._entityNameEl) {
      this._entityNameEl.textContent = name || 'Sem nome';
    }
    this._view.dispatch({
      changes: { from: 0, to: this._view.state.doc.length, insert: script || '' },
    });
    this._view.focus();
    this.clearErrors();
  }

  _createCompletions() {
    const allKeywords = [
      ...KSCRIPT_KEYWORDS,
      ...KSCRIPT_TYPES,
      ...LIFECYCLE_HOOKS,
      ...ASYNC_PRIMITIVES,
    ];

    const completions = allKeywords.map(kw => ({
      label: kw,
      type: KSCRIPT_TYPES.has(kw) ? 'type' : 'keyword',
    }));

    return completions;
  }

  _bindButtons() {
    if (this._applyBtnEl) {
      this._applyBtnEl.addEventListener('click', () => this._apply());
    }
  }

  async _apply() {
    if (!this._view) await this._whenReady();
    const script = this._view.state.doc.toString();
    this._onApply(this._entityId, script);
    this._dirty = false;
    if (this._applyBtnEl) {
      this._applyBtnEl.textContent = '\u2713 Aplicado (Ctrl+Enter)';
      setTimeout(() => {
        if (this._applyBtnEl) this._applyBtnEl.textContent = 'Aplicar (Ctrl+Enter)';
      }, 1500);
    }
  }

  async loadForEntity(id, name, script) {
    await this._whenReady();
    if (this._dirty && this._entityId !== null && this._entityId !== id) {
      if (!confirm('Há alterações n\xe3o salvas neste script. Salvar antes de trocar?')) {
        return false;
      }
      this._apply();
    }
    this._doLoad(id, name, script);
    return true;
  }

  async clearEntity() {
    this._entityId = null;
    this._pendingLoad = null;
    if (this._entityNameEl) {
      this._entityNameEl.textContent = 'Nenhuma entidade selecionada';
    }
    await this._whenReady();
    this._view.dispatch({
      changes: { from: 0, to: this._view.state.doc.length, insert: '' },
    });
    this._dirty = false;
    this.clearErrors();
  }

  getScript() {
    return this._view ? this._view.state.doc.toString() : '';
  }

  isDirty() { return this._dirty; }

  setErrors(errors) {
    this._errors = errors || [];
    this._renderErrors();

    if (this._errorBadgeEl) {
      this._errorBadgeEl.hidden = this._errors.length === 0;
      this._errorBadgeEl.textContent = `${this._errors.length} erro${this._errors.length !== 1 ? 's' : ''}`;
      this._errorBadgeEl.title = this._errors.map(e => e.message).join('\n');
    }
  }

  clearErrors() {
    this._errors = [];
    this._renderErrors();
    if (this._errorBadgeEl) {
      this._errorBadgeEl.hidden = true;
    }
  }

  _renderErrors() {
    if (!this._view) return;

    this._view.contentDOM.querySelectorAll('.cm-errorLine').forEach(el => {
      el.classList.remove('cm-errorLine');
      el.removeAttribute('title');
    });
    const gutters = this._view.scrollDOM.querySelector('.cm-gutters');
    if (gutters) {
      gutters.querySelectorAll('.cm-errorLineGutter').forEach(el => el.remove());
    }

    const cmLines = this._view.contentDOM.querySelectorAll('.cm-line');
    for (const err of this._errors) {
      const idx = (err.line || 1) - 1;
      if (idx >= 0 && idx < cmLines.length) {
        cmLines[idx].classList.add('cm-errorLine');
        cmLines[idx].title = err.message;
      }
    }
  }

  focus() {
    if (this._view) this._view.focus();
  }

  destroy() {
    if (this._view) {
      this._view.destroy();
      this._view = null;
    }
  }
}

window.CodePanel = CodePanel;
