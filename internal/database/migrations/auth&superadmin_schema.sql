-- ============================================
-- EXTENSIONS
-- ============================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- ============================================
-- ENUMS
-- ============================================
CREATE TYPE user_role AS ENUM (
    'super_admin',    -- Full system access across all branches
    'admin',          -- Branch level full access
    'doctor',         -- View patients, appointments
    'receptionist',   -- Manage patients, appointments
    'billing_staff',  -- Manage billing only
    'pharmacist'      -- View prescriptions
);

CREATE TYPE user_status AS ENUM (
    'active',
    'inactive',
    'suspended',
    'pending'
);

CREATE TYPE token_type AS ENUM (
    'refresh',
    'reset_password',
    'email_verify'
);


-- ============================================
-- BRANCHES TABLE
-- Each clinic branch has its own record
-- ============================================
CREATE TABLE branches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(150) NOT NULL,
    code            VARCHAR(20) NOT NULL UNIQUE,  -- e.g. BRANCH01 (used in mobile app)
    address         TEXT,
    phone           VARCHAR(20),
    email           VARCHAR(100),
    logo_url        TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);


-- ============================================
-- USERS TABLE
-- ============================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id       UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    email           VARCHAR(150) NOT NULL UNIQUE,
    phone           VARCHAR(20),
    password_hash   TEXT NOT NULL,
    role            user_role NOT NULL DEFAULT 'receptionist',
    status          user_status NOT NULL DEFAULT 'active',
    avatar_url      TEXT,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);


-- ============================================
-- TOKENS TABLE
-- Refresh tokens, reset password, email verify
-- ============================================
CREATE TABLE tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL UNIQUE,           -- Store hashed token
    token_type      token_type NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    is_used         BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);


-- ============================================
-- SESSIONS TABLE
-- Track active login sessions per device
-- ============================================
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL UNIQUE,           -- Hashed refresh token
    device_name     VARCHAR(100),                   -- e.g. "iPhone 14", "Chrome Browser"
    device_type     VARCHAR(50),                    -- e.g. "mobile", "desktop", "tablet"
    ip_address      INET,
    user_agent      TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    last_active_at  TIMESTAMPTZ DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);


-- ============================================
-- LOGIN AUDIT LOG
-- Track all login attempts (success + failed)
-- ============================================
CREATE TABLE login_audit (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    email           VARCHAR(150),                   -- Store email even if user not found
    ip_address      INET,
    user_agent      TEXT,
    device_type     VARCHAR(50),
    status          VARCHAR(20) NOT NULL,           -- 'success', 'failed', 'blocked'
    failure_reason  VARCHAR(100),                   -- 'wrong_password', 'user_inactive' etc
    created_at      TIMESTAMPTZ DEFAULT NOW()
);


-- ============================================
-- PERMISSIONS TABLE (Optional - for fine control)
-- ============================================
CREATE TABLE role_permissions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role            user_role NOT NULL,
    resource        VARCHAR(100) NOT NULL,          -- e.g. 'billing', 'patient', 'report'
    can_view        BOOLEAN DEFAULT FALSE,
    can_create      BOOLEAN DEFAULT FALSE,
    can_edit        BOOLEAN DEFAULT FALSE,
    can_delete      BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role, resource)
);


-- ============================================
-- MENUS TABLE
-- Sidebar/navigation menus with tree structure
-- ============================================
CREATE TABLE menus (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id       UUID REFERENCES menus(id) ON DELETE CASCADE,
    label           VARCHAR(100) NOT NULL,
    path            VARCHAR(200),              -- NULL for parent menus that have children
    resource        VARCHAR(100),              -- matches role_permissions resource column
    sort_order      INT DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);


-- ============================================
-- INDEXES
-- ============================================

-- Users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_branch_id ON users(branch_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

-- Tokens
CREATE INDEX idx_tokens_user_id ON tokens(user_id);
CREATE INDEX idx_tokens_token_hash ON tokens(token_hash);
CREATE INDEX idx_tokens_expires_at ON tokens(expires_at);

-- Sessions
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_is_active ON sessions(is_active);

-- Login audit
CREATE INDEX idx_login_audit_user_id ON login_audit(user_id);
CREATE INDEX idx_login_audit_email ON login_audit(email);
CREATE INDEX idx_login_audit_created_at ON login_audit(created_at);

-- Branches
CREATE INDEX idx_branches_code ON branches(code);

-- Menus
CREATE INDEX idx_menus_parent_id ON menus(parent_id);
CREATE INDEX idx_menus_resource ON menus(resource);
CREATE INDEX idx_menus_sort_order ON menus(sort_order);
CREATE INDEX idx_menus_is_active ON menus(is_active);


-- ============================================
-- AUTO UPDATE updated_at TRIGGER
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_branches_updated_at
    BEFORE UPDATE ON branches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_menus_updated_at
    BEFORE UPDATE ON menus
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();


-- ============================================
-- SEED DEFAULT BRANCH + SUPER ADMIN
-- ============================================

-- Insert default branch
INSERT INTO branches (id, name, code, address, phone, email)
VALUES (
    uuid_generate_v4(),
    'Main Branch',
    'BRANCH01',
    '123 Clinic Street',
    '+1234567890',
    'main@yourclinic.com'
);

-- Insert super admin (password: Admin@123 - change immediately)
INSERT INTO users (
    branch_id,
    first_name,
    last_name,
    email,
    password_hash,
    role,
    status
)
VALUES (
    (SELECT id FROM branches WHERE code = 'BRANCH01'),
    'Super',
    'Admin',
    'admin@yourclinic.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj4J/HS.iK8i', -- Admin@123
    'super_admin',
    'active'
);

-- Seed default role permissions
INSERT INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete) VALUES
('super_admin',   'patient',         TRUE, TRUE,  TRUE,  TRUE),
('super_admin',   'medical_history', TRUE, TRUE,  TRUE,  TRUE),
('super_admin',   'lab_result',      TRUE, TRUE,  TRUE,  TRUE),
('super_admin',   'billing',         TRUE, TRUE,  TRUE,  TRUE),
('super_admin',   'appointment',     TRUE, TRUE,  TRUE,  TRUE),
('super_admin',   'doctor',          TRUE, TRUE,  TRUE,  TRUE),
('super_admin',   'report',          TRUE, TRUE,  TRUE,  TRUE),
('super_admin',   'settings',        TRUE, TRUE,  TRUE,  TRUE),

