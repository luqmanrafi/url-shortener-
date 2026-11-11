![test](https://github.com/luqmanrafi/url-shortener-/actions/workflows/test.yml/badge.svg)

# URL Shortener

REST API pemendek URL yang ditulis dengan Go (standard library), SQLite, dan Docker.
Dibuat sebagai proyek belajar backend dengan arsitektur berlapis dan test otomatis.

## Fitur

- Memendekkan URL dengan kode acak (base62, 7 karakter)
- Custom alias dengan validasi (3-30 karakter, kata terlarang diblokir)
- Link kedaluwarsa (`410 Gone`)
- Statistik klik per link
- Endpoint `/healthz` (liveness) dan `/readyz` (readiness)
- Rate limiting per IP (token bucket) dan logging request
- Graceful shutdown
- Unit test dengan fake repository, CI lewat GitHub Actions

## Menjalankan

### Lokal

```bash
go run ./cmd/server
```

### Docker

```bash
docker build -t urlshortener .
docker run --rm -p 3000:3000 -v urldata:/data urlshortener
```

Volume `urldata` menyimpan database SQLite agar data tidak hilang saat container dihapus.

## Konfigurasi

| Variabel | Default | Keterangan |
|---|---|---|
| `PORT` | `3000` | Port server |
| `BASE_URL` | `http://localhost:PORT` | Dipakai untuk membentuk short URL |
| `DB_PATH` | `urls.db` | Lokasi file SQLite |

## Endpoint

| Method | Path | Keterangan |
|---|---|---|
| POST | `/api/shorten` | Membuat short URL |
| GET | `/{code}` | Redirect (302) ke URL asli |
| GET | `/api/stats/{code}` | Statistik klik |
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness (cek database) |

### Contoh

```bash
# kode acak
curl -X POST localhost:3000/api/shorten -d '{"url":"https://go.dev"}'
# custom alias + kedaluwarsa 1 jam
curl -X POST localhost:3000/api/shorten \
  -d '{"url":"https://go.dev","alias":"belajar-go","expires_in":3600}'
```

Respons:

```json
{"code":"belajar-go","short_url":"http://localhost:3000/belajar-go"}
```

| Status | Arti |
|---|---|
| 201 | Berhasil dibuat |
| 400 | URL, alias, atau `expires_in` tidak valid |
| 409 | Alias sudah dipakai |
| 410 | Link sudah kedaluwarsa |
| 429 | Terlalu banyak request |

## Arsitektur

```
request → logging → rate limit → handler → service → repository → SQLite
```

```
cmd/server/          entry point
internal/handler/    HTTP: parse request, petakan error ke status
internal/service/    logika bisnis dan validasi
internal/repository/ akses database (SQLite)
internal/model/      struct dan error bersama
internal/middleware/ logging dan rate limiter
```

`service` bergantung pada interface `Repository`, jadi database bisa ditukar dan mudah dites.

## Test

```bash
go test ./...
```