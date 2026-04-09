# Guia de Assets — Editor Kora

## Importando Assets

O Editor Kora suporta múltiplos formatos de arquivos.

### Formatos de Imagem

| Formato | Extensão | Uso |
|---------|----------|-----|
| PNG | `.png` | Sprites, ícones, transparencies |
| JPG/JPEG | `.jpg`, `.jpeg` | Texturas, backgrounds |
| WebP | `.webp` | Sprites otimizados (web) |
| GIF | `.gif` | Animações simples |
| SVG | `.svg` | Vetores (escala infinita) |

### Formatos de Áudio

| Formato | Extensão | Uso |
|---------|----------|-----|
| OGG Vorbis | `.ogg` | Músicas e SFX (recomendado) |
| WAV | `.wav` | SFX sem compressão |
| MP3 | `.mp3` | Músicas compatibilidade |

### Formatos de Tilemap

| Formato | Extensão | Uso |
|---------|----------|-----|
| TMJ (Kora) | `.tmj` | Tilemap custom |
| Tiled | `.tmx` | Tiled map editor export |
| JSON | `.json` | Tilemap genérico |

### Importação

1. Abra a aba **Assets** no editor
2. Clique **+ Importar** ou arraste arquivos do sistema
3. Assets aparecem como cards com thumbnail (imagens) ou ícone (áudio)
4. Arraste um asset para o canvas → cria entidade automaticamente
5. Duplo clique no asset → cria entidade no centro da cena

## Gerenciando Assets

### Excluir Asset

- Hover sobre o card do asset
- Clique no botão **"×"** no canto superior direito
- Asset é removido da grid e do IndexedDB

### Filtros

- **Todos**: Mostra todos os assets
- **🖼️ Imagens**: Mostra apenas imagens
- **🔊 Áudio**: Mostra apenas áudios
- **🗺️ Tilemaps**: Mostra apenas tilemaps

## Persistência

Assets são salvos automaticamente em **IndexedDB** (`kora-editor` database).

- Assets persistem entre sessões do navegador
- Dados não são afetados por limpeza de cache comum
- Cada asset armazena: `id`, `name`, `kind`, `url`, `size`, `ext`, `blob`, `contentType`

### Limpar Assets

Para limpar todos os assets:

```javascript
// No console do navegador
const db = await indexedDB.deleteDatabase('kora-editor');
location.reload();
```

## Importando com IndexedDB

O editor usa `AssetDB` (idb.js) para persistência:

```javascript
// Adicionar asset
await AssetDB.add({
  id: 'asset_123',
  name: 'player.png',
  kind: 'image',
  url: 'data:image/png;base64,...',
  size: 4096,
  ext: 'png',
  blob: new Blob([...], {type: 'image/png'}),
  contentType: 'image/png'
});

// Listar todos
const assets = await AssetDB.getAll();

// Deletar
await AssetDB.delete('asset_123');

// Limpar tudo
await AssetDB.clear();
```

## Tamanho e Performance

- Imagens são limitadas a 192x192px na entidade (configurável)
- Thumbnails no painel: 56x56px
- Assets grandes podem demorar para carregar no preview

## Dicas

1. **Exporte sprites com fundo transparente** (PNG com alpha)
2. **Nomes amigáveis** evitam confusão (ex: `player-idle.png` vs `spr001.png`)
3. **Organize em grupos** usando filtros de categoria
4. **Verifique tamanhos** antes de importar (imagens muito grandes = slow load)
5. **Use OGG para áudio** — melhor qualidade/peso

## Troubleshooting

| Problema | Solução |
|----------|---------|
| Asset não importa | Verifique formato e extensão |
| Imagem não mostra thumbnail | Abra console — erro de leitura de arquivo |
| Asset some ao recarregar | IndexedDB pode estar cheio — limpe cache |
| Arrastar não funciona | Use navegador compatível (Chrome, Edge, Firefox) |

---

**Importar sprites e arrastar para a cena** agora é o fluxo principal de composição de níveis no Editor Kora.
