VERDICT: CHANGES_REQUESTED

Sicherheitsprüfung des zusammengeführten Produkts. Für diesen Projekttyp (`go-backend`) wurde kein anwendbarer Scanner-Output geliefert; die Beurteilung basiert daher auf manueller Codeanalyse.

## Befunde

### 1. Mittel – Unbegrenztes Speicherwachstum des In-Memory-Stores (Ressourcen-DoS)
**Betroffen:** `store/store.go`, `handlers/handlers.go`, `main.go`

**Problem:** Ein unauthentifizierter Client kann beliebig viele `POST /pastes` ohne Ablaufzeit absetzen. Jedes Paste verbleibt dauerhaft in der Map `s.pastes`. Eine Hintergrundbereinigung existiert nicht; abgelaufene Einträge werden nur bei `Get`, `List` oder `Delete` entfernt. Ein Angreifer kann so den Prozess durch Speichererschöpfung zum Absturz bringen (Verfügbarkeitsverlust).

**Konkreter Fix:**
- Konfigurierbare Limits einführen, z. B. maximale Anzahl Pastes oder maximaler Gesamtspeicher des Stores.
- Bei Überschreitung `429 Too Many Requests` oder `507 Insufficient Storage` zurückgeben.
- Alternativ oder ergänzend eine moderate Standard-TTL für Pastes ohne explizite Ablaufzeit einführen.
- Optional ein konfigurierbares Rate-Limit pro Client-IP oder Prozesstoken. Wichtig: Die IP darf dabei nicht im Paste-Store gespeichert werden (Datenschutz-AC-15 bleibt erfüllt). Die Funktion „Pastes ohne explizite Ablaufzeit“ muss in einem klar begrenzten Rahmen weiter möglich bleiben.

### 2. Mittel – Fehlende Server-Timeouts (Slow-Body / Slow-Response)
**Betroffen:** `main.go` (`http.Server`)

**Problem:** Es wird nur `ReadHeaderTimeout` gesetzt. `ReadTimeout`, `WriteTimeout` und `IdleTimeout` fehlen. Ein Client kann den Request-Body extrem langsam senden und so Verbindungen lange offen halten, was zu Ressourcenerschöpfung führt.

**Konkreter Fix:**
```go
srv := &http.Server{
    Addr:              addr,
    Handler:           newMux(api),
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       15 * time.Second, // groß genug für das konfigurierte Body-Limit
    WriteTimeout:      15 * time.Second,
    IdleTimeout:       60 * time.Second,
}
```
Die Werte müssen mit dem legitimen 1-MiB-Body-Limit vereinbar sein.

### 3. Niedrig – Fehlende Obergrenze für `expires_in_seconds` (Integer-Überlauf)
**Betroffen:** `handlers/handlers.go` (`CreatePaste`)

**Problem:** `req.ExpiresInSeconds` wird nur auf `>0` geprüft. Ein sehr großer Wert kann bei `time.Duration(*req.ExpiresInSeconds) * time.Second` einen `int64`-Überlauf verursachen. Der Paste wird dann fälschlich als sofort abgelaufen behandelt. Für andere Nutzer nicht unmittelbar sicherheitskritisch, aber ein Zuverlässigkeitsproblem.

**Konkreter Fix:**
```go
const maxExpirySeconds = 315_576_000 // z. B. 10 Jahre
if *req.ExpiresInSeconds > maxExpirySeconds {
    writeError(w, http.StatusUnprocessableEntity, "expires_in_seconds exceeds maximum allowed value")
    return
}
```

## Positiv geprüft
- Crypto-rand-basierte ID mit >=128 Bit Entropie und URL-sicherem Alphabet (`idgen/idgen.go`).
- `http.MaxBytesReader` begrenzt den Request-Body; Überschreitung liefert korrekt `413` (`handlers/handlers.go`).
- Fehlerantworten enthalten ausschließlich das JSON-Feld `error`; keine Stacktraces oder internen Fehlermeldungen.
- Alle JSON-Antworten setzen `Content-Type: application/json`.
- Keine Speicherung von Client-IP, User-Agent oder sonstigen Request-Headern im Store.
- Mutex-Schutz (`sync.RWMutex`) verhindert Datenverlust/Races im Store.
- Keine hartkodierten Secrets, keine Injection, keine unsichere AuthN/AuthZ erkennbar, keine bekannte verwundbare Drittanbieter-Abhängigkeit.