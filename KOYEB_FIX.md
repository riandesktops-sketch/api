# Zodiac AI Backend - Koyeb Deployment Fix

## ❌ Error: "no command to run your application"

### Solusi:

#### **Option 1: Set Run Command di Koyeb Dashboard**

1. Go to your service in Koyeb
2. Click "Settings"
3. Scroll to "Run command"
4. Set: `./app`
5. Click "Save"
6. Redeploy

#### **Option 2: Use Procfile (Recommended)**

Saya sudah buatkan file `Procfile` di root project. Push ke GitHub:

```bash
git add Procfile
git commit -m "Add Procfile for Koyeb"
git push origin main
```

Koyeb akan auto-detect Procfile dan tahu cara run app.

---

## 🚀 Complete Koyeb Setup (Step by Step)

### 1. **Builder Settings**
- **Builder**: Buildpack
- **Language**: Go
- **Build command**: `go build -o app ./cmd/all-in-one`
- **Run command**: `./app`  ← **PENTING!**

### 2. **Instance Settings**
- **Type**: Free (Nano)
- **Region**: Frankfurt atau Singapore
- **Port**: 8080

### 3. **Environment Variables** (WAJIB!)
```
MONGODB_URI=mongodb+srv://user:password@cluster.mongodb.net/zodiac_ai
MONGODB_DATABASE=zodiac_ai
JWT_SECRET=your-super-secret-key-min-32-karakter
GEMINI_API_KEY=your-gemini-api-key
PORT=8080
```

### 4. **Deploy**
- Click "Deploy"
- Wait 2-3 minutes
- Check logs

---

## 🔍 Troubleshooting Deployment Errors

### Error: "Application exited with code 1"
**Cause**: Missing environment variables atau MongoDB connection failed

**Solution**:
1. Check semua env vars sudah diset
2. Test MongoDB connection string di local dulu
3. Pastikan MongoDB Atlas allow 0.0.0.0/0

### Error: "Port already in use"
**Cause**: PORT env var tidak diset

**Solution**: Set `PORT=8080` di environment variables

### Error: "Cannot connect to MongoDB"
**Cause**: MongoDB Atlas network access atau connection string salah

**Solution**:
1. MongoDB Atlas → Network Access → Add IP: 0.0.0.0/0
2. Check connection string format:
   ```
   mongodb+srv://username:password@cluster.mongodb.net/zodiac_ai
   ```
3. Pastikan password tidak ada special characters yang perlu di-encode

---

## ✅ Verification Steps

Setelah deploy sukses:

### 1. Check Health Endpoint
```bash
curl https://your-app.koyeb.app/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "zodiac-ai-all-in-one"
}
```

### 2. Check Logs
Di Koyeb dashboard → Logs, pastikan tidak ada error

### 3. Test API
```bash
# Register user
curl -X POST https://your-app.koyeb.app/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User",
    "date_of_birth": "1995-03-21T00:00:00Z",
    "gender": "male"
  }'
```

---

## 📝 Quick Checklist

Before deploying:
- [ ] MongoDB Atlas cluster created & accessible
- [ ] Gemini API key obtained
- [ ] All code pushed to GitHub
- [ ] `Procfile` added (or Run command set)
- [ ] Environment variables configured
- [ ] Build command: `go build -o app ./cmd/all-in-one`
- [ ] Run command: `./app`
- [ ] Port: 8080

---

## 🆘 Still Not Working?

Share the **full error logs** from Koyeb dashboard (click "Download" on logs).

Common issues:
1. Missing `GEMINI_API_KEY` → App crashes on startup
2. Wrong `MONGODB_URI` → Cannot connect to database
3. Missing `JWT_SECRET` → Auth endpoints fail
4. Wrong `PORT` → Koyeb cannot route traffic

**Pro tip**: Test locally first dengan:
```bash
make dev-all-in-one
```

Jika jalan di local, pasti bisa di Koyeb!
