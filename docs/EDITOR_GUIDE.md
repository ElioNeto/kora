# Guia do Editor Kora

O Editor Kora é uma ferramenta visual para criação de cenas e entidades 2D.

## Interface

```
┌────────────────────────────────────────────────┐
│  Kora Editor                    [+ Nova] [Abrir]│  ← Topbar
├──────────────────────┬──────────────────┬──────┤
│  HIERARQUIA          │  CANVAS          │INFO  │
│                      │                  │Zoom  │ │
│  ● Player            │  +────────────+  │ 100% │ │
│  ○ Ground            │  │            │  │      │ │
│  ○ MainCamera        │  │    CENA    │  │      │ │
│                      │  │            │  │      │ │
│                      │  +────────────+  │      │ │
├──────────────────────┴──────────────────┴──────┤
│  CONSOLE: [Limpar]                             │  ← Console
│  [12:34] Cena de exemplo carregada.           │
└────────────────────────────────────────────────┘
```

## Abaras

### 1. Cena (Scene)

Área de edição principal onde você posiciona entidades.

**Ferramentas de visualização:**
- `👈 Select` (V): Selecionar entidades
- `✱ Move` (G): Mover entidades
- `⊞ Scale` (S): Escalar entidades
- `⎕ Fit` (F): Encastrar zoom na cena

**Interatividade:**
- **Clicar** em uma entidade → seleciona
- **Arrastar** → move a entidade
- **Wheel** → zoom (0.1x a 4x)
- **Delete/Backspace** → remove entidade selecionada

### 2. Preview

Visualiza a cena em tempo real com física 2D embutida.

**Ativar:** Clique em **Preview** ou pressione **F5**

**Ações disponíveis:**
- Iniciar/Parar jogo
- Ver inputs em tempo real
- Ver física, colisão, gravidade

### 3. Assets

Biblioteca de assets importados.

**Importar:**
- Clique **+ Importar** ou
- Arraste arquivos do sistema

**Arrastar para cena:**
1. Selecione um asset no painel
2. Arraste para o canvas
3. Solte → entidade criada automaticamente

**Excluir:** Hover no card → clique **"×"**

**Filtros:**
- **Todos** | **🖼️ Imagens** | **🔊 Áudio** | **🗺️ Tilemaps**

### 4. Script

Editor KScript para lógica das entidades.

**Em construção (#5)** — em breve você editará KScript diretamente.

## Inspetor (Direita)

Edita propriedades da entidade selecionada.

### Propriedades

| Seção | Campos |
|-------|--------|
| **Identidade** | Nome, Tipo, Asset |
| **Transform** | X, Y, Largura, Altura, Rotação |
| **Visual** | Cor, Visível, Travado |
| **Script KScript** | Área de texto para lógica |
| **Ações** | Duplicar, Excluir |

## Console (Inferior)

Logs do editor com timestamps.

- **Verde:** Sucesso
- **Amarelo:** Aviso
- **Vermelho:** Erro

**Clear:** Botão "Limpar"

## Modals

### Adicionar Entidade

**Ativar:** Bot **+** na hierarquia

```
┌─────────────────────────┐
│ Adicionar entidade      │
├─────────────────────────┤
│ Nome: [Player         ]│
│ Tipo: [Sprite         ]│
│                    [X]  │
│ [Cancelar] [Adicionar]  │
└─────────────────────────┘
```

**Opções de Tipo:**
- **Sprite**: Sprite/Imagem
- **Tilemap**: Tilemap/Grid
- **Camera**: Câmera da cena
- **AudioEmitter**: Emissor de áudio
- **Custom**: Entidade genérica

## Atalhos de Teclado

| Tecla | Ação |
|-------|------|
| `Ctrl+N` | Nova cena |
| `Ctrl+O` | Abrir cena |
| `Ctrl+S` | Salvar cena |
| `F5` | Rodar preview |
| `Delete/Backspace` | Excluir entidade |
| `V` | Ferramenta Select |
| `G` | Ferramenta Move |
| `S` | Ferramenta Scale |
| `F` | Zoom fit |
| `Esc` | Deselecionar |

## Exportação

### JSON

**Atalho:** `Ctrl+S`

Salva cena no formato `.kora.json` (estrutura de entidades).

### KScript

**Botão:** `.ks` na topbar

Gera arquivo `.ks` executável pelo runtime Go.

**Uso:**
```bash
go run cmd/build.go -o jogo cena.ks
```

## Fluxo de Trabalho

### Criar Nova Cena

1. `Ctrl+N` ou clique **+ Nova**
2. Abra aba **Assets**
3. Importe sprites (PNG) e áudios (OGG/WAV)
4. Arraste assets para o canvas
5. Posicione e edite propriedades no Inspetor
6. Adicione lógica KScript (se necessário)
7. `Ctrl+S` para salvar
8. **Preview** para testar

### Modificar Cena Existente

1. `Ctrl+O` para abrir `.kora.json`
2. Arraste novas entidades do Assets
3. Seleções múltiplas (segure Shift)
4. Duplicate entidades (Inspector → Duplicar)
5. Salve com `Ctrl+S`

### Exportar Jogo

1. Prepare cena no editor
2. Botão **.ks** na topbar
3. Compile: `go run cmd/build.go -o jogo scena.ks`
4. APK: `./android/build.sh release`

## Dicas

1. **Use nomes descritivos** ("Player", "GroundLeft", "SFXJump")
2. **Crie grupos lógicos** na hierarquia (ordena entidades)
3. **Teste frequentemente** (F5)
4. **Salve versões** (cena_v1.json, cena_v2.json)
5. **Importe assets grandes** → editor otimiza automat.

## Troubleshooting

| Problema | Solução |
|----------|---------|
| Canvas em branco | Ajuste zoom com wheel |
| Entidade não aparece | Verifique "Visível" no inspector |
| Drag não funciona | Use Chrome/Edge/Firefox |
| Asset não aparece | Abra console — erro de carregamento |
| Salva mas não abre | Verifique console — erro de JSON parse |

---

**Editor v4** • Assets com IndexedDB • Drag-drop • Preview físico
