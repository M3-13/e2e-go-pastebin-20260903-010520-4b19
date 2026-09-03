VERDICT: CHANGES_REQUESTED

## 1. DSGVO

### D1 – Keine Obergrenze für die Speicherdauer
- **Schweregrad:** mittel
- **Befund:** `CreatePaste` akzeptiert `expires_in_seconds` nur als positiven Wert, setzt aber keine maximale Ablaufzeit. Ohne Angabe bleibt ein Paste unbegrenzt im In-Memory-Store. Das kollidiert mit dem Grundsatz der Speicherbegrenzung aus Art. 5 Abs. 1 lit. e DSGVO, sofern kein Zweck eine unbefristete Speicherung rechtfertigt.
- **Abhilfe:** In `main.go` eine neue Umgebungsvariable `PASTEBIN_MAX_EXPIRY_SECONDS` lesen, z. B. Standard `2592000` (30 Tage). Den Wert an `handlers.API` übergeben und in `CreatePaste` prüfen: Ist `req.ExpiresInSeconds == nil`, einen konfigurierbaren Standard-TTL verwenden; ist der Wert größer als das Maximum, mit `422` ablehnen oder auf das Maximum begrenzen. In `README.md` die gewählte Speicherdauer dokumentieren.

### D2 – Abgelaufene Pastes werden nur bei Zugriff entfernt
- **Schweregrad:** mittel
- **Befund:** `store.Get`, `Store.List` und `Store.Delete` entfernen abgelaufene Einträge erst beim jeweiligen Zugriff. Ohne Zugriff bleiben personenbezogene Daten nach Ablauf weiter im Arbeitsspeicher, bis der Prozess endet. Damit ist keine vollständige Löschung nach Ablauf garantiert.
- **Abhilfe:** In `store/store.go` eine Methode `CleanupExpired(now time.Time)` ergänzen, die unter Lock alle abgelaufenen Einträge löscht. In `main.go` einen `time.Ticker` starten, z. B. alle 60 Sekunden, der `CleanupExpired` aufruft, und den Ticker beim Graceful Shutdown stoppen. Alternativ eine Methode `StartCleanup(ctx, interval)` am Store anbieten.

### D3 – Löschung beruht ausschließlich auf geheimer ID
- **Schweregrad:** niedrig
- **Befund:** `DELETE /pastes/{id}` löscht jeden Paste, wenn die ID bekannt ist. Das ist datenschutzrechtlich vertretbar, weil die 132-Bit-ID als Bearer-Token wirkt und schwer zu erraten ist. Es muss aber dokumentiert sein, dass die ID nicht in Logs, Fehlertexte oder Referrer gelangen darf.
- **Abhilfe:** In `README.md` oder `SECURITY.md` dokumentieren, dass die ID einem Lese-/Löschtoken entspricht und vertraulich behandelt werden muss. Optional: Bei der Erstellung einen separaten `delete_token` erzeugen und `DELETE` nur mit diesem Token erlauben. Kein zwingender Umbau.

### D4 – Datenschutzhinweise des Betreibers nicht im Projekt sichtbar
- **Schweregrad:** mittel
- **Befund:** Die API hat keine UI und kann selbst keine Datenschutzerklärung ausliefern. Der Verantwortliche muss dennoch Zweck, Rechtsgrundlage, Speicherdauer und Betroffenenrechte bereitstellen.
- **Abhilfe:** In `README.md` einen Abschnitt „DSGVO/Datenschutz“ ergänzen: Verantwortlicher, Rechtsgrundlage z. B. Art. 6 Abs. 1 lit. b oder f DSGVO, Zweck des Dienstes, Speicherdauer, Löschmöglichkeit, Betroffenenrechte und Kontakt. Der vorgeschaltete Client oder Dienst muss diese Erklärung anzeigen.

## 2. EU Cyber Resilience Act (CRA)

### C1 – Unsicherer Standard-Bind auf `:8080`, kein TLS
- **Schweregrad:** mittel
- **Befund:** `PASTEBIN_ADDR` hat den Default `:8080`, bindet also auf allen Interfaces. Der Server spricht Klartext-HTTP. Bei direkter Exposition ist das ein Sicherheitsmangel nach dem CRA-Grundsatz „security by design/default“.
- **Abhilfe:** In `main.go` den Default von `PASTEBIN_ADDR` auf `127.0.0.1:8080` ändern. Optional `PASTEBIN_TLS_CERT` und `PASTEBIN_TLS_KEY` einführen; wenn beide gesetzt sind, `srv.ListenAndServeTLS(certFile, keyFile)` verwenden. Im `README.md` klarstellen, dass die API öffentlich nur hinter einem TLS-terminierenden Reverse Proxy betrieben werden darf.

