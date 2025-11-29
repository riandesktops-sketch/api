# Zodiac AI Chat & Social - Deploy ke Koyeb

## 🚀 Deploy ke Koyeb (100% GRATIS, TIDAK SLEEP)

### Keunggulan Koyeb
- ✅ **GRATIS** tanpa kartu kredit
- ✅ **TIDAK SLEEP** (always on)
- ✅ 512MB RAM gratis
- ✅ Auto-deploy dari GitHub
- ✅ Custom domain gratis
- ✅ SSL otomatis

---

## 📋 Persiapan

### 1. Setup MongoDB Atlas (Gratis)
```bash
# 1. Buka https://mongodb.com/cloud/atlas
# 2. Sign up dengan Google/GitHub (TIDAK PERLU KARTU KREDIT)
# 3. Create Free Cluster:
#    - Provider: AWS
#    - Region: Singapore
#    - Cluster Tier: M0 Sandbox (FREE)
# 4. Create Database User:
#    - Username: zodiac_user
#    - Password: (generate strong password, simpan!)
# 5. Network Access:
#    - Add IP: 0.0.0.0/0 (allow from anywhere)
# 6. Get Connection String:
#    - Click "Connect" → "Connect your application"
#    - Copy: mongodb+srv://zodiac_user:<password>@cluster.mongodb.net/zodiac_ai
```

### 2. Get Gemini API Key (Gratis)
```bash
# 1. Buka https://makersuite.google.com/app/apikey
# 2. Login dengan Google
# 3. Create API Key
# 4. Copy API key (simpan!)
```

### 3. Push ke GitHub
```bash
cd zodiac-ai-backend

# Initialize git (jika belum)
git init
git add .
git commit -m "All-in-one version for Koyeb"

# Create repo di GitHub, lalu:
git remote add origin https://github.com/YOUR_USERNAME/zodiac-ai-backend.git
git branch -M main
git push -u origin main
```

---

## 🚀 Deploy ke Koyeb

### Step 1: Sign Up Koyeb
```bash
# 1. Buka https://koyeb.com
# 2. Sign up dengan GitHub
# 3. TIDAK PERLU KARTU KREDIT!
# 4. Authorize Koyeb untuk akses GitHub repos
```

### Step 2: Create Service
```bash
# Di Koyeb Dashboard:

1. Click "Create Web Service"

2. Deployment Method:
   - Select: "GitHub"
   - Repository: zodiac-ai-backend
   - Branch: main

3. Builder:
   - Select: "Dockerfile" (auto-detect)
   - Atau pilih "Buildpack" → Go

4. Build Settings:
   - Build command: go build -o app ./cmd/all-in-one
   - Run command: ./app

5. Instance:
   - Type: Free (Nano)
   - Region: Frankfurt atau Singapore (terdekat)

6. Environment Variables:
   Click "Add Variable" untuk setiap:
   
   MONGODB_URI=mongodb+srv://zodiac_user:PASSWORD@cluster.mongodb.net/zodiac_ai
   MONGODB_DATABASE=zodiac_ai
   JWT_SECRET=your-super-secret-jwt-key-2024-change-this
   GEMINI_API_KEY=your-gemini-api-key-here
   PORT=8080

7. Service Name:
   - Name: zodiac-ai-backend

8. Click "Deploy"
```

### Step 3: Wait for Deployment
```bash
# Koyeb akan:
# 1. Clone repo dari GitHub
# 2. Build aplikasi Go
# 3. Deploy ke server
# 4. Assign public URL

# Proses: ~3-5 menit
```

---

## 🧪 Testing

### 1. Get Public URL
```bash
# Di Koyeb Dashboard, copy URL:
# https://zodiac-ai-backend-YOUR-APP.koyeb.app
```

### 2. Test Health Check
```bash
curl https://zodiac-ai-backend-YOUR-APP.koyeb.app/health

# Response:
# {
#   "status": "healthy",
#   "service": "zodiac-ai-all-in-one"
# }
```

### 3. Test Register
```bash
curl -X POST https://zodiac-ai-backend-YOUR-APP.koyeb.app/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User",
    "date_of_birth": "1995-03-21T00:00:00Z",
    "gender": "male"
  }'

# Response: User dengan zodiac_sign = "Aries"
```

### 4. Test Login
```bash
curl -X POST https://zodiac-ai-backend-YOUR-APP.koyeb.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Response: access_token dan refresh_token
```

