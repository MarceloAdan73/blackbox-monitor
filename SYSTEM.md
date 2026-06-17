# SYSTEM.md — BlackBox Monitor

> Documento maestro del proyecto. Leer esto primero para entender el estado actual, la arquitectura y los proximos pasos.

---

## 1. IDENTIDAD DEL PROYECTO

- **Nombre:** BlackBox Monitor
- **Version actual:** v1.3 (SSL, Healthcheck, Export CSV, Diseño oscuro)
- **Proposito:** Monitoreo de salud web con dashboard visual en terminal y navegador
- **Lenguaje:** Go 1.22+
- **Autor:** Marcelo Adan
- **Repositorio:** `github.com/marcelo-adan/blackbox-monitor`
- **Modulo Go:** `github.com/marcelo-adan/blackbox-monitor`

---

## 2. STACK TECNOLOGICO

| Componente | Tecnologia | Uso |
|---|---|---|
| Lenguaje | Go 1.22 | Binario unico, stdlib completa |
| UI Terminal | `charmbracelet/lipgloss v1.1.0` | Estilos modernos en terminal |
| Dashboard Web | Bootstrap 3 + ApexCharts | Panel visual en navegador |
| Configuracion | `gopkg.in/yaml.v3` | Config YAML flexible |
| Almacenamiento | `mattn/go-sqlite3` (build tag) | Historial de checks |
| Embed | `embed` (stdlib) | Embeber HTML/CSS/JS en binario |
| Notificaciones | `net/http` (stdlib) | Bot API de Telegram |

---

## 3. ARQUITECTURA

### 3.1. Estructura de directorios

```
blackbox-monitor/
├── main.go                 # Entry point, flags, loop, señales
├── server.go               # Servidor HTTP + endpoint /api/status (embed)
├── state.go                # Estado compartido thread-safe (DashboardState)
├── storage_entry.go        # Interfaz Store + tipo LogEntry
├── storage_sqlite.go       # Implementacion SQLite (build tag: sqlite)
├── storage_nosqlite.go     # Implementacion no-op (build tag: !sqlite)
├── go.mod / go.sum         # Dependencias
├── config.yaml             # Configuracion de sitios (interval + sites) — ignorado por git
├── config.example.yaml     # Ejemplo de configuracion para compartir en GitHub
├── bin/                    # Ejecutable compilado
├── internal/
│   ├── monitor/
│   │   └── checker.go      # CheckSite() - verificacion HTTP GET
│   ├── notifier/
│   │   └── telegram.go     # SendStateChangeAlert() - notificaciones Telegram
│   ├── storage/
│   │   └── sqlite.go       # Store SQLite con migracion auto
│   └── ui/
│       └── styles.go       # RenderTitle, RenderSiteBox, RenderSummaryBox, etc.
└── web/
    └── static/
        ├── index.html      # Dashboard web (Bootstrap + ApexCharts + AJAX)
        ├── assets/css/     # style.css del template
        ├── assets/js/      # custom.js del template
        ├── images/         # Iconos y logos
        └── plugins/        # Themify icons
```

### 3.2. Flujo de ejecucion

```
[Inicio]
   │
   ▼
Parse flags: -config, -interval, -port
   │
   ▼
loadConfig(path) → Config{Interval, Sites[]}
   ├── Error → RenderError() + os.Exit(1)
   │
   ▼
openStore("blackbox.db") → Store (SQLite o no-op)
   │
   ▼
state = NewDashboardState()
   │
   ▼
if -port != "" → go startServer(port, state)
   │                    ├── GET / → index.html (embed)
   │                    ├── GET /static/* → archivos estaticos
   │                    └── GET /api/status → JSON del estado
   │
   ▼
Ticker = time.NewTicker(interval * second)
   │
   ▼
runChecks(cfg, store, state, prevStates)
   ├── Por cada site:
   │   ├── monitor.CheckSite(name, url, timeout)
   │   │   ├── HTTP GET con timeout
   │   │   ├── Mide latencia
   │   │   └── Devuelve SiteResult
   │   ├── RenderSiteBox() → terminal
   │   ├── store.SaveLog() → SQLite
   │   ├── Comparar con prevStates (detectar cambio Online↔Offline)
   │   │   └── Si cambió y Telegram habilitado → go notifier.SendStateChangeAlert() (goroutine)
   │   └── Agregar a siteStatuses[]
   ├── RenderSummaryBox() → terminal
   └── state.Update(sites, online, total, avg, checks, failures)
   │
   ▼
Loop:
   ├── ticker.C → runChecks() de nuevo (limpia pantalla)
   └── sigChan → RenderShutdown() + return
```

