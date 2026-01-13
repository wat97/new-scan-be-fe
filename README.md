# Scandata - Warehouse Grade Checker

Aplikasi untuk verifikasi grade unit warehouse dengan QR code scanner.

## 🚀 Quick Start (Development)

```bash
# Start semua services
docker compose up --build

# Akses di http://localhost
```

## 🌐 Production Deployment (dengan SSL)

### 1. Setup Domain

Pastikan domain Anda sudah pointing ke server:
```bash
# A Record
scandata.yourdomain.com -> YOUR_SERVER_IP
```

### 2. Configure Environment

```bash
# Copy template
cp .env.production.example .env.production

# Edit dengan domain Anda
nano .env.production
```

Isi `.env.production`:
```env
DOMAIN=scandata.yourdomain.com
MYSQL_ROOT_PASSWORD=your-secure-password
MYSQL_PASSWORD=your-db-password
JWT_SECRET=your-very-long-secret-key
```

### 3. Update Caddyfile

Edit `Caddyfile` dan ganti `{$DOMAIN}` dengan domain Anda:
```
scandata.yourdomain.com {
    # ... config
}
```

### 4. Deploy

```bash
# Build dan start dengan production config
docker compose -f docker-compose.prod.yml --env-file .env.production up -d --build

# Check logs
docker compose -f docker-compose.prod.yml logs -f
```

### 5. SSL Certificate

Caddy akan **otomatis**:
- ✅ Generate SSL certificate dari Let's Encrypt
- ✅ Redirect HTTP ke HTTPS
- ✅ Auto-renew certificate

## 📁 Project Structure

```
scandata/
├── backend/                 # Go API Server
│   ├── Dockerfile
│   ├── main.go
│   ├── handlers/
│   ├── models/
│   └── ...
├── frontend/                # Static Frontend
│   ├── index.html
│   ├── css/
│   └── js/
├── docker-compose.yml       # Development (nginx)
├── docker-compose.prod.yml  # Production (Caddy + SSL)
├── Caddyfile               # Caddy configuration
├── .env                    # Development env
└── .env.production.example # Production env template
```

## 🔧 Useful Commands

```bash
# Development
docker compose up --build              # Start dev
docker compose down                    # Stop
docker compose logs -f backend         # View backend logs

# Production
docker compose -f docker-compose.prod.yml up -d --build
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml logs -f

# Database
docker compose exec db mysql -u scandata -p scandata  # Access MySQL
```

## 👤 Default Login

- **Username:** admin
- **Password:** admin123

> ⚠️ Ganti password default setelah login pertama kali!

## 📱 Features

- ✅ QR Code Scanner
- ✅ Unit Management (CRUD)
- ✅ User Management
- ✅ Scan History
- ✅ Reports & Analytics
- ✅ Excel Export
- ✅ Mobile Responsive
- ✅ Dark Theme
