# Kora Editor Desktop

## 🎮 Engine de Jogos 2D para Android

Kora é uma engine de jogos 2D inspirada no GameMaker Studio, projetada especificamente para criar jogos nativos para Android. Utilizamos uma linguagem de script própria, o **KScript**, que é compilada para Go e depois para código nativo Android - sem VM, sem overhead.

## ✨ Características

### Editor Desktop
- **Editor visual** - Crie cenas arrastando e soltando entidades
- **Importação de assets** - PNG, JPG, WebP, OGG, WAV
- **Preview em tempo real** - Teste sua cena instantaneamente
- **Inspector completo** - Edite propriedades de entidades
- **Hierarquia** - Gerencie suas entidades de forma organizada
- **Console de debug** - Logs e mensagens em tempo real

### KScript - Nossa Linguagem
- **Sintaxe TypeScript-like** - Familiar e poderosa
- **Tipagem estática** - Erros detectados em tempo de compilação
- **Async/await nativo** - Corrotinas para lógica não-blocking
- **Compilação AOT** - Performance máxima, sem VM

```kscript
object Player {
  speed: float = 180
  
  update(dt: float) {
    const move = Input.axisX()
    this.x += move * this.speed * dt
  }
}
```

### Export para Android
- **APK nativo** - Compile para Android com um clique
- **Sem WebView** - Runtime 100% nativo em Go
- **Otimizado** - Código Go compilado para ARM64

## 🚀 Quickstart

### Pré-requisitos

- **Node.js 18+**
- **Go 1.22+**
- **Android SDK** (para build APK)

### Instalação

```bash
# Clone o repositório
git clone https://github.com/koraengine/kora.git
cd kora

# Instalare dependências
cd apps/desktop
npm install

# Inicie o editor
npm run dev
```

### Criando seu Primeiro Jogo

1. **Abra o Kora Editor Desktop**
2. **Importe um sprite** (File → Import → Selecione PNG)
3. **Arraste para a cena**
4. **Adicione script KScript:**
```kscript
on Update(dt) {
  this.x += 100 * dt
}
```
5. **Exporte como APK**
6. **Instale no Android**

## 📚 Documentação

- [Guia do Editor](./docs/EDITOR_GUIDE.md)
- [Linguagem KScript](./docs/SCRIPT.md)
- [API Reference](./docs/API_REFERENCE.md)
- [Assets Guide](./docs/ASSETS_GUIDE.md)
- [Desktop App](./docs/DESKTOP_APP.md)

## 🏗️ Build

### Development

```bash
# Editor desktop (Electron)
cd apps/desktop
npm install
npm run dev

# Compilador KScript
cd compiler
go build

# Runtime
go run main.go
```

### Release

```bash
# Build desktop app
cd apps/desktop
npm run build

# Build APK (Android)
# Requer ANDROID_HOME configurado
./android/build.sh release
```

## 📊 Arquitetura

```
┌─────────────────────────────────────────────────────┐
│           Kora Editor Desktop (Electron)            │
│  Scene Editor · Asset Management · Inspector        │
└─────────────────────┬───────────────────────────────┘
                      │ KScript (.ks)
                      ▼
┌─────────────────────────────────────────────────────┐
│              KScript Compiler (Go)                  │
│  Lexer · Parser · Type Checker · Go Emitter         │
└─────────────────────┬───────────────────────────────┘
                      │ Go generated code
                      ▼
┌─────────────────────────────────────────────────────┐
│           Kora Runtime (Go + gomobile)              │
│  Render · Input · Physics · Scene · Asset Loader    │
└─────────────────────┬───────────────────────────────┘
                      │ Native ARM64
                      ▼
              Android APK / AAB
```

## 🎯 Recursos

| Feature | Desktop | Web | Status |
|---------|---------|-----|--------|
| Editor Visual | ✅ | ✅ | Done |
| Asset Import | ✅ | ✅ | Done |
| KScript Editor | ✅ | 🚧 | v1.0 |
| Physics Preview | ✅ | ✅ | Done |
| APK Export | ✅ | ❌ | Desktop only |
| Offline Support | ✅ | ✅ | Done |
| File System Access | ✅ | ❌ | Done |
| Multi-window | 🔜 | ❌ | Roadmap |

## 🔧 Tecnologias

| Camada | Tecnologia |
|--------|------------|
| Editor Desktop | Electron 28 + Vite |
| Scripting | KScript (custom, Go-compiled) |
| Runtime Core | Go 1.22+ |
| Export | gomobile + Android SDK |
| Preview | HTML5 Canvas + Web Audio |

## 📝 Workflow

1. **Desenvolver** no Kora Editor Desktop
2. **Testar** no Preview integrado
3. **Salvar** cena (.kora.json)
4. **Exportar** KScript (.ks)
5. **Compilar** para APK
6. **Deploy** para Android

## 🤝 Contribuindo

Veja [CONTRIBUTING.md](./docs/CONTRIBUTING.md) para como contribuir com o projeto.

## 💡 Exemplo Rápido

```bash
# 1. Abra o editor
cd apps/desktop
npm install
npm run dev

# 2. Importe assets
# 3. Arraste sprites para cena
# 4. Adicione KScript

# 5. Exporte
# File → Export APK
# ou use terminal:
go build -o bin/kora-compiler ./compiler
bin/kora-compiler game.ks

# 6. Build APK
./android/build.sh release
```

## 📄 License

MIT License - See [LICENSE](./LICENSE)

---

**Kora Engine** - Crie jogos para Android com poder nativo e desenvolvimento simplificado.
