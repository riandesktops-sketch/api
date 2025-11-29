# Zodiac AI Chat & Social - Deployment Guide (Render.com)

## 🚀 Deploy ke Render.com (100% GRATIS)

### Persiapan

#### 1. Setup MongoDB Atlas (Database Gratis)
```bash
# 1. Buka https://mongodb.com/cloud/atlas
# 2. Sign up dengan Google/GitHub (TIDAK PERLU KARTU KREDIT)
# 3. Create Free Cluster:
#    - Provider: AWS
#    - Region: Singapore (terdekat)
#    - Cluster Tier: M0 Sandbox (FREE)
# 4. Create Database User:
#    - Username: zodiac_user
#    - Password: (generate strong password)
# 5. Network Access:
#    - Add IP: 0.0.0.0/0 (allow from anywhere)
# 6. Get Connection String:
#    - Click "Connect" → "Connect your application"
#    - Copy connection string
#    - Format: mongodb+srv://zodiac_user:<password>@cluster.mongodb.net/zodiac_ai
```

#### 2. Push ke GitHub
```bash
cd zodiac-ai-backend

# Initialize git (jika belum)
git init
git add .
git commit -m "Initial commit - Zodiac AI Backend"

# Create repo di GitHub, lalu:
git remote add origin https://github.com/YOUR_USERNAME/zodiac-ai-backend.git
git branch -M main
git push -u origin main
```

---

### Deploy ke Render

#### 1. Sign Up Render.com
```bash
# 1. Buka https://render.com
# 2. Sign up dengan GitHub (GRATIS, TIDAK PERLU KARTU KREDIT)
# 3. Authorize Render untuk akses GitHub repos
```

#### 2. Deploy Menggunakan render.yaml (Otomatis)
```bash
# 1. Di Render Dashboard, klik "New" → "Blueprint"
# 2. Connect repository: zodiac-ai-backend
# 3. Render akan otomatis detect render.yaml
# 4. Klik "Apply"
# 5. Render akan deploy semua 5 services sekaligus!
```

#### 3. Set Environment Variables
Setelah services dibuat, set environment variables untuk setiap service:

**Semua Services (Auth, Chat, Social):**
- `MONGODB_URI`: `mongodb+srv://zodiac_user:<password>@cluster.mongodb.net/zodiac_ai`
- `JWT_SECRET`: (generate random string, misal: `your-super-secret-jwt-key-2024`)

