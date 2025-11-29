# Zodiac AI Backend - API Documentation

> **Dokumentasi Lengkap API untuk Frontend Integration**
> 
> Base URL: `https://your-domain.com/api/v1` atau `http://localhost:8080/api/v1`

## 📋 Table of Contents

- [Response Format](#response-format)
- [Authentication](#authentication)
- [Auth Service](#auth-service)
- [Friend Service](#friend-service)
- [Chat Service](#chat-service)
- [Room Service](#room-service)
- [Social Service](#social-service)
- [AI Service](#ai-service)
- [Error Handling](#error-handling)
- [Common Issues & Troubleshooting](#common-issues--troubleshooting)

---

## Response Format

Semua endpoint menggunakan format response yang konsisten:

### Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    // Response data here
  },
  "meta": {
    // Optional metadata for pagination
    "next_cursor": "cursor_string",
    "has_more": true,
    "limit": 20
  }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error message",
  "error": {
    "code": "ERROR_CODE",
    "message": "Detailed error message",
    "details": {
      // Optional additional error details
    }
  }
}
```

### HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `429` - Too Many Requests
- `500` - Internal Server Error
- `503` - Service Unavailable

---

## Authentication

### Token Types
API menggunakan **JWT (JSON Web Token)** dengan dua jenis token:
- **Access Token**: Untuk autentikasi request (expired dalam 15 menit)
- **Refresh Token**: Untuk mendapatkan access token baru (expired dalam 30 hari)

### Cara Menggunakan Token

**Header untuk Protected Endpoints:**
```
Authorization: Bearer <access_token>
```

**Contoh di JavaScript:**
```javascript
const response = await fetch('http://localhost:8080/api/v1/users/me', {
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'Content-Type': 'application/json'
  }
});
```

---

## Auth Service

### 1. Register User

**Endpoint:** `POST /api/v1/auth/register`

**Authentication:** ❌ Not Required

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "full_name": "John Doe",
  "date_of_birth": "1995-03-15T00:00:00Z",
  "gender": "male"
}
```

**Field Validations:**
- `email`: Required, valid email format
- `password`: Required, minimum 8 characters
- `full_name`: Required
- `date_of_birth`: Required, ISO 8601 format
- `gender`: Required, one of: `male`, `female`, `other`

**Success Response (201):**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "email": "user@example.com",
      "full_name": "John Doe",
      "display_name": "",
      "date_of_birth": "1995-03-15T00:00:00Z",
      "gender": "male",
      "zodiac_sign": "Pisces",
      "bio": "",
      "avatar_url": "",
      "total_posts": 0,
      "friends_count": 0,
      "created_at": "2025-11-29T10:00:00Z",
      "updated_at": "2025-11-29T10:00:00Z"
    }
  }
}
```

**Error Response (409):**
```json
{
  "success": false,
  "message": "Email already exists",
  "error": {
    "code": "CONFLICT",
    "message": "Email already exists"
  }
}
```

**Frontend Example:**
```javascript
async function register(userData) {
  try {
    const response = await fetch('http://localhost:8080/api/v1/auth/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: userData.email,
        password: userData.password,
        full_name: userData.fullName,
        date_of_birth: new Date(userData.dateOfBirth).toISOString(),
        gender: userData.gender
      })
    });
    
    const result = await response.json();
    
    if (result.success) {
      // Save tokens
      localStorage.setItem('access_token', result.data.access_token);
      localStorage.setItem('refresh_token', result.data.refresh_token);
      localStorage.setItem('user', JSON.stringify(result.data.user));
      return result.data;
    } else {
      throw new Error(result.message);
    }
  } catch (error) {
    console.error('Registration error:', error);
    throw error;
  }
}
```

---

### 2. Login

**Endpoint:** `POST /api/v1/auth/login`

**Authentication:** ❌ Not Required

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "email": "user@example.com",
      "full_name": "John Doe",
      "display_name": "Johnny",
      "date_of_birth": "1995-03-15T00:00:00Z",
      "gender": "male",
      "zodiac_sign": "Pisces",
      "bio": "Love astrology!",
      "avatar_url": "https://example.com/avatar.jpg",
      "total_posts": 5,
      "friends_count": 10,
      "created_at": "2025-11-29T10:00:00Z",
      "updated_at": "2025-11-29T15:00:00Z"
    }
  }
}
```

**Error Response (401):**
```json
{
  "success": false,
  "message": "Invalid email or password",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid email or password"
  }
}
```

**Frontend Example:**
```javascript
async function login(email, password) {
  const response = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  
  const result = await response.json();
  
  if (result.success) {
    localStorage.setItem('access_token', result.data.access_token);
    localStorage.setItem('refresh_token', result.data.refresh_token);
    localStorage.setItem('user', JSON.stringify(result.data.user));
  }
  
  return result;
}
```

---

### 3. Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Authentication:** ❌ Not Required

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Error Response (401):**
```json
{
  "success": false,
  "message": "Invalid or expired refresh token",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired refresh token"
  }
}
```

**Frontend Example:**
```javascript
async function refreshAccessToken() {
  const refreshToken = localStorage.getItem('refresh_token');
  
  const response = await fetch('http://localhost:8080/api/v1/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken })
  });
  
  const result = await response.json();
  
  if (result.success) {
    localStorage.setItem('access_token', result.data.access_token);
    return result.data.access_token;
  } else {
    // Refresh token expired, redirect to login
    localStorage.clear();
    window.location.href = '/login';
  }
}
```

---

### 4. Get Profile

**Endpoint:** `GET /api/v1/users/me`

**Authentication:** ✅ Required

**Headers:**
```
Authorization: Bearer <access_token>
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "full_name": "John Doe",
    "display_name": "Johnny",
    "date_of_birth": "1995-03-15T00:00:00Z",
    "gender": "male",
    "zodiac_sign": "Pisces",
    "bio": "Love astrology!",
    "avatar_url": "https://example.com/avatar.jpg",
    "total_posts": 5,
    "friends_count": 10,
    "created_at": "2025-11-29T10:00:00Z",
    "updated_at": "2025-11-29T15:00:00Z"
  }
}
```

**Frontend Example:**
```javascript
async function getProfile() {
  const token = localStorage.getItem('access_token');
  
  const response = await fetch('http://localhost:8080/api/v1/users/me', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  const result = await response.json();
  return result.data;
}
```

---

### 5. Update Profile

**Endpoint:** `PUT /api/v1/users/me`

**Authentication:** ✅ Required

**Request Body:**
```json
{
  "display_name": "Johnny Star",
  "bio": "Astrology enthusiast and Pisces lover!",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

**Note:** Semua field optional. Hanya kirim field yang ingin diupdate.

**Success Response (200):**
```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "full_name": "John Doe",
    "display_name": "Johnny Star",
    "date_of_birth": "1995-03-15T00:00:00Z",
    "gender": "male",
    "zodiac_sign": "Pisces",
    "bio": "Astrology enthusiast and Pisces lover!",
    "avatar_url": "https://example.com/new-avatar.jpg",
    "total_posts": 5,
    "friends_count": 10,
    "created_at": "2025-11-29T10:00:00Z",
    "updated_at": "2025-11-29T16:00:00Z"
  }
}
```

**Frontend Example:**
```javascript
async function updateProfile(updates) {
  const token = localStorage.getItem('access_token');
  
  const response = await fetch('http://localhost:8080/api/v1/users/me', {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(updates)
  });
  
  const result = await response.json();
  
  if (result.success) {
    localStorage.setItem('user', JSON.stringify(result.data));
  }
  
  return result;
}
```

---

## Friend Service

### 1. Send Friend Request

**Endpoint:** `POST /api/v1/friends/requests`

**Authentication:** ✅ Required

**Request Body:**
```json
{
  "target_user_id": "507f1f77bcf86cd799439012"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Friend request sent successfully",
  "data": null
}
```

**Error Response (409):**
```json
{
  "success": false,
  "message": "Already friends",
  "error": {
    "code": "CONFLICT",
    "message": "Already friends"
  }
}
```

---

### 2. Accept/Reject Friend Request

**Endpoint:** `PUT /api/v1/friends/requests/:id`

**Authentication:** ✅ Required

**URL Parameters:**
- `id`: Friend request ID

**Request Body:**
```json
{
  "action": "accept"
}
```

**Valid Actions:** `accept` atau `reject`

**Success Response (200):**
```json
{
  "success": true,
  "message": "Friend request accepted",
  "data": null
}
```

**Error Response (404):**
```json
{
  "success": false,
  "message": "Friend request not found",
  "error": {
    "code": "NOT_FOUND",
    "message": "Friend request not found"
  }
}
```

---

### 3. Get Friends List

**Endpoint:** `GET /api/v1/friends`

**Authentication:** ✅ Required

**Success Response (200):**
```json
{
  "success": true,
  "message": "Friends retrieved successfully",
  "data": {
    "friend_ids": [
      "507f1f77bcf86cd799439012",
      "507f1f77bcf86cd799439013",
      "507f1f77bcf86cd799439014"
    ],
    "count": 3
  }
}
```

**Frontend Example:**
```javascript
async function getFriends() {
  const token = localStorage.getItem('access_token');
  
  const response = await fetch('http://localhost:8080/api/v1/friends', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  const result = await response.json();
  return result.data; // { friend_ids: [...], count: 3 }
}
```

---

### 4. Check Friendship Status

**Endpoint:** `GET /api/v1/friends/status/:user_id`

**Authentication:** ✅ Required

**URL Parameters:**
- `user_id`: Target user ID

**Success Response (200):**
```json
{
  "success": true,
  "message": "Friendship status retrieved",
  "data": {
    "status": "friends"
  }
}
```

**Possible Status Values:**
- `none`: Tidak ada relasi
- `pending`: Friend request pending
- `friends`: Sudah berteman

---

## Chat Service

### 1. Create Chat Session

**Endpoint:** `POST /api/v1/chat/sessions`

**Authentication:** ✅ Required

**Request Body:**
```json
{
  "title": "My First Chat"
}
```

**Note:** Field `title` optional. Jika tidak dikirim, akan auto-generate.

**Success Response (201):**
```json
{
  "success": true,
  "message": "Chat session created",
  "data": {
    "id": "507f1f77bcf86cd799439020",
    "user_id": "507f1f77bcf86cd799439011",
    "title": "My First Chat",
    "created_at": "2025-11-29T10:00:00Z",
    "updated_at": "2025-11-29T10:00:00Z"
  }
}
```

---

### 2. Get All Chat Sessions

**Endpoint:** `GET /api/v1/chat/sessions`

**Authentication:** ✅ Required

**Success Response (200):**
```json
{
  "success": true,
  "message": "Sessions retrieved successfully",
  "data": [
    {
      "id": "507f1f77bcf86cd799439020",
      "user_id": "507f1f77bcf86cd799439011",
      "title": "My First Chat",
      "created_at": "2025-11-29T10:00:00Z",
      "updated_at": "2025-11-29T10:00:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439021",
      "user_id": "507f1f77bcf86cd799439011",
      "title": "Career Advice",
      "created_at": "2025-11-29T11:00:00Z",
      "updated_at": "2025-11-29T11:30:00Z"
    }
  ]
}
```

---

### 3. Send Message

**Endpoint:** `POST /api/v1/chat/sessions/:id/messages`

**Authentication:** ✅ Required

**URL Parameters:**
- `id`: Session ID

**Request Body:**
```json
{
  "message": "What's my horoscope for today?"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "user_message": {
      "id": "507f1f77bcf86cd799439030",
      "session_id": "507f1f77bcf86cd799439020",
      "user_id": "507f1f77bcf86cd799439011",
      "sender": "USER",
      "content": "What's my horoscope for today?",
      "created_at": "2025-11-29T10:05:00Z"
    },
    "ai_message": {
      "id": "507f1f77bcf86cd799439031",
      "session_id": "507f1f77bcf86cd799439020",
      "user_id": "507f1f77bcf86cd799439011",
      "sender": "AI",
      "content": "As a Pisces, today is a great day for creativity...",
      "created_at": "2025-11-29T10:05:02Z"
    }
  }
}
```

**Error Response (404):**
```json
{
  "success": false,
  "message": "Chat session not found",
  "error": {
    "code": "NOT_FOUND",
    "message": "Chat session not found"
  }
}
```

**Error Response (503):**
```json
{
  "success": false,
  "message": "AI service temporarily unavailable",
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "AI service temporarily unavailable"
  }
}
```

**Frontend Example:**
```javascript
async function sendMessage(sessionId, message) {
  const token = localStorage.getItem('access_token');
  
  const response = await fetch(`http://localhost:8080/api/v1/chat/sessions/${sessionId}/messages`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ message })
  });
  
  const result = await response.json();
  return result.data; // { user_message: {...}, ai_message: {...} }
}
```

---

### 4. Get Messages (with Pagination)

**Endpoint:** `GET /api/v1/chat/sessions/:id/messages`

**Authentication:** ✅ Required

**URL Parameters:**
- `id`: Session ID

**Query Parameters:**
- `cursor` (optional): Cursor untuk pagination
- `limit` (optional): Jumlah messages per page (default: 20)

**Example Request:**
```
GET /api/v1/chat/sessions/507f1f77bcf86cd799439020/messages?limit=10&cursor=abc123
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Messages retrieved successfully",
  "data": [
    {
      "id": "507f1f77bcf86cd799439031",
      "session_id": "507f1f77bcf86cd799439020",
      "user_id": "507f1f77bcf86cd799439011",
      "sender": "AI",
      "content": "As a Pisces, today is a great day for creativity...",
      "created_at": "2025-11-29T10:05:02Z"
    },
    {
      "id": "507f1f77bcf86cd799439030",
      "session_id": "507f1f77bcf86cd799439020",
      "user_id": "507f1f77bcf86cd799439011",
      "sender": "USER",
      "content": "What's my horoscope for today?",
      "created_at": "2025-11-29T10:05:00Z"
    }
  ],
  "meta": {
    "next_cursor": "xyz789",
    "has_more": true,
    "limit": 10
  }
}
```

**Frontend Example:**
```javascript
async function getMessages(sessionId, cursor = '', limit = 20) {
  const token = localStorage.getItem('access_token');
  const params = new URLSearchParams({ limit: limit.toString() });
  if (cursor) params.append('cursor', cursor);
  
  const response = await fetch(
    `http://localhost:8080/api/v1/chat/sessions/${sessionId}/messages?${params}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );
  
  const result = await response.json();
  
  return {
    messages: result.data,
    nextCursor: result.meta?.next_cursor,
    hasMore: result.meta?.has_more
  };
}
```

---

### 5. Generate Insight

**Endpoint:** `POST /api/v1/chat/sessions/:id/generate-insight`

**Authentication:** ✅ Required

**URL Parameters:**
- `id`: Session ID

**Request Body:** Empty (no body required)

**Success Response (200):**
```json
{
  "success": true,
  "message": "Insight generated successfully",
  "data": {
    "insight": "Based on our conversation, you seem to be going through a period of self-reflection..."
  }
}
```

**Frontend Example:**
```javascript
async function generateInsight(sessionId) {
  const token = localStorage.getItem('access_token');
  
  const response = await fetch(
    `http://localhost:8080/api/v1/chat/sessions/${sessionId}/generate-insight`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );
  
  const result = await response.json();
  return result.data.insight;
}
```

---

## Room Service

### 1. Create Room

**Endpoint:** `POST /api/v1/rooms`

**Authentication:** ✅ Required

**Request Body:**
```json
{
  "name": "Pisces Support Group",
  "zodiac_filter": "Pisces"
}
```

**Success Response (201):**
```json
{
  "success": true,
  "message": "Room created successfully",
  "data": {
    "id": "507f1f77bcf86cd799439040",
    "name": "Pisces Support Group",
    "zodiac_filter": "Pisces",
    "created_by": "507f1f77bcf86cd799439011",
    "created_at": "2025-11-29T10:00:00Z"
  }
}
```

---

### 2. Get All Rooms

**Endpoint:** `GET /api/v1/rooms`

**Authentication:** ✅ Required

**Success Response (200):**
```json
{
  "success": true,
  "message": "Rooms retrieved successfully",
  "data": [
    {
      "id": "507f1f77bcf86cd799439040",
      "name": "Pisces Support Group",
      "zodiac_filter": "Pisces",
      "created_by": "507f1f77bcf86cd799439011",
      "created_at": "2025-11-29T10:00:00Z"
    }
  ]
}
```

---

### 3. Join Room (WebSocket)

**Endpoint:** `GET /api/v1/rooms/:id/ws`

**Authentication:** ✅ Required (via query parameter)

**Connection URL:**
```
ws://localhost:8080/api/v1/rooms/507f1f77bcf86cd799439040/ws?token=<access_token>
```

**Frontend Example:**
```javascript
function joinRoom(roomId) {
  const token = localStorage.getItem('access_token');
  const ws = new WebSocket(
    `ws://localhost:8080/api/v1/rooms/${roomId}/ws?token=${token}`
  );
  
  ws.onopen = () => {
    console.log('Connected to room');
  };
  
  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
  };
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };
  
  ws.onclose = () => {
    console.log('Disconnected from room');
  };
  
  return ws;
}

