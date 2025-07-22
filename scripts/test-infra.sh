#!/bin/bash

# Script para gerenciar infraestrutura de teste (Redis + OpenTelemetry)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

case "$1" in
    "start")
        echo "🚀 Iniciando infraestrutura de teste..."
        docker-compose -f docker-compose.test.yml up -d
        
        echo "⏳ Aguardando serviços ficarem prontos..."
        echo "Redis..."
        timeout 30 bash -c 'until docker exec deco-redis-test redis-cli ping > /dev/null 2>&1; do sleep 1; done'
        
        echo "OpenTelemetry Collector..."
        timeout 30 bash -c 'until curl -s http://localhost:13133 > /dev/null 2>&1; do sleep 1; done'
        
        echo "✅ Infraestrutura de teste iniciada com sucesso!"
        echo "📊 Redis: localhost:6379"
        echo "📊 OpenTelemetry: localhost:4318 (HTTP), localhost:4317 (gRPC)"
        ;;
        
    "stop")
        echo "🛑 Parando infraestrutura de teste..."
        docker-compose -f docker-compose.test.yml down
        echo "✅ Infraestrutura de teste parada!"
        ;;
        
    "restart")
        echo "🔄 Reiniciando infraestrutura de teste..."
        docker-compose -f docker-compose.test.yml down
        docker-compose -f docker-compose.test.yml up -d
        
        echo "⏳ Aguardando serviços ficarem prontos..."
        echo "Redis..."
        timeout 30 bash -c 'until docker exec deco-redis-test redis-cli ping > /dev/null 2>&1; do sleep 1; done'
        
        echo "OpenTelemetry Collector..."
        timeout 30 bash -c 'until curl -s http://localhost:13133 > /dev/null 2>&1; do sleep 1; done'
        
        echo "✅ Infraestrutura de teste reiniciada!"
        ;;
        
    "status")
        echo "📊 Status da infraestrutura de teste:"
        echo ""
        echo "Redis:"
        if docker exec deco-redis-test redis-cli ping > /dev/null 2>&1; then
            echo "  ✅ Rodando (localhost:6379)"
        else
            echo "  ❌ Parado"
        fi
        
        echo ""
        echo "OpenTelemetry Collector:"
        if curl -s http://localhost:13133 > /dev/null 2>&1; then
            echo "  ✅ Rodando (localhost:4318)"
        else
            echo "  ❌ Parado"
        fi
        ;;
        
    "logs")
        echo "📋 Logs da infraestrutura de teste:"
        docker-compose -f docker-compose.test.yml logs -f
        ;;
        
    *)
        echo "Uso: $0 {start|stop|restart|status|logs}"
        echo ""
        echo "Comandos:"
        echo "  start   - Inicia a infraestrutura de teste"
        echo "  stop    - Para a infraestrutura de teste"
        echo "  restart - Reinicia a infraestrutura de teste"
        echo "  status  - Mostra o status dos serviços"
        echo "  logs    - Mostra os logs dos serviços"
        exit 1
        ;;
esac 