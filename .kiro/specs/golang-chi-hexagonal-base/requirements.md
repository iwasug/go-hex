# Requirements Document

## Introduction

Proyek ini adalah base project Golang menggunakan framework Go Chi dengan arsitektur Hexagonal (Ports and Adapters). Sistem ini akan menggunakan PostgreSQL sebagai database tanpa ORM, dengan implementasi raw SQL queries. Proyek ini mencakup tiga fitur utama: Authentication (login, register, forgot password, refresh token), User Management, dan Role Management dengan RBAC (Role-Based Access Control).

Arsitektur Hexagonal akan memisahkan business logic dari infrastructure concerns, memungkinkan fleksibilitas dalam mengganti implementasi database atau framework tanpa mengubah core business logic.

## Requirements

### Requirement 1: Project Structure dan Setup

**User Story:** Sebagai developer, saya ingin memiliki struktur project yang mengikuti Hexagonal Architecture, sehingga kode mudah dimaintain dan di-test.

#### Acceptance Criteria

1. WHEN project diinisialisasi THEN sistem SHALL memiliki struktur folder yang jelas memisahkan domain, ports, adapters, dan infrastructure
2. WHEN project dijalankan THEN sistem SHALL dapat terkoneksi ke PostgreSQL database
3. WHEN environment variables diload THEN sistem SHALL membaca konfigurasi dari file .env
4. WHEN aplikasi dijalankan THEN sistem SHALL menjalankan HTTP server menggunakan Go Chi router

### Requirement 2: Authentication - Register

**User Story:** Sebagai user baru, saya ingin mendaftar akun dengan email dan password, sehingga saya dapat mengakses sistem.

#### Acceptance Criteria

1. WHEN user mengirim POST request ke /api/v1/auth/register dengan email dan password THEN sistem SHALL membuat user baru di database
2. WHEN user mendaftar dengan email yang sudah ada THEN sistem SHALL mengembalikan error "email already exists"
3. WHEN user mendaftar THEN sistem SHALL meng-hash password menggunakan bcrypt sebelum menyimpan ke database
4. WHEN user berhasil mendaftar THEN sistem SHALL mengembalikan response dengan status 201 dan data user (tanpa password)
5. WHEN user mengirim data registrasi yang tidak valid THEN sistem SHALL mengembalikan error validation dengan status 400
6. WHEN user mendaftar THEN sistem SHALL assign default role "user" ke user tersebut

### Requirement 3: Authentication - Login

**User Story:** Sebagai registered user, saya ingin login dengan email dan password, sehingga saya mendapatkan access token untuk mengakses protected resources.

#### Acceptance Criteria

1. WHEN user mengirim POST request ke /api/v1/auth/login dengan email dan password yang benar THEN sistem SHALL mengembalikan access token dan refresh token
2. WHEN user login dengan credentials yang salah THEN sistem SHALL mengembalikan error "invalid credentials" dengan status 401
3. WHEN user berhasil login THEN sistem SHALL membuat JWT access token dengan expiry 15 menit
4. WHEN user berhasil login THEN sistem SHALL membuat JWT refresh token dengan expiry 7 hari
5. WHEN user berhasil login THEN sistem SHALL menyimpan refresh token di database
6. WHEN access token dibuat THEN token SHALL berisi user_id, email, dan roles dalam payload

### Requirement 4: Authentication - Refresh Token

**User Story:** Sebagai authenticated user, saya ingin me-refresh access token yang expired menggunakan refresh token, sehingga saya tidak perlu login ulang.

#### Acceptance Criteria

1. WHEN user mengirim POST request ke /api/v1/auth/refresh dengan valid refresh token THEN sistem SHALL mengembalikan access token baru
2. WHEN user mengirim invalid atau expired refresh token THEN sistem SHALL mengembalikan error dengan status 401
3. WHEN refresh token digunakan THEN sistem SHALL memverifikasi token ada di database dan belum expired
4. WHEN access token baru dibuat THEN sistem SHALL menggunakan data user terbaru dari database

### Requirement 5: Authentication - Forgot Password

**User Story:** Sebagai user yang lupa password, saya ingin request reset password link, sehingga saya dapat membuat password baru.

#### Acceptance Criteria

1. WHEN user mengirim POST request ke /api/v1/auth/forgot-password dengan email THEN sistem SHALL membuat reset token dan menyimpannya di database
2. WHEN reset token dibuat THEN token SHALL memiliki expiry 1 jam
3. WHEN user mengirim POST request ke /api/v1/auth/reset-password dengan valid reset token dan password baru THEN sistem SHALL update password user
4. WHEN reset token sudah digunakan atau expired THEN sistem SHALL mengembalikan error
5. WHEN password berhasil direset THEN sistem SHALL menghapus reset token dari database
6. WHEN email tidak ditemukan di forgot-password request THEN sistem SHALL tetap mengembalikan success response (untuk security reasons)

### Requirement 6: User Management - Get Users

**User Story:** Sebagai admin, saya ingin melihat daftar users dengan fitur search, sorting, dan pagination, sehingga saya dapat mengelola users di sistem dengan mudah.