// Send message
function sendRoomMessage(ws, content) {
  ws.send(JSON.stringify({
    type: 'message',
    content: content
  }));
}
```

---

## Social Service

### 1. Publish Post

**Endpoint:** `POST /api/v1/posts`

**Authentication:** ✅ Required

**Request Body:**
```json
{
  "title": "My Zodiac Journey",
  "content": "Today I learned something amazing about my zodiac sign...",
  "mood_tags": ["happy", "inspired"]
}
```

**Success Response (201):**
```json
{
  "success": true,
  "message": "Post published successfully",
  "data": {
    "id": "507f1f77bcf86cd799439050",
    "author_zodiac": "Pisces",
    "title": "My Zodiac Journey",
    "content": "Today I learned something amazing about my zodiac sign...",
    "mood_tags": ["happy", "inspired"],
    "status": "PUBLISHED",
    "likes_count": 0,
    "comments_count": 0,
    "created_at": "2025-11-29T10:00:00Z",
    "updated_at": "2025-11-29T10:00:00Z"
  }
}
```

---

### 2. Get Feed (with Filters & Pagination)

**Endpoint:** `GET /api/v1/posts`

**Authentication:** ❌ Not Required (Public)

**Query Parameters:**
- `cursor` (optional): Cursor untuk pagination
- `limit` (optional): Jumlah posts per page (default: 20)
- `zodiac` (optional): Filter by zodiac sign
- `mood` (optional): Filter by mood tag
- `sort` (optional): Sort by `latest` atau `most_liked` (default: `latest`)

**Example Request:**
```
GET /api/v1/posts?limit=10&zodiac=Pisces&sort=most_liked&cursor=abc123
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Feed retrieved successfully",
  "data": [
    {
      "id": "507f1f77bcf86cd799439050",
      "author_zodiac": "Pisces",
      "title": "My Zodiac Journey",
      "content": "Today I learned something amazing...",
      "mood_tags": ["happy", "inspired"],
      "status": "PUBLISHED",
      "likes_count": 15,
      "comments_count": 3,
      "created_at": "2025-11-29T10:00:00Z",
      "updated_at": "2025-11-29T10:00:00Z"
    }
  ],
  "meta": {
    "next_cursor": "xyz789",
    "has_more": true,
    "limit": 10
  }
}
```

**Frontend Example:**
```javascript
async function getFeed(options = {}) {
  const { cursor = '', limit = 20, zodiac = '', mood = '', sort = 'latest' } = options;
  
  const params = new URLSearchParams({ limit: limit.toString(), sort });
  if (cursor) params.append('cursor', cursor);
  if (zodiac) params.append('zodiac', zodiac);
  if (mood) params.append('mood', mood);
  
  const response = await fetch(`http://localhost:8080/api/v1/posts?${params}`);
  const result = await response.json();
  
  return {
    posts: result.data,
    nextCursor: result.meta?.next_cursor,
    hasMore: result.meta?.has_more
  };
}
```

---

### 3. Get Single Post

**Endpoint:** `GET /api/v1/posts/:id`

**Authentication:** ❌ Not Required (Public)

**URL Parameters:**
- `id`: Post ID

**Success Response (200):**
```json
{
  "success": true,
  "message": "Post retrieved successfully",
  "data": {
    "id": "507f1f77bcf86cd799439050",
    "author_zodiac": "Pisces",
    "title": "My Zodiac Journey",
    "content": "Today I learned something amazing...",
    "mood_tags": ["happy", "inspired"],
    "status": "PUBLISHED",
    "likes_count": 15,
    "comments_count": 3,
    "created_at": "2025-11-29T10:00:00Z",
    "updated_at": "2025-11-29T10:00:00Z"
  }
}
```

---

### 4. Like Post

**Endpoint:** `POST /api/v1/posts/:id/like`

**Authentication:** ✅ Required

**URL Parameters:**
- `id`: Post ID

**Request Body:** Empty (no body required)

**Success Response (200):**
```json
{
  "success": true,
  "message": "Post liked successfully",
  "data": null
}
```

**Error Response (409):**
```json
{
  "success": false,
  "message": "Post already liked",
  "error": {
    "code": "CONFLICT",
    "message": "Post already liked"
  }
}
```

---

### 5. Unlike Post

**Endpoint:** `DELETE /api/v1/posts/:id/like`

**Authentication:** ✅ Required

**URL Parameters:**
- `id`: Post ID

**Success Response (200):**
```json
{
  "success": true,
  "message": "Post unliked successfully",
  "data": null
}
```

---

### 6. Add Comment

**Endpoint:** `POST /api/v1/posts/:id/comments`

**Authentication:** ✅ Required

**URL Parameters:**
- `id`: Post ID

**Request Body:**
```json
{
  "content": "Great post! I totally agree.",
  "parent_id": "507f1f77bcf86cd799439060"
}
```

**Note:** Field `parent_id` optional. Digunakan untuk nested comments (reply to comment).

**Success Response (201):**
```json
{
  "success": true,
  "message": "Comment added successfully",
  "data": {
    "id": "507f1f77bcf86cd799439061",
    "post_id": "507f1f77bcf86cd799439050",
    "user_id": "507f1f77bcf86cd799439011",
    "username": "Johnny",
    "content": "Great post! I totally agree.",
    "parent_id": "507f1f77bcf86cd799439060",
    "created_at": "2025-11-29T10:30:00Z"
  }
}
```

---

### 7. Get Comments

**Endpoint:** `GET /api/v1/posts/:id/comments`

**Authentication:** ❌ Not Required (Public)

**URL Parameters:**
- `id`: Post ID

**Success Response (200):**
```json
{
  "success": true,
  "message": "Comments retrieved successfully",
  "data": [
    {
      "id": "507f1f77bcf86cd799439060",
      "post_id": "507f1f77bcf86cd799439050",
      "user_id": "507f1f77bcf86cd799439012",
      "username": "Jane",
      "content": "Amazing insights!",
      "parent_id": null,
      "created_at": "2025-11-29T10:15:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439061",
      "post_id": "507f1f77bcf86cd799439050",
      "user_id": "507f1f77bcf86cd799439011",
      "username": "Johnny",
      "content": "Great post! I totally agree.",
      "parent_id": "507f1f77bcf86cd799439060",
      "created_at": "2025-11-29T10:30:00Z"
    }
  ]
}
```

---

## AI Service

### 1. Generate Chat Response

**Endpoint:** `POST /api/v1/ai/chat`

**Authentication:** ❌ Not Required (Internal Use)

**Note:** Endpoint ini biasanya dipanggil secara internal oleh Chat Service. Frontend tidak perlu memanggil langsung.

**Request Body:**
```json
{
  "zodiac_sign": "Pisces",
  "user_message": "What's my horoscope for today?"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "AI response generated",
  "data": {
    "response": "As a Pisces, today is a great day for creativity and intuition..."
  }
}
```

---

### 2. Generate Insight

**Endpoint:** `POST /api/v1/ai/insight`

**Authentication:** ❌ Not Required (Internal Use)

**Note:** Endpoint ini dipanggil secara internal oleh Chat Service saat generate insight.

**Request Body:**
```json
{
  "chat_history": "USER: What's my horoscope?\nAI: As a Pisces...\nUSER: Tell me more\nAI: ..."
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Insight generated",
  "data": {
    "insight": "Based on our conversation, you seem to be seeking clarity about your path..."
  }
}
```

---

## Error Handling

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid request body atau parameter |
| `UNAUTHORIZED` | 401 | Token tidak valid atau expired |
| `FORBIDDEN` | 403 | Tidak memiliki akses ke resource |
| `NOT_FOUND` | 404 | Resource tidak ditemukan |
| `CONFLICT` | 409 | Conflict (e.g., email sudah ada, already liked) |
| `TOO_MANY_REQUESTS` | 429 | Rate limit exceeded |
| `INTERNAL_SERVER_ERROR` | 500 | Server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily down |

### Frontend Error Handling Example

```javascript
async function apiCall(url, options = {}) {
  try {
    const response = await fetch(url, options);
    const result = await response.json();
    
    if (!result.success) {
      // Handle API errors
      switch (result.error?.code) {
        case 'UNAUTHORIZED':
          // Try refresh token
          const newToken = await refreshAccessToken();
          if (newToken) {
            // Retry request with new token
            options.headers['Authorization'] = `Bearer ${newToken}`;
            return apiCall(url, options);
          }
          break;
          
        case 'TOO_MANY_REQUESTS':
          alert('Too many requests. Please wait a moment.');
          break;
          
        case 'CONFLICT':
          console.log('Conflict:', result.message);
          break;
          
        default:
          console.error('API Error:', result.message);
      }
      
      throw new Error(result.message);
    }
    
    return result.data;
  } catch (error) {
    console.error('Request failed:', error);
    throw error;
  }
}
```

---

## Common Issues & Troubleshooting

### 1. ❌ Error: "Cannot read properties of null (reading 'nextCursor')"

**Penyebab:** Response tidak memiliki `meta` field atau struktur data tidak sesuai.

**Solusi:**
```javascript
// ❌ SALAH
const nextCursor = result.meta.next_cursor;

