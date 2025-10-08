#!/bin/bash

# Este script simplifica a execução e compilação da aplicação Wails,
# incluindo as tags de compilação necessárias para ambientes Linux.

# Para o script em caso de erro
set -e

# Cria um diretório temporário local para evitar problemas de permissão com /tmp (noexec)
TMP_DIR=$(pwd)/.tmp
mkdir -p "$TMP_DIR"
chmod 755 "$TMP_DIR" # Garante permissão de execução
export TMPDIR="$TMP_DIR"
 echo "🚀 Diretório temporário criado em: $TMPDIR..."

COMMAND=$1
shift # Remove o primeiro argumento (dev/build) para que $@ contenha o resto

case "$COMMAND" in
  dev)
    echo "🚀 Iniciando servidor de desenvolvimento com a tag 'netgo'..."
    echo "Usando diretório temporário: $TMPDIR"
    # Verifica se o strace está instalado para um log mais detalhado
    if command -v strace &> /dev/null; then
      echo "🔍 Executando com 'strace' para depuração detalhada de permissões..."
      # O -f segue os processos filhos, e -o salva o log em um arquivo
      strace -f -o strace.log wails dev -tags netgo -v 2 "$@"
    else
      wails dev -tags netgo -v 2 "$@"
    fi
    ;;
  build)
    echo "📦 Compilando a aplicação com a tag 'netgo'..."
    echo "Usando diretório temporário: $TMPDIR"
    wails build -tags netgo -v 2 "$@"
    ;;
  *)
    echo "Uso: $0 {dev|build}"
    exit 1
    ;;
esac