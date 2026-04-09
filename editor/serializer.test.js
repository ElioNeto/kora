/* =============================================================
   Kora Editor — serializer.test.js
   Pure-JS unit tests (no DOM, no framework).
   Run with: node editor/serializer.test.js
============================================================= */
'use strict';

// ---- Inline the module (strip browser globals) ------------------

function sceneToJSON(entities, meta = {}) {
  return {
    kora:     '1.0',
    name:     meta.name    || 'Untitled',
    version:  meta.version || 1,
    logicalW: meta.logicalW || 360,
    logicalH: meta.logicalH || 640,
    entities: entities.map(e => ({
      id: e.id, name: e.name, type: e.type,
      x: e.x, y: e.y, w: e.w, h: e.h,
      rotation: e.rotation, visible: e.visible,
      locked: e.locked, color: e.color, script: e.script || '',
    })),
  };
}

function sanitizeIdent(str) {
  return str.replace(/[^a-zA-Z0-9_]/g, '_').replace(/^([0-9])/, '_$1');
}

function jsonToScene(doc, nextId) {
  if (!doc || doc.kora !== '1.0') throw new Error('Formato inválido');
  let seq = 0;
  const id = nextId || (() => ++seq);
  return {
    entities: (doc.entities || []).map(e => ({
      id: id(), name: e.name || 'Entity', type: e.type || 'custom',
      x: Number(e.x)||0, y: Number(e.y)||0, w: Number(e.w)||48, h: Number(e.h)||48,
      rotation: Number(e.rotation)||0, visible: e.visible !== false,
      locked: !!e.locked, color: e.color||'#00e5a0', script: e.script||'',
    })),
    meta: { name: doc.name||'Untitled', version: doc.version||1, logicalW: doc.logicalW||360, logicalH: doc.logicalH||640 },
  };
}

const KSCRIPT_TYPE_MAP = { sprite:'SpriteEntity', tilemap:'TilemapEntity', camera:'CameraEntity', audio:'AudioEmitter', custom:'Entity' };
function sceneToKScript(entities, sceneName = 'GameScene') {
  const lines = [`// Scene: ${sceneName}`, ''];
  for (const e of entities) {
    const cls = sanitizeIdent(e.name);
    const typ = KSCRIPT_TYPE_MAP[e.type] || 'Entity';
    lines.push(`entity ${cls} : ${typ} {`);
    lines.push(`  x = ${e.x}`, `  y = ${e.y}`, `  width = ${e.w}`, `  height = ${e.h}`);
    if (e.script?.trim()) e.script.trim().split('\n').forEach(l => lines.push('  ' + l));
    lines.push('}', '');
  }
  lines.push(`scene ${sanitizeIdent(sceneName)} {`);
  entities.filter(e => e.visible).forEach(e => lines.push(`  spawn ${sanitizeIdent(e.name)}()`));
  lines.push('}', '');
  return lines.join('\n');
}

// ---- Test runner ------------------------------------------------
let passed = 0, failed = 0;
function test(name, fn) {
  try { fn(); console.log(`  ✅ ${name}`); passed++; }
  catch(e) { console.error(`  ❌ ${name}\n     ${e.message}`); failed++; }
}
function assert(cond, msg) { if (!cond) throw new Error(msg || 'assertion failed'); }
function assertEqual(a, b) { if (a !== b) throw new Error(`expected ${JSON.stringify(b)}, got ${JSON.stringify(a)}`); }

console.log('\nKora Serializer Tests\n');

// ---- sanitizeIdent
test('sanitizeIdent: spaces → underscores', () => assertEqual(sanitizeIdent('My Entity'), 'My_Entity'));
test('sanitizeIdent: leading digit → underscore prefix', () => assertEqual(sanitizeIdent('2Player'), '_2Player'));
test('sanitizeIdent: specials stripped', () => assertEqual(sanitizeIdent('hello!@#'), 'hello___'));
test('sanitizeIdent: already valid', () => assertEqual(sanitizeIdent('Player'), 'Player'));

// ---- sceneToJSON round-trip
const SAMPLE = [
  { id:1, name:'Player', type:'sprite', x:10, y:20, w:32, h:32, rotation:0, visible:true, locked:false, color:'#0f0', script:'' },
  { id:2, name:'Ground', type:'tilemap', x:0, y:200, w:360, h:24, rotation:0, visible:true, locked:true, color:'#00f', script:'' },
];

test('sceneToJSON: kora version field', () => assertEqual(sceneToJSON(SAMPLE).kora, '1.0'));
test('sceneToJSON: entity count', () => assertEqual(sceneToJSON(SAMPLE).entities.length, 2));
test('sceneToJSON: preserves x/y', () => {
  const doc = sceneToJSON(SAMPLE);
  assertEqual(doc.entities[0].x, 10);
  assertEqual(doc.entities[0].y, 20);
});
test('sceneToJSON: default meta name', () => assertEqual(sceneToJSON(SAMPLE).name, 'Untitled'));
test('sceneToJSON: custom meta name', () => assertEqual(sceneToJSON(SAMPLE, { name: 'Level1' }).name, 'Level1'));

// ---- jsonToScene round-trip
test('jsonToScene: re-hydrates entity count', () => {
  const doc = sceneToJSON(SAMPLE);
  const { entities } = jsonToScene(doc);
  assertEqual(entities.length, 2);
});
test('jsonToScene: preserves name', () => {
  const { entities } = jsonToScene(sceneToJSON(SAMPLE));
  assertEqual(entities[0].name, 'Player');
});
test('jsonToScene: invalid doc throws', () => {
  let threw = false;
  try { jsonToScene({ kora: '0.5' }); } catch { threw = true; }
  assert(threw, 'should throw on bad version');
});
test('jsonToScene: locked flag preserved', () => {
  const { entities } = jsonToScene(sceneToJSON(SAMPLE));
  assert(entities[1].locked === true);
});
test('jsonToScene: invisible entity preserved', () => {
  const invisible = [{ ...SAMPLE[0], visible: false }];
  const { entities } = jsonToScene(sceneToJSON(invisible));
  assertEqual(entities[0].visible, false);
});

// ---- sceneToKScript
test('sceneToKScript: contains entity keyword', () => {
  assert(sceneToKScript(SAMPLE).includes('entity Player'));
});
test('sceneToKScript: contains scene keyword', () => {
  assert(sceneToKScript(SAMPLE, 'Level1').includes('scene Level1'));
});
test('sceneToKScript: spawn for visible entity', () => {
  assert(sceneToKScript(SAMPLE).includes('spawn Player()'));
});
test('sceneToKScript: no spawn for invisible entity', () => {
  const inv = [{ ...SAMPLE[0], visible: false }];
  assert(!sceneToKScript(inv).includes('spawn Player()'));
});
test('sceneToKScript: embedded script indented', () => {
  const withScript = [{ ...SAMPLE[0], script: 'on Update(dt) {\n  move(dt)\n}' }];
  const src = sceneToKScript(withScript);
  assert(src.includes('  on Update(dt) {'));
});
test('sceneToKScript: type mapping sprite → SpriteEntity', () => {
  assert(sceneToKScript(SAMPLE).includes('SpriteEntity'));
});
test('sceneToKScript: type mapping tilemap → TilemapEntity', () => {
  assert(sceneToKScript(SAMPLE).includes('TilemapEntity'));
});

console.log(`\n${passed} passed, ${failed} failed\n`);
if (failed > 0) process.exit(1);
