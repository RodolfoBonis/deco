#!/bin/bash

# Script para escapar caracteres especiais para Telegram MarkdownV2
# Uso: ./escape_telegram_message.sh "mensagem com caracteres especiais"

escape_telegram_message() {
    local message="$1"
    
    # Escapa caracteres especiais do MarkdownV2
    message=$(echo "$message" | sed 's/\\/\\\\/g')
    message=$(echo "$message" | sed 's/_/\\_/g')
    message=$(echo "$message" | sed 's/\*/\\*/g')
    message=$(echo "$message" | sed 's/\[/\\[/g')
    message=$(echo "$message" | sed 's/\]/\\]/g')
    message=$(echo "$message" | sed 's/(/\\(/g')
    message=$(echo "$message" | sed 's/)/\\)/g')
    message=$(echo "$message" | sed 's/~/\\~/g')
    message=$(echo "$message" | sed 's/`/\\`/g')
    message=$(echo "$message" | sed 's/>/\\>/g')
    message=$(echo "$message" | sed 's/#/\\#/g')
    message=$(echo "$message" | sed 's/+/\\+/g')
    message=$(echo "$message" | sed 's/-/\\-/g')
    message=$(echo "$message" | sed 's/=/\\=/g')
    message=$(echo "$message" | sed 's/|/\\|/g')
    message=$(echo "$message" | sed 's/{/\\{/g')
    message=$(echo "$message" | sed 's/}/\\}/g')
    message=$(echo "$message" | sed 's/\./\\./g')
    message=$(echo "$message" | sed 's/!/\\!/g')
    
    echo "$message"
}

# Se executado diretamente, escapa a mensagem passada como argumento
if [ $# -eq 1 ]; then
    escape_telegram_message "$1"
fi 