**AI Service:**
- `GEMINI_API_KEY`: (dari https://makersuite.google.com/app/apikey)

**API Gateway:**
- `MONGODB_URI`: (sama seperti di atas)
- `JWT_SECRET`: (sama seperti di atas)
- Service URLs akan otomatis terisi oleh Render

---

### Deploy Manual (Alternatif)

Jika tidak mau pakai render.yaml, deploy satu per satu:

#### 1. Deploy Auth Service
```bash
# Di Render Dashboard:
# 1. New → Web Service
# 2. Connect repository: zodiac-ai-backend
# 3. Name: zodiac-auth-service
# 4. Environment: Go
# 5. Build Command: go build -o bin/auth ./services/auth-service
# 6. Start Command: ./bin/auth
# 7. Add Environment Variables:
#    - MONGODB_URI
#    - MONGODB_DATABASE=zodiac_ai
#    - JWT_SECRET
#    - AUTH_SERVICE_PORT=8001
# 8. Create Web Service
```

#### 2. Deploy AI Service
```bash
# Ulangi langkah di atas dengan:
# - Name: zodiac-ai-service
# - Build Command: go build -o bin/ai ./services/ai-service
# - Start Command: ./bin/ai
# - Environment Variables:
#   - GEMINI_API_KEY
#   - AI_SERVICE_PORT=8004
```

#### 3. Deploy Chat Service
```bash
# - Name: zodiac-chat-service
# - Build Command: go build -o bin/chat ./services/chat-service
# - Start Command: ./bin/chat
# - Environment Variables:
#   - MONGODB_URI
#   - MONGODB_DATABASE=zodiac_ai
#   - AI_SERVICE_URL=https://zodiac-ai-service.onrender.com
#   - CHAT_SERVICE_PORT=8002
```

#### 4. Deploy Social Service
```bash
# - Name: zodiac-social-service
# - Build Command: go build -o bin/social ./services/social-service
# - Start Command: ./bin/social
# - Environment Variables:
#   - MONGODB_URI
#   - MONGODB_DATABASE=zodiac_ai
#   - SOCIAL_SERVICE_PORT=8003
```

#### 5. Deploy API Gateway
```bash
# - Name: zodiac-api-gateway
# - Build Command: go build -o bin/gateway ./api-gateway
# - Start Command: ./bin/gateway
# - Environment Variables:
#   - MONGODB_URI
#   - JWT_SECRET
#   - AUTH_SERVICE_URL=https://zodiac-auth-service.onrender.com
#   - CHAT_SERVICE_URL=https://zodiac-chat-service.onrender.com
#   - SOCIAL_SERVICE_URL=https://zodiac-social-service.onrender.com
#   - AI_SERVICE_URL=https://zodiac-ai-service.onrender.com
#   - API_GATEWAY_PORT=8000
```

---

### Run Migrations

Setelah semua services deploy, jalankan migrations:

```bash
# 1. Clone repo di local
# 2. Update .env dengan MongoDB Atlas URI
# 3. Run migration:
make migrate

# Atau manual:
go run scripts/migrate.go
```

---

### Testing

#### 1. Get API Gateway URL
```bash
# Di Render Dashboard, klik API Gateway service
# Copy URL: https://zodiac-api-gateway.onrender.com
```

#### 2. Test Register
```bash
curl -X POST https://zodiac-api-gateway.onrender.com/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User",
    "date_of_birth": "1995-03-21T00:00:00Z",
    "gender": "male"
  }'
```

#### 3. Test Login
```bash
curl -X POST https://zodiac-api-gateway.onrender.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

---

## ⚠️ Catatan Penting

### Free Tier Limitations
- **Sleep after 15 minutes idle**: Services akan sleep setelah 15 menit tidak ada request
- **Wake-up time**: ~30 detik untuk wake up dari sleep
- **750 hours/month per service**: Cukup untuk development/demo

### Tips Optimasi
1. **Gunakan cron job** untuk keep services awake:
   ```bash
   # Setup cron di cron-job.org
   # Ping setiap 10 menit:
   https://zodiac-api-gateway.onrender.com/health
   ```

2. **Gabung services** jika ingin mengurangi sleep time:
   - Gabung Auth + Chat jadi 1 service
   - Gabung Social + AI jadi 1 service
   - Total: 3 services (Gateway + 2 combined)

---

## 🎉 Selesai!

Backend Anda sekarang live di:
- **API Gateway**: https://zodiac-api-gateway.onrender.com
- **Auth Service**: https://zodiac-auth-service.onrender.com
- **Chat Service**: https://zodiac-chat-service.onrender.com
- **Social Service**: https://zodiac-social-service.onrender.com
- **AI Service**: https://zodiac-ai-service.onrender.com

**Dokumentasi API**: Lihat `README.md` untuk semua endpoints

**Monitoring**: 
- Render Dashboard → Logs untuk setiap service
- MongoDB Atlas → Metrics untuk database usage

---

## 🔧 Troubleshooting

### Service tidak start?
```bash
# Check logs di Render Dashboard
# Common issues:
# 1. Missing environment variables
# 2. MongoDB connection failed (check URI)
# 3. Port already in use (pastikan PORT env var benar)
```

### MongoDB connection error?
```bash
# 1. Check MongoDB Atlas Network Access (0.0.0.0/0)
# 2. Check connection string format
# 3. Check database user credentials
```

### AI Service error?
```bash
# 1. Check GEMINI_API_KEY valid
# 2. Test API key di https://makersuite.google.com
```

---

## 💰 Biaya

**Total: $0/bulan** 🎉

- Render: Gratis (dengan sleep)
- MongoDB Atlas: Gratis (512MB)
- Gemini AI: Gratis (dengan rate limit)

**Upgrade Options** (jika perlu):
- Render Pro: $7/month per service (no sleep)
- MongoDB Atlas: $9/month (2GB storage)
