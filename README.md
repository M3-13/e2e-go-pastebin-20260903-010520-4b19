# Pastebin-REST-API

Eine kleine Pastebin-REST-API in Go, ausschließlich mit `net/http` aus der
Standardbibliothek. Pastes lassen sich anlegen, abrufen, auflisten und löschen,
mit In-Memory-Speicher, Mutex-Synchronisierung, Ablaufzeiten, sauberen
Statuscodes und JSON-Fehlern.

## Tech Stack

- **Sprache**: Go (1.22+)
- **HTTP**: `net/http` (Standardbibliothek, Go-1.22-ServeMux)
- **Store**: In-Memory mit `sync.RWMutex`
- **Tests**: `go test` mit `net/http/httptest`
- **Abhängigkeiten**: keine externen

## Installation

Voraussetzung ist eine Go-Toolchain (1.22 oder neuer).

```sh
go mod download
```

## Run in Dev

```sh
go run .
```

Der Server startet dann auf `:8080` (bzw. auf `PASTEBIN_ADDR`).

## Build für Produktion

```sh
go build -o pastebin .
./pastebin
```

## Environment-Variablen

| Variable | Default | Bedeutung |
| --- | --- | --- |
| `PASTEBIN_ADDR` | `:8080` | Adresse, auf der der HTTP-Server lauscht |
| `PASTEBIN_MAX_BODY_BYTES` | `1048576` | Maximale Größe des Request-Bodys für `POST /pastes` in Bytes |

## Endpunkte

| Methode | Pfad | Beschreibung |
| --- | --- | --- |
| `GET` | `/healthz` | Health-Check, antwortet `200 {"status":"ok"}` |
| `POST` | `/pastes` | Legt einen Paste an |
| `GET` | `/pastes/{id}` | Holt einen Paste inkl. `content` |
| `GET` | `/pastes` | Listet aktive Pastes (nur Metadaten, ohne `content`) |
| `DELETE` | `/pastes/{id}` | Löscht einen Paste |

Alle JSON-Antworten (Erfolg und Fehler) setzen `Content-Type: application/json`.
Fehlerantworten enthalten ausschließlich das Feld `error`.

## Features

- Paste anlegen mit Ablaufzeit (`expires_in_seconds`)
- Paste abrufen, auflisten und löschen
- In-Memory-Speicher mit Mutex-Synchronisierung und lazy Ablauf-Entfernung
- Kryptographisch sichere, URL-sichere IDs
- Begrenzung des Request-Bodys auf eine konfigurierbare Maximalgröße
- Health-Endpoint für Liveness-Checks
