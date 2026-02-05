# go-serverless-event-platform

Das Projekt ist eine serverless Event-Sourcing-Plattform in Go, die Bestellungen verwaltet. HTTP-Requests werden vom Command Handler Lambda verarbeitet, Events im DynamoDB Event Store gespeichert und über EventBridge publiziert. Ein Projection Handler Lambda konsumiert die Events asynchron und erstellt Read Models in DynamoDB für schnelle Abfragen. Die Architektur ist für hohe Event-Volumina optimiert mit Idempotenz auf mehreren Ebenen, Concurrency-Limits für Backpressure, strukturiertem Logging mit Payload-Redaction, CloudWatch-Metriken statt Info-Logs, Retriable/Non-Retriable Fehlerklassifizierung und Dead Letter Queues für fehlgeschlagene Events.

## Setup

```bash
go mod download
make build
```

## Deployment

```bash
make deploy-dev
```

## Tests

```bash
make test
```

## Struktur

- `cmd/` - Lambda Handlers
- `internal/domain/` - Domain Model
- `internal/app/` - Use Cases
- `internal/infra/` - Infrastructure (DynamoDB, EventBridge)
- `pkg/observability/` - Logging & Metrics

## Umgebungsvariablen

- `EVENT_STORE_TABLE` - DynamoDB Event Store Tabelle
- `ORDERS_READ_TABLE` - DynamoDB Read Model Tabelle
- `PROCESSED_EVENTS_TABLE` - DynamoDB Processed Events Tabelle
- `EVENT_BUS_NAME` - EventBridge Bus Name
- `LOG_LEVEL` - Log-Level (ERROR, INFO, WARN, DEBUG), Default: ERROR
- `LOG_LEVEL` - Log-Level (ERROR, INFO, WARN, DEBUG), Default: ERROR