// ✅ BENAR
const nextCursor = result.meta?.next_cursor || '';
const hasMore = result.meta?.has_more || false;
```

---

### 2. ❌ Error: "friends.filter is not a function"

**Penyebab:** Response `GET /friends` mengembalikan object `{ friend_ids: [...], count: 3 }`, bukan array.

**Solusi:**
```javascript
// ❌ SALAH
const friends = await getFriends();
const filtered = friends.filter(...); // Error!

// ✅ BENAR
const result = await getFriends();
const friendIds = result.friend_ids; // Array of IDs
const count = result.count;
```

---

### 3. ❌ Error: 401 Unauthorized

**Penyebab:** Token expired atau tidak valid.

**Solusi:**
```javascript
// Implement automatic token refresh
async function fetchWithAuth(url, options = {}) {
  let token = localStorage.getItem('access_token');
  
  options.headers = {
    ...options.headers,
    'Authorization': `Bearer ${token}`
  };
  
  let response = await fetch(url, options);
  
  // If unauthorized, try refresh
  if (response.status === 401) {
    token = await refreshAccessToken();
    if (token) {
      options.headers['Authorization'] = `Bearer ${token}`;
      response = await fetch(url, options);
    }
  }
  
  return response.json();
}
```

---

### 4. ❌ Data Structure Mismatch

**Penyebab:** Frontend expect array tapi backend return object, atau sebaliknya.

**Solusi:** Selalu cek struktur response di `data` field:

```javascript
// Endpoint yang return array
GET /api/v1/chat/sessions
// result.data = [...]

