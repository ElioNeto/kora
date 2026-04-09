#!/bin/bash
# Kora Editor Desktop - Setup Script

set -e

echo "🚀 Configurando Kora Editor Desktop..."
echo ""

# Verificar Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js não encontrado! Instale Node.js 18+"
    exit 1
fi

echo "✅ Node.js: $(node -v)"

# Criar pastas necessárias
mkdir -p src/renderer/assets

# Copiar assets do editor original
echo "📁 Copiando assets do editor..."
cp ../editor/style.css src/renderer/assets/

# Instalar dependências
echo "📦 Instalando dependências..."
npm install

echo ""
echo "✅ Setup completo!"
echo ""
echo "Para desenvolver:"
echo "  npm run electron:dev"
echo ""
echo "Para construir:"
echo "  npm run electron:build"
echo ""
