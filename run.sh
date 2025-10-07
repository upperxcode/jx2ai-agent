#!/bin/bash

# Este script simplifica a execução e compilação da aplicação Wails,
# incluindo as tags de compilação necessárias para ambientes Linux.

# Para o script em caso de erro
set -e

COMMAND=$1
shift # Remove o primeiro argumento (dev/build) para que $@ contenha o resto

case "$1" in
  dev)
    echo "🚀 Iniciando servidor de desenvolvimento com a tag 'netgo'..."
    wails dev -tags netgo "$@"
    ;;
  build)
    echo "📦 Compilando a aplicação com a tag 'netgo'..."
    wails build -tags netgo "$@"
    ;;
  *)
    echo "Uso: $0 {dev|build}"
    exit 1
    ;;
esac