// Endpoint yang return object
GET /api/v1/users/me
// result.data = { id: "...", email: "...", ... }

// Endpoint dengan pagination
GET /api/v1/posts
// result.data = [...]
// result.meta = { next_cursor: "...", has_more: true }
```

---

### 5. ❌ CORS Error

**Penyebab:** Frontend dan backend di domain berbeda.

**Solusi:** Backend sudah setup CORS middleware. Pastikan request include credentials:

```javascript
fetch('http://localhost:8080/api/v1/users/me', {
  credentials: 'include', // Include cookies
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});
```

---

### 6. ❌ WebSocket Connection Failed

**Penyebab:** Token tidak dikirim atau format URL salah.

**Solusi:**
```javascript
// ✅ BENAR - Token di query parameter
const ws = new WebSocket(
  `ws://localhost:8080/api/v1/rooms/${roomId}/ws?token=${token}`
);

// ❌ SALAH - WebSocket tidak support Authorization header
const ws = new WebSocket(`ws://localhost:8080/api/v1/rooms/${roomId}/ws`);
ws.setRequestHeader('Authorization', `Bearer ${token}`); // Tidak work!
```

---

### 7. 🔧 Best Practices

#### A. Always Check Response Structure
```javascript
const result = await response.json();

// Check success field
if (!result.success) {
  console.error('Error:', result.error);
  return;
}