('admin',         'patient',         TRUE, TRUE,  TRUE,  TRUE),
('admin',         'medical_history', TRUE, TRUE,  TRUE,  TRUE),
('admin',         'lab_result',      TRUE, TRUE,  TRUE,  TRUE),
('admin',         'billing',         TRUE, TRUE,  TRUE,  TRUE),
('admin',         'appointment',     TRUE, TRUE,  TRUE,  TRUE),
('admin',         'doctor',          TRUE, TRUE,  TRUE,  FALSE),
('admin',         'report',          TRUE, FALSE, FALSE, FALSE),
('admin',         'settings',        TRUE, TRUE,  TRUE,  FALSE),

('doctor',        'patient',         TRUE, FALSE, FALSE, FALSE),
('doctor',        'medical_history', TRUE, TRUE,  TRUE,  FALSE),
('doctor',        'lab_result',      TRUE, TRUE,  TRUE,  FALSE),
('doctor',        'appointment',     TRUE, TRUE,  TRUE,  FALSE),
('doctor',        'billing',         TRUE, FALSE, FALSE, FALSE),

('receptionist',  'patient',         TRUE, TRUE,  TRUE,  FALSE),
('receptionist',  'medical_history', TRUE, FALSE, FALSE, FALSE),
('receptionist',  'lab_result',      TRUE, FALSE, FALSE, FALSE),
('receptionist',  'appointment',     TRUE, TRUE,  TRUE,  FALSE),
('receptionist',  'billing',         TRUE, FALSE, FALSE, FALSE),

('billing_staff', 'billing',         TRUE, TRUE,  TRUE,  FALSE),
('billing_staff', 'patient',         TRUE, FALSE, FALSE, FALSE),

('pharmacist',    'patient',         TRUE, FALSE, FALSE, FALSE),
('pharmacist',    'medical_history', TRUE, FALSE, FALSE, FALSE);

-- Seed default menus (sidebar navigation)
-- Single-page menus: leaf items link directly to a page
-- Parent menus: only used when children are genuinely different features/views

-- Top level menus (parent_id = NULL)
INSERT INTO menus (id, parent_id, label, path, resource, sort_order) VALUES
(uuid_generate_v4(), NULL, 'Dashboard',    '/dashboard',    NULL,          1),
(uuid_generate_v4(), NULL, 'Patients',     NULL,            'patient',     2),
(uuid_generate_v4(), NULL, 'Appointments', '/appointments', 'appointment', 3),
(uuid_generate_v4(), NULL, 'Billing',      '/billing',      'billing',     4),
(uuid_generate_v4(), NULL, 'Doctors',      '/doctors',      'doctor',      5),
(uuid_generate_v4(), NULL, 'Reports',      '/reports',      'report',      6),
(uuid_generate_v4(), NULL, 'Settings',     '/settings',     'settings',    7);

-- Sub menus for Patients (genuinely different features/views)
INSERT INTO menus (parent_id, label, path, resource, sort_order) VALUES
((SELECT id FROM menus WHERE label = 'Patients' AND parent_id IS NULL), 'Patient Records',  '/patients',          'patient', 1),
((SELECT id FROM menus WHERE label = 'Patients' AND parent_id IS NULL), 'Medical History',  '/patients/history',  'medical_history', 2),
((SELECT id FROM menus WHERE label = 'Patients' AND parent_id IS NULL), 'Lab Results',      '/patients/labs',     'lab_result', 3);


-- ============================================
-- SCHEMA OVERVIEW
-- ============================================
--
-- branches          → Each clinic branch (identified by code e.g. BRANCH01)
--     │
--     └── users     → Staff accounts linked to a branch
--             │
--             ├── tokens    → Reset password / email verify tokens
--             ├── sessions  → Active login sessions (per device)
--             └── login_audit → All login attempts logged
--
-- role_permissions  → What each role can do per resource
-- menus             → Sidebar navigation with tree structure (filtered by role_permissions)