### 5. Test AI Chat
```bash
# Gunakan access_token dari login
curl -X POST https://zodiac-ai-backend-YOUR-APP.koyeb.app/api/v1/chat/sessions \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "My First Chat"}'

# Lalu send message:
curl -X POST https://zodiac-ai-backend-YOUR-APP.koyeb.app/api/v1/chat/sessions/SESSION_ID/messages \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message": "Saya merasa sedih hari ini"}'

# Response: AI response dalam Bahasa Indonesia!
```

---

## 🔄 Run Migrations

Setelah deploy, jalankan migrations untuk create indexes:

```bash
# Di local machine:

# 1. Update .env dengan MongoDB Atlas URI
MONGODB_URI=mongodb+srv://zodiac_user:PASSWORD@cluster.mongodb.net/zodiac_ai

# 2. Run migration
make migrate

# Atau manual:
go run scripts/migrate.go
```

---

## 🎯 Semua Endpoints

Base URL: `https://zodiac-ai-backend-YOUR-APP.koyeb.app`

### Auth
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
GET    /api/v1/users/me
PUT    /api/v1/users/me
```

### Friends
```
POST   /api/v1/friends/requests
PUT    /api/v1/friends/requests/:id
GET    /api/v1/friends
GET    /api/v1/friends/status/:user_id
```

### Chat
```
POST   /api/v1/chat/sessions
GET    /api/v1/chat/sessions
POST   /api/v1/chat/sessions/:id/messages
GET    /api/v1/chat/sessions/:id/messages
POST   /api/v1/chat/sessions/:id/generate-insight
```

### Rooms
```
POST   /api/v1/rooms
GET    /api/v1/rooms
WS     /api/v1/rooms/:id/ws
```

### Social
```
POST   /api/v1/posts
GET    /api/v1/posts
GET    /api/v1/posts/:id
POST   /api/v1/posts/:id/like
DELETE /api/v1/posts/:id/like
POST   /api/v1/posts/:id/comments
GET    /api/v1/posts/:id/comments
```

---

## 🔧 Troubleshooting

### Build Failed?
```bash
# Check logs di Koyeb Dashboard
# Common issues:
# 1. Missing dependencies → run: go mod tidy
# 2. Import path error → check module name di go.mod
# 3. Build command salah → pastikan: go build -o app ./cmd/all-in-one
```

### MongoDB Connection Error?
```bash
# 1. Check MongoDB Atlas Network Access (0.0.0.0/0)
# 2. Check connection string format
# 3. Check password (no special characters yang perlu escape)
# 4. Test connection di local dulu
```

### AI Service Error?
```bash
# 1. Check GEMINI_API_KEY valid
# 2. Test API key di https://makersuite.google.com
# 3. Check quota (free tier: 60 requests/minute)
```

### Service Slow?
```bash
# Free tier Koyeb:
# - 512MB RAM
# - Shared CPU
# - Bisa lambat saat traffic tinggi
# 
# Upgrade ke Eco ($5/month) untuk:
# - 1GB RAM
# - Dedicated CPU
```

---

## 📊 Monitoring

### Koyeb Dashboard
```
- Logs: Real-time logs
- Metrics: CPU, Memory, Network usage
- Deployments: History & rollback
```

### MongoDB Atlas
```
- Metrics: Database operations
- Storage: Current usage (max 512MB free)
- Alerts: Email notifications
```

---

## 🎉 Selesai!

Backend Anda sekarang **LIVE 24/7** di Koyeb!

**URL:** `https://zodiac-ai-backend-YOUR-APP.koyeb.app`

**Keunggulan:**
- ✅ Gratis selamanya
- ✅ Tidak sleep (always on)
- ✅ Auto-deploy dari GitHub
- ✅ SSL otomatis
- ✅ Custom domain (jika mau)

**Monitoring:**
- Koyeb Dashboard: Logs & metrics
- MongoDB Atlas: Database usage
- UptimeRobot (optional): Uptime monitoring

---

## 💰 Biaya

**Total: $0/bulan** 🎉

- Koyeb: Gratis (512MB RAM)
- MongoDB Atlas: Gratis (512MB storage)
- Gemini AI: Gratis (60 req/min)

**Upgrade Options** (jika perlu):
- Koyeb Eco: $5/month (1GB RAM, dedicated CPU)
- MongoDB Atlas: $9/month (2GB storage)
- Gemini AI Pro: Pay per use

---

## 🔄 Auto-Deploy

Setiap kali Anda push ke GitHub:
```bash
git add .
git commit -m "Update feature"
git push origin main

# Koyeb akan otomatis:
# 1. Detect changes
# 2. Rebuild app
# 3. Deploy new version
# 4. Zero downtime!
```

**Perfect untuk development! 🚀**