#### Acceptance Criteria

1. WHEN admin mengirim GET request ke /api/v1/users THEN sistem SHALL mengembalikan list users dengan pagination
2. WHEN request memiliki query parameter page dan limit THEN sistem SHALL mengembalikan users sesuai pagination
3. WHEN request memiliki query parameter search THEN sistem SHALL mengembalikan users yang name atau email-nya match dengan search term
4. WHEN request memiliki query parameter order_by THEN sistem SHALL mengurutkan users berdasarkan field tersebut (name, email, created_at)
5. WHEN request memiliki query parameter order THEN sistem SHALL mengurutkan dengan direction asc atau desc
6. WHEN order_by tidak valid THEN sistem SHALL mengembalikan error validation
7. WHEN user yang bukan admin mencoba akses endpoint ini THEN sistem SHALL mengembalikan error 403 Forbidden
8. WHEN users dikembalikan THEN response SHALL tidak include password hash
9. WHEN users dikembalikan THEN response SHALL include role information untuk setiap user
10. WHEN users dikembalikan THEN response SHALL include pagination metadata (total, page, limit, total_pages)

### Requirement 7: User Management - Get User by ID

**User Story:** Sebagai admin, saya ingin melihat detail user tertentu, sehingga saya dapat melihat informasi lengkap user tersebut.

#### Acceptance Criteria

1. WHEN admin mengirim GET request ke /api/v1/users/:id THEN sistem SHALL mengembalikan detail user
2. WHEN user ID tidak ditemukan THEN sistem SHALL mengembalikan error 404
3. WHEN user yang bukan admin mencoba akses user lain THEN sistem SHALL mengembalikan error 403
4. WHEN user mengakses data dirinya sendiri THEN sistem SHALL mengizinkan akses

### Requirement 8: User Management - Update User

**User Story:** Sebagai user, saya ingin update profile saya, sehingga informasi saya tetap up-to-date.

#### Acceptance Criteria

1. WHEN user mengirim PUT request ke /api/v1/users/:id dengan data baru THEN sistem SHALL update user di database dan set updated_by ke user ID yang melakukan update
2. WHEN user mencoba update user lain dan bukan admin THEN sistem SHALL mengembalikan error 403
3. WHEN admin update user THEN sistem SHALL dapat mengubah semua field termasuk roles
4. WHEN regular user update profile sendiri THEN sistem SHALL tidak dapat mengubah roles
5. WHEN email diubah ke email yang sudah digunakan user lain THEN sistem SHALL mengembalikan error
6. WHEN user di-update THEN sistem SHALL set updated_at ke timestamp sekarang

### Requirement 9: User Management - Delete User

**User Story:** Sebagai admin, saya ingin menghapus user (soft delete), sehingga user yang tidak aktif dapat di-nonaktifkan dari sistem tanpa menghilangkan data historis.

#### Acceptance Criteria

1. WHEN admin mengirim DELETE request ke /api/v1/users/:id THEN sistem SHALL soft delete user (set is_deleted=true, deleted_at=now, deleted_by=admin_id)
2. WHEN user yang bukan admin mencoba delete user THEN sistem SHALL mengembalikan error 403
3. WHEN user ID tidak ditemukan THEN sistem SHALL mengembalikan error 404
4. WHEN user di-soft delete THEN sistem SHALL juga menghapus semua refresh tokens milik user tersebut
5. WHEN user sudah di-soft delete THEN user tidak dapat login
6. WHEN list users dipanggil tanpa include_deleted parameter THEN soft deleted users tidak muncul di hasil

### Requirement 10: Role Management - Get Roles

**User Story:** Sebagai admin, saya ingin melihat daftar roles yang tersedia dengan fitur search dan sorting, sehingga saya dapat mengelola roles di sistem dengan mudah.

#### Acceptance Criteria

1. WHEN admin mengirim GET request ke /api/v1/roles THEN sistem SHALL mengembalikan list semua roles
2. WHEN request memiliki query parameter search THEN sistem SHALL mengembalikan roles yang name atau description-nya match dengan search term
3. WHEN request memiliki query parameter order_by THEN sistem SHALL mengurutkan roles berdasarkan field tersebut (name, created_at)
4. WHEN request memiliki query parameter order THEN sistem SHALL mengurutkan dengan direction asc atau desc
5. WHEN order_by tidak valid THEN sistem SHALL mengembalikan error validation
6. WHEN user yang bukan admin mencoba akses endpoint ini THEN sistem SHALL mengembalikan error 403
7. WHEN roles dikembalikan THEN response SHALL include permissions untuk setiap role

### Requirement 11: Role Management - Create Role

**User Story:** Sebagai admin, saya ingin membuat role baru dengan permissions tertentu, sehingga saya dapat mengatur akses control yang lebih granular.

#### Acceptance Criteria

