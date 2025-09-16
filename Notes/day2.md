manager.go , password.go , store.go 

# Authentication Project – Notes (Day X)

This file documents the concepts I learned while building my **Auth & Tokens Done Right** project.

---

## 🔑 Key Concepts Learned

### 1. User Store (in-memory DB)
- The **`Store`** struct acts like a temporary database:
  - Keeps all registered users in memory (`map[email]*User`).
  - Auto-increments `ID` for each new user with the `seq` counter.
  - Provides `Create()` for signup and `GetByEmail()` for login lookups.

### 2. User Struct
- Represents a single registered user:
  - `ID` → unique integer, assigned by `seq`.
  - `Email` → unique email used for login.
  - `PasswordHash` → bcrypt-hashed password.
  - `Role` → user role (default `"user"`).

### 3. Password Hashing
- Raw passwords are **never stored**.
- Bcrypt is used:
  - `HashPassword(plain)` → returns salted, hashed password string.
  - `VerifyPassword(plain, hash)` → checks login password against stored hash.
- **Cost** controls bcrypt’s work factor (e.g., 12). Higher cost = more secure, but slower.

### 4. Concurrency Safety (RWMutex)
- Go maps are **not safe** if one goroutine writes while another reads → panic: *“concurrent map read and map write”*.
- Fix: use `sync.RWMutex` inside the store.
  - **`RLock()`** → for reads (`GetByEmail`), many readers allowed in parallel.
  - **`Lock()`** → for writes (`Create`), only one writer allowed, blocks readers until done.
- This prevents race conditions when multiple users sign up or log in at the same time.

### 5. Predefined Errors
- `ErrEmailExists` → returned when trying to create a user with a duplicate email.
- `ErrNotFound` → returned when looking up an email that doesn’t exist.
- Handlers check these errors and return the proper JSON + HTTP status.

---

## 🧭 Flow of Store Functions

1. **Signup (`Create`)**
   - Handler hashes password.
   - Calls `userStore.Create(email, hash)`.
   - Store:
     - Locks with `Lock()`.
     - Increments `seq` → assigns new `User.ID`.
     - Saves user in `byEmail[email]`.
   - Returns `User` (without password hash).

2. **Login (`GetByEmail`)**
   - Handler receives email/password.
   - Calls `userStore.GetByEmail(email)`.
   - Store:
     - Locks with `RLock()`.
     - Looks up user in `byEmail[email]`.
   - Handler compares password with `PasswordHash` using bcrypt.

---

## ✅ Beginner Takeaways
- **Store** = in-memory mini database of users.
- **User struct** = one record per user.
- **Password hashing** = bcrypt with cost (default 10, better 12+).
- **RWMutex** = prevents crashes and ensures thread-safety.
- **Errors** = use named error variables for clarity and handler logic.