// Access data safely
const data = result.data;
const meta = result.meta; // Might be undefined
```

#### B. Handle Pagination Correctly
```javascript
async function loadMorePosts() {
  const result = await fetch(`/api/v1/posts?cursor=${nextCursor}&limit=20`);
  const json = await result.json();
  
  setPosts(prev => [...prev, ...json.data]);
  setNextCursor(json.meta?.next_cursor || '');
  setHasMore(json.meta?.has_more || false);
}
```

#### C. Store User Data Properly
```javascript
// After login/register
localStorage.setItem('access_token', result.data.access_token);
localStorage.setItem('refresh_token', result.data.refresh_token);
localStorage.setItem('user', JSON.stringify(result.data.user));

// Retrieve user data
const user = JSON.parse(localStorage.getItem('user'));
console.log(user.zodiac_sign); // "Pisces"
```

#### D. Handle Date Formats
```javascript
// Backend mengirim ISO 8601 format
const dateOfBirth = "1995-03-15T00:00:00Z";

// Convert to Date object
const date = new Date(dateOfBirth);

// Send to backend
const payload = {
  date_of_birth: new Date('1995-03-15').toISOString()
};
```

---

## Rate Limiting

API menggunakan rate limiting untuk mencegah abuse:
- **Limit:** 100 requests per 15 menit per IP
- **Response saat limit exceeded:**

```json
{
  "success": false,
  "message": "Too many requests",
  "error": {
    "code": "TOO_MANY_REQUESTS",
    "message": "Too many requests"
  }
}
```

**Headers yang dikembalikan:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1638360000
```

---

## Testing Endpoints

### Using cURL

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User",
    "date_of_birth": "1995-03-15T00:00:00Z",
    "gender": "male"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Get Profile (with token)
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Using Postman

1. Import collection dari file (jika ada)
2. Set environment variable `BASE_URL` = `http://localhost:8080/api/v1`
3. Set environment variable `TOKEN` setelah login
4. Gunakan `{{BASE_URL}}` dan `{{TOKEN}}` di requests

---

## Support & Contact

Jika ada pertanyaan atau menemukan bug:
1. Check dokumentasi ini terlebih dahulu
2. Check [Common Issues](#common-issues--troubleshooting)
3. Contact developer team

---

**Last Updated:** 2025-11-29
**API Version:** v1
**Backend Version:** All-in-One Deployment