1. WHEN admin mengirim POST request ke /api/v1/roles dengan nama dan permissions THEN sistem SHALL membuat role baru
2. WHEN role dengan nama yang sama sudah ada THEN sistem SHALL mengembalikan error
3. WHEN role dibuat THEN sistem SHALL menyimpan permissions yang di-assign ke role tersebut
4. WHEN user yang bukan admin mencoba create role THEN sistem SHALL mengembalikan error 403

### Requirement 12: Role Management - Update Role

**User Story:** Sebagai admin, saya ingin update role dan permissions-nya, sehingga saya dapat menyesuaikan access control sesuai kebutuhan.

#### Acceptance Criteria

1. WHEN admin mengirim PUT request ke /api/v1/roles/:id dengan data baru THEN sistem SHALL update role di database
2. WHEN role ID tidak ditemukan THEN sistem SHALL mengembalikan error 404
3. WHEN permissions diubah THEN sistem SHALL update relasi role-permissions di database
4. WHEN user yang bukan admin mencoba update role THEN sistem SHALL mengembalikan error 403

### Requirement 13: Role Management - Delete Role

**User Story:** Sebagai admin, saya ingin menghapus role yang tidak digunakan (soft delete), sehingga sistem tetap clean dan organized tanpa menghilangkan data historis.

#### Acceptance Criteria

1. WHEN admin mengirim DELETE request ke /api/v1/roles/:id THEN sistem SHALL soft delete role (set is_deleted=true, deleted_at=now, deleted_by=admin_id)
2. WHEN role masih digunakan oleh users THEN sistem SHALL mengembalikan error "role is in use"
3. WHEN role ID tidak ditemukan THEN sistem SHALL mengembalikan error 404
4. WHEN user yang bukan admin mencoba delete role THEN sistem SHALL mengembalikan error 403
5. WHEN list roles dipanggil tanpa include_deleted parameter THEN soft deleted roles tidak muncul di hasil

### Requirement 14: Role Management - Assign Role to User

**User Story:** Sebagai admin, saya ingin assign role ke user, sehingga user memiliki permissions sesuai role-nya.

#### Acceptance Criteria

1. WHEN admin mengirim POST request ke /api/v1/users/:id/roles dengan role_id THEN sistem SHALL assign role ke user
2. WHEN user sudah memiliki role tersebut THEN sistem SHALL mengembalikan error "role already assigned"
3. WHEN role_id tidak valid THEN sistem SHALL mengembalikan error 404
4. WHEN user yang bukan admin mencoba assign role THEN sistem SHALL mengembalikan error 403

### Requirement 15: Role Management - Remove Role from User

**User Story:** Sebagai admin, saya ingin remove role dari user, sehingga user tidak lagi memiliki permissions dari role tersebut.

#### Acceptance Criteria

1. WHEN admin mengirim DELETE request ke /api/v1/users/:id/roles/:role_id THEN sistem SHALL remove role dari user
2. WHEN user tidak memiliki role tersebut THEN sistem SHALL mengembalikan error 404
3. WHEN user yang bukan admin mencoba remove role THEN sistem SHALL mengembalikan error 403

### Requirement 16: Authorization Middleware

**User Story:** Sebagai sistem, saya ingin memverifikasi JWT token dan permissions pada setiap protected request, sehingga hanya authorized users yang dapat akses resources.

#### Acceptance Criteria

1. WHEN request ke protected endpoint tidak memiliki Authorization header THEN sistem SHALL mengembalikan error 401
2. WHEN JWT token invalid atau expired THEN sistem SHALL mengembalikan error 401
3. WHEN JWT token valid THEN sistem SHALL extract user information dan inject ke request context
4. WHEN endpoint memerlukan specific permission dan user tidak memilikinya THEN sistem SHALL mengembalikan error 403
5. WHEN user memiliki required permission THEN sistem SHALL mengizinkan akses ke endpoint

### Requirement 17: Database Migrations

**User Story:** Sebagai developer, saya ingin memiliki database migration system, sehingga schema database dapat di-manage dengan baik.

#### Acceptance Criteria

1. WHEN migration dijalankan THEN sistem SHALL membuat tables: users, roles, permissions, user_roles, role_permissions, refresh_tokens, password_reset_tokens
2. WHEN migration dijalankan THEN sistem SHALL membuat default roles (admin, user) dan permissions
3. WHEN migration dijalankan THEN sistem SHALL membuat indexes yang diperlukan untuk performa
4. WHEN migration dijalankan THEN sistem SHALL membuat foreign key constraints untuk referential integrity

### Requirement 18: Error Handling dan Logging

**User Story:** Sebagai developer, saya ingin sistem memiliki error handling dan logging yang konsisten, sehingga mudah untuk debugging dan monitoring.

#### Acceptance Criteria

1. WHEN error terjadi THEN sistem SHALL mengembalikan response dengan format yang konsisten (error code, message, details)
2. WHEN error terjadi THEN sistem SHALL log error dengan level yang sesuai (error, warn, info)
3. WHEN request masuk THEN sistem SHALL log request method, path, dan response time
4. WHEN panic terjadi THEN sistem SHALL recover dan mengembalikan error 500 tanpa crash
