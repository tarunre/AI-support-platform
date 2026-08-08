run:
	@pkill -9 -f "/tmp/supportiq" 2>/dev/null || true
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@sleep 1
	@go build -o /tmp/supportiq-api ./cmd/ && go build -o /tmp/supportiq-worker ./cmd/worker/
	@/tmp/supportiq-api > /tmp/api.log 2>&1 &
	@/tmp/supportiq-worker > /tmp/worker.log 2>&1 &
	@sleep 4 && curl -s http://localhost:8080/api/v1/health && echo ""

stop:
	@lsof -ti:8080 | xargs kill -9 2>/dev/null; pkill -9 -f "/tmp/supportiq" 2>/dev/null; echo "Stopped"

.PHONY: run stop