### C2 – Keine sichtbare SBOM-/Dependency-Erklärung
- **Schweregrad:** niedrig
- **Befund:** Im sichtbaren Code gibt es keine SBOM oder einen dokumentierten Abhängigkeitsnachweis. Da aktuell offenbar nur die Go-Standardbibliothek genutzt wird, ist das Risiko gering; für CRA-Konformität sollte die Abhängigkeitsliste trotzdem nachgewiesen werden.
- **Abhilfe:** Im Build oder in der CI `go list -m -json all` ausführen und das Ergebnis als SBOM-Artefakt ablegen. Falls künftig externe Module verwendet werden, `go.sum` einchecken. In `SECURITY.md` festhalten, dass derzeit nur die Go-Standardbibliothek verwendet wird.

### C3 – Sicherheitsdokumentation nicht sichtbar
- **Schweregrad:** niedrig
- **Befund:** Sicherheitsmodell, Bedrohungsanalyse und Update-/Patch-Prozess sind im gezeigten Code nicht dokumentiert. CRA verlangt dokumentierte Sicherheitseigenschaften und eine nachvollziehbare Update-Fähigkeit.
- **Abhilfe:** Eine `SECURITY.md` anlegen mit Sicherheitsmodell, Bedrohungen, Gegenmaßnahmen und Patch-/Release-Prozess. Darin mindestens dokumentieren: 132-Bit-ID, MaxBodyBytes, keine PII-Logs, TLS-Erfordernis, SBOM-Verweis. Die Datei in `README.md` verlinken.

### C4 – Kein globales Speicher-/Mengenlimit
- **Schweregrad:** mittel
- **Befund:** Der In-Memory-Store akzeptiert unbegrenzt viele Pastes. Ein Angreifer kann mit vielen kleinen gültigen Requests den Arbeitsspeicher erschöpfen und den Dienst lahmlegen. Das betrifft die Verfügbarkeit und den CRA-Grundsatz „security by default“.
- **Abhilfe:** In `store.Store` ein Maximum `MaxPastes` einführen; `Create` gibt bei Überschreitung einen Fehler wie `ErrStoreFull` zurück. In `handlers.CreatePaste` diesen Fehler auf `503` oder `507` mappen. Kein IP-basiertes Rate-Limit im Backend einführen, weil das personenbezogene IP-Adressen verarbeiten würde; ein Mengenlimit ist datenschutzfreundlicher.

### C5 – ID-Pfadparameter nicht validiert
- **Schweregrad:** niedrig
- **Befund:** `GetPaste` und `DeletePaste` nehmen jede beliebige Zeichenkette als `id` an. Kein direktes Datenschutzrisiko, aber unerwartete Eingaben sollten früh abgewiesen werden.
- **Abhilfe:** In `handlers/handlers.go` eine Funktion `isValidID(id string) bool` ergänzen, die Länge 22 und das Alphabet `A-Za-z0-9-_` prüft. Bei ungültiger ID direkt `404 paste not found` zurückgeben, bevor der Store belastet wird.

### C6 – Keine vollständigen Server-Timeouts
- **Schweregrad:** mittel
- **Befund:** `http.Server` setzt nur `ReadHeaderTimeout: 5 * time.Second`. `ReadTimeout`, `WriteTimeout` und `IdleTimeout` fehlen. Das erleichtert Slowloris-artige Angriffe oder Ressourcenblockaden durch langsame Request-Bodies.
- **Abhilfe:** In `main.go` am `http.Server` zusätzlich setzen, z. B.:
  - `ReadTimeout: 15 * time.Second`
  - `WriteTimeout: 15 * time.Second`
  - `IdleTimeout: 60 * time.Second`

## 3. EU AI Act
Nicht anwendbar. Im sichtbaren Code ist keine KI-Funktion oder KI-basierte Verarbeitung enthalten.

## 4. Pflichttexte & UI
Nicht anwendbar. Reine Backend-API ohne öffentliche Endnutzer-UI; Impressum, Cookie-Banner und Barrierefreiheitspflichten gelten nicht für diese Codebasis. Die DSGVO-Pflichttexte des Betreibers müssen jedoch außerhalb der API bereitgestellt werden, siehe D4.

## 5. Barrierefreiheit
Nicht anwendbar. Keine öffentliche Web-UI vorhanden.