### 3.3. Paquete `internal/monitor` (checker.go)

- **Struct `SiteResult`**: Name, URL, Online (bool), HTTPCode, LatencyMs, Error, CheckedAt
- **`CheckSite(name, url string, timeoutMs int) SiteResult`**:
  - HTTP GET con timeout configurable
  - Status 200-399 = Online
  - Status >= 400 o error de red = Offline
  - Mide latencia con `time.Since()`

### 3.4. Paquete `internal/ui` (styles.go)

Funciones de renderizado:
- `RenderTitle()` → Header con borde doble, titulo + subtitle
- `RenderSiteBox(name, url, online, code, latency, time, err)` → Caja por sitio con borde coloreado
- `RenderSummaryBox(online, total, avg, ok, fail)` → Dashboard con barra de progreso y metricas
- `RenderFooter(nextCheck, interval)` → Proximo chequeo + countdown
- `RenderShutdown()` → Mensaje de despedida
- `RenderError(msg)` → Error con borde rojo

Colores:
- `Purple` (#7C3AED) — Titulo, bordes principales
- `Green` (#10B981) — Online, latencia < 300ms
- `Red` (#EF4444) — Offline, latencia > 1000ms
- `Yellow` (#F59E0B) — Latencia 300-1000ms
- `Cyan` (#06B6D4) — Proximo chequeo, dashboard
- `Orange` (#F97316) — Errores
- `Gray` (#6B7280) — Texto secundario

### 3.5. Paquete `internal/storage` (sqlite.go)

- Build tag: `//go:build sqlite`
- Tabla `check_logs`: id, name, url, online, http_code, latency_ms, error_msg, checked_at
- Funciones: `New(dbPath)`, `SaveLog(entry)`, `GetRecentLogs(limit)`, `Close()`
- Requiere CGO + gcc

### 3.6. Storage facade (root level)

- `storage_entry.go` — Interfaz `Store` + tipo `LogEntry` (sin build tags)
- `storage_sqlite.go` — Implementacion real con `//go:build sqlite`
- `storage_nosqlite.go` — No-op con `//go:build !sqlite`
- `openStore(path)` devuelve `Store` dependiendo del build tag

### 3.7. Server web (server.go)

- Embebe `web/static/` con `//go:embed web/static`
- `startServer(addr, state, store)` lanza `http.ListenAndServe`
- Endpoints:
  - `GET /` → Sirve `index.html`
  - `GET /static/*` → Archivos estaticos (CSS, JS, imagenes)
  - `GET /api/status` → JSON con estado actual de todos los sitios
  - `GET /health` → Healthcheck con uptime, version, sitios monitoreados
  - `GET /api/export` → Descarga CSV con historial de chequeos

### 3.8. Estado compartido (state.go)

- `DashboardState` con `sync.RWMutex` para concurrencia
- `Update()` — Escribe nuevo estado (despues de cada ciclo de checks)
- `Get()` — Lee estado actual (desde el handler HTTP)
- Thread-safe: el loop de checks escribe, el servidor HTTP lee

### 3.9. Paquete `internal/notifier` (telegram.go)

- **`SendStateChangeAlert(botToken, chatID, siteName, oldStatus, newStatus, details string) error`**:
  - Construye mensaje en formato Markdown con emojis, nombre del sitio, estado anterior → nuevo, detalles y hora
  - Llama a `sendTelegramMessage()` que hace POST a `https://api.telegram.org/bot{token}/sendMessage`
  - Manejo de errores: loggea el error pero nunca detiene el monitor
- **Ejecución no bloqueante:** Se llama con `go notifier.SendStateChangeAlert()` desde `runChecks()`
- **Configuración:** Sección `telegram` en `config.yaml` (`enabled`, `bot_token`, `chat_id`)
- **Detección de cambios:** `prevStates map[string]bool` en `main.go` — compara estado actual vs anterior por cada sitio

---

## 4. ESTADO ACTUAL (v1.3)

### 4.1. Funcionalidades implementadas

- [x] Verificacion HTTP GET con status code y latencia
- [x] UI terminal con Lip Gloss (cajas, colores, barras de progreso)
- [x] Multiples URLs (8 sitios Vercel)
- [x] Loop continuo con time.Ticker e intervalo configurable
- [x] Graceful shutdown con SIGINT/SIGTERM
- [x] Flags -config, -interval, -port
- [x] Dashboard web con CSS Grid + ApexCharts (modo oscuro)
- [x] Endpoint /api/status (JSON)
- [x] Embed de archivos estaticos en binario Go
- [x] SQLite storage con build tag opcional
- [x] Estado compartido thread-safe para dashboard
- [x] Auto-refresh del dashboard cada 10 segundos
- [x] Limpieza de pantalla automatica entre ciclos
- [x] **Alertas por Telegram** (opcional, configurable via YAML, goroutine no bloqueante)
- [x] **Notificacion de inicio por Telegram** (startup alert)
- [x] **SSL Cert Expiry Check** (captura expiracion, alerta visual < 30 dias)
- [x] **Healthcheck** (`GET /health` — uptime, sitios monitoreados, version)
- [x] **Export CSV** (`GET /api/export` — descarga historial de chequeos)
- [x] **Dashboard oscuro moderno** (sin Bootstrap, CSS Grid, paleta consistente con terminal)
- [x] **Tests unitarios** (checker, server, notifier con httptest)

### 4.2. Sitios monitoreados

| # | Nombre | URL | Plataforma |
|---|---|---|---|
| 1 | Mi Portfolio | marcelo-palma-portfolio.vercel.app | Vercel |
| 2 | Angular Music Player | music-player-roan-eight.vercel.app | Vercel |
| 3 | CodeMp AI | code-mp-ai.vercel.app | Vercel |
| 4 | DevNotes | dev-notes-ruby.vercel.app | Vercel |
| 5 | Markdown Converter | markdown-converter-six.vercel.app | Vercel |
| 6 | Task Manager Pro | taskmanager-pro-pi.vercel.app | Vercel |
| 7 | Web Vault | web-vault-tawny.vercel.app | Vercel |
| 8 | Study Apps | matematicas-t.vercel.app | Vercel |

### 4.3. Pendientes (Proximas features — v1.3+)

| # | Feature | Descripcion | Ubicacion sugerida | Prioridad |
|---|---|---|---|---|
| 1 | Pruebas unitarias | Tests con `httptest` para checker, storage y server. Coverage minimo 70%. | `*_test.go` en cada paquete | Alta |
| 2 | Exportar historial a CSV | Endpoint `GET /api/export` o flag `-export` para generar CSV con historial completo de checks. | `server.go` + `storage_entry.go` | Media |
| 3 | SSL cert expiry check | Verificar expiracion del certificado TLS de cada sitio. Alertar si quedan < 30 dias. | `internal/monitor/checker.go` | Media |
| 4 | Intervalo configurable desde dashboard | Boton/slider en el dashboard web para cambiar intervalo en tiempo real sin reiniciar. | `web/static/index.html` + `server.go` | Baja |
| 5 | Healthcheck del monitor | Endpoint `GET /health` que devuelve OK si el monitor esta vivo y corriendo. | `server.go` | Baja |

### 4.4. Comandos de verificacion

```bash
cd blackbox-monitor

# Compilar
go build -o bin/blackbox-monitor .

# Ejecutar con dashboard web
./bin/blackbox-monitor -port :8080

# Ejecutar solo terminal
./bin/blackbox-monitor -port ""

# Con SQLite
go build -tags sqlite -o bin/blackbox-monitor .

# Verificar codigo
go vet ./...

# Abrir dashboard
# http://localhost:8080
```

---

## 5. CONFIGURACION (config.example.yaml)

```yaml
interval: 60

# ===== ALERTAS POR TELEGRAM (OPCIONAL) =====
telegram:
  enabled: false
  bot_token: "YOUR_BOT_TOKEN_HERE"
  chat_id: "YOUR_CHAT_ID_HERE"

sites:
  - name: "Mi Portfolio"
    url: "https://marcelo-palma-portfolio.vercel.app"
    timeout: 5000
  - name: "Angular Music Player"
    url: "https://music-player-roan-eight.vercel.app"
    timeout: 5000
```

> **Nota:** `config.yaml` contiene tus credenciales reales y esta ignorado por git.  
> El repositorio contiene `config.example.yaml` con placeholders.  
> Copia `cp config.example.yaml config.yaml` y edita con tus datos.

### Flags

| Flag | Default | Descripcion |
|---|---|---|
| `-config` | `config.yaml` | Ruta del archivo de configuracion |
| `-interval` | valor del config | Intervalo en segundos |
| `-port` | `:8080` | Puerto del dashboard (vacio = sin dashboard) |

---

## 6. DEPENDENCIAS

### Directas
| Paquete | Version | Uso |
|---|---|---|
| `github.com/charmbracelet/lipgloss` | v1.1.0 | Estilos de terminal |
| `gopkg.in/yaml.v3` | v3.0.1 | Parseo YAML |

### Opcionales (build tag)
| Paquete | Version | Tag |
|---|---|---|
| `github.com/mattn/go-sqlite3` | latest | `sqlite` (requiere CGO) |

---

## 7. CONVENCIONES DE CODIGO

- **Idioma:** Codigo en ingles (variables, funciones, comentarios); strings de UI en espanol.
- **Estilo:** Sin comentarios innecesarios. Seguir estilo existente.
- **Errores:** `fmt.Errorf("mensaje: %w", err)` con wrapping.
- **Nuevos paquetes:** Siempre en `internal/`.
- **Dependencias:** Preferir stdlib. Si se agrega una, justificar.
- **Build tags:** Archivos con `//go:build sqlite` y `//go:build !sqlite` para alternativas.

---

## 8. HISTORIAL DE CAMBIOS

### v1.3 (Junio 2026) — SSL, Healthcheck, Export, Diseño oscuro
- SSL cert expiry check: nuevo campo CertExpiry en SiteResult, extracción TLS en CheckSite
- Aviso visual en terminal (colores) y dashboard (columna SSL) cuando el certificado expira en < 30 días
- Healthcheck endpoint `GET /health`: uptime, version, sitios monitoreados
- Export CSV endpoint `GET /api/export`: descarga historial en formato CSV
- `GetRecentLogs()` agregado a la interfaz Store y ambas implementaciones
- Dashboard web rediseñado: modo oscuro, CSS Grid, tipografía Inter, sin Bootstrap
- Paleta de dashboard unificada con los colores de la terminal (Lip Gloss)
- Notificación de inicio por Telegram (SendStartupNotification)
- Startup alert enviada en goroutine al arrancar con Telegram habilitado
- Tests existentes actualizados y pasando (go test -race)
- README.md y SYSTEM.md actualizados

### v1.2 (Junio 2026) — Alertas Telegram
- Nuevo paquete `internal/notifier/telegram.go` con envio no bloqueante via Bot API
- Configuracion opcional en `config.yaml` (enabled, bot_token, chat_id)
- Deteccion de cambios de estado con mapa `prevStates` en `main.go`
- Alertas solo en cambio real (Online↔Offline), sin repeticiones
- Goroutine para no ralentizar el monitoreo
- Manejo de errores sin detener el programa
- `config.example.yaml` para compartir sin exponer credenciales
- `config.yaml` agregado a `.gitignore`

### v1.1 (Junio 2026)
- Loop de monitoreo continuo con time.Ticker
- Graceful shutdown con SIGINT/SIGTERM
- Flags -config, -interval, -port
- Dashboard web con Bootstrap 3 + ApexCharts (template Super Admin)
- Endpoint /api/status con JSON
- Embed de archivos estaticos en binario Go
- Estado compartido thread-safe (DashboardState)
- UI terminal rediseñada con cajas por sitio
- 8 sitios Vercel monitoreados
- Storage SQLite con build tag (conectado al main loop)

### v1.0 (MVP — Junio 2026)
- Verificacion HTTP GET con status code
- UI terminal con Lip Gloss
- Configuracion via YAML
- Medicion de latencia
- Manejo de errores de red

---

*Documento actualizado el Junio 2026 — BlackBox Monitor v1.3*

---

## 9. ROADMAP DETALLADO (PROXIMAS FEATURES)

### 9.1. Intervalo configurable desde dashboard (v1.4)

- **Endpoint:** `POST /api/interval` con body `{"interval": 30}`
- **Frontend:** Slider o input en el dashboard web
- **Backend:** Actualizar ticker dinamicamente con `ticker.Reset()`

### 9.2. Uptime history chart (v1.4)

- **Ubicacion:** Dashboard web, nuevo gráfico de área
- **Datos:** Consultar logs históricos de SQLite para calcular uptime por sitio
- **Endpoint:** `GET /api/history?site=nombre&days=7`

### 9.3. Soporte para múltiples canales de notificación (v1.5)

- Webhooks genéricos, email (SMTP), Slack
- Interfaz `Notifier` en `internal/notifier/`
