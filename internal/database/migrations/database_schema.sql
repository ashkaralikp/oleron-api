-- ============================================
-- EXTENSIONS
-- ============================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- ============================================
-- ENUMS
-- ============================================
CREATE TYPE user_role AS ENUM (
    'super_admin',       -- Full system access across all branches
    'regional_manager',  -- Branch level full access (Manager)
    'manager',        -- Manages employees within a branch
    'employee',       -- Punches in/out; views own attendance and salary
    'consultant'      -- External consultant with limited access (e.g. payroll only)
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

CREATE TYPE attendance_status AS ENUM (
    'present',            -- Punched in on time, punched out at regular time
    'absent',             -- No punch-in recorded for the day
    'half_day',           -- Worked less than half of expected hours
    'late_in',            -- Punched in after expected start time
    'early_out',          -- Punched out before expected end time
    'late_in_early_out',  -- Both late punch-in AND early punch-out
    'on_leave'            -- Employee was on approved leave
);


-- ============================================
-- BRANCHES TABLE
-- Each organization branch
-- ============================================
CREATE TABLE branches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    office_timing_id UUID REFERENCES office_timings(id) ON DELETE SET NULL,
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
    role            user_role NOT NULL DEFAULT 'employee',
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
-- EMPLOYEES TABLE
-- HR profile linked to a user account
-- ============================================
CREATE TABLE employees (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    branch_id        UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    office_timing_id UUID REFERENCES office_timings(id) ON DELETE SET NULL,  -- overrides branch default timing; NULL = use branch timing
    manager_id       UUID REFERENCES users(id) ON DELETE SET NULL,  -- manager of this employee
    employee_code    VARCHAR(20) NOT NULL UNIQUE,                   -- used for mobile punch-in/out
    designation      VARCHAR(100),
    employment_type  VARCHAR(20) DEFAULT 'full_time',               -- full_time, part_time, contract, consultant
    fixed_monthly_salary  NUMERIC(10,2),   -- paid when full month is worked
    ot_rate               NUMERIC(10,2),   -- per OT hour, on top of fixed
    currency         VARCHAR(3) NOT NULL DEFAULT 'USD',  -- e.g. USD, INR, EUR
    joining_date     DATE NOT NULL,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);


-- ============================================
-- OFFICE TIMINGS TABLE
-- Weekly work schedule assigned to a branch
-- ============================================
CREATE TABLE office_timings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id       UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,          -- e.g. "Standard Week", "Night Shift"
    is_active       BOOLEAN DEFAULT TRUE,           -- only one active timing per branch recommended
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- OFFICE TIMING DAYS TABLE
-- Per-day schedule for each office timing
-- ============================================
CREATE TABLE office_timing_days (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    office_timing_id  UUID NOT NULL REFERENCES office_timings(id) ON DELETE CASCADE,
    day_of_week       SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sun, 1=Mon, ..., 6=Sat
    is_working_day    BOOLEAN DEFAULT TRUE,
    start_time        TIME,                         -- e.g. 09:00:00
    end_time          TIME,                         -- e.g. 18:00:00
    break_minutes     SMALLINT DEFAULT 0,           -- break duration in minutes
    UNIQUE(office_timing_id, day_of_week)           -- one entry per day per timing
);


-- ============================================
-- BRANCH CALENDAR TABLE
-- Per-branch overrides to the weekly schedule for a given date
-- Used to mark public holidays, branch closures, or makeup working days
-- ============================================
CREATE TYPE calendar_day_type AS ENUM (
    'holiday',      -- non-working: public holiday, branch closure, etc.
    'working_day'   -- override: this date is a working day (e.g. makeup Saturday)
);

CREATE TABLE branch_calendar (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id   UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    date        DATE NOT NULL,
    type        calendar_day_type NOT NULL,
    name        VARCHAR(150),             -- e.g. "Christmas Day", "Makeup Saturday"
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(branch_id, date)               -- one entry per branch per date
);


-- ============================================
-- ATTENDANCE TABLE
-- Employee punch-in / punch-out records
-- ============================================
CREATE TABLE attendance (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- branch_id derived via users.branch_id
    work_date       DATE NOT NULL,
    punch_in        TIMESTAMPTZ,
    punch_out       TIMESTAMPTZ,
    work_hours      NUMERIC(5,2),                  -- computed on punch-out: (punch_out - punch_in) in hours
    status          attendance_status NOT NULL DEFAULT 'absent', -- set to absent on record creation; updated on punch_in/punch_out
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, work_date)                     -- one record per user per day
);


-- ============================================
-- PAYROLL TABLES
-- payroll_runs  → one record per branch per pay period
-- payroll_items → one record per employee per run
-- ============================================
CREATE TYPE payroll_status AS ENUM (
    'draft',    -- Generated, not yet finalised
    'approved', -- Approved by admin/manager
    'paid'      -- Payment processed
);

CREATE TYPE work_order_status AS ENUM (
    'draft',     -- Created but not finalized
    'issued',    -- Ready to share / generate PDF
    'cancelled'  -- Cancelled after creation
);

CREATE TYPE invoice_status AS ENUM (
    'draft',      -- Created but not finalized
    'issued',     -- Sent/generated for the customer
    'cancelled',  -- Cancelled before payment completion
    'void'        -- Voided after issue for audit trail
);

CREATE TYPE invoice_payment_status AS ENUM (
    'unpaid',         -- No confirmed payment yet
    'partially_paid', -- Confirmed payments are less than total
    'paid',           -- Confirmed payments cover the invoice total
    'overpaid',       -- Confirmed payments exceed the invoice total
    'refunded'        -- Invoice has been fully refunded
);

CREATE TYPE invoice_payment_method AS ENUM (
    'bank_transfer',
    'cash',
    'cheque',
    'credit_card',
    'debit_card',
    'upi',
    'other'
);

CREATE TYPE invoice_payment_record_status AS ENUM (
    'pending',   -- Added but not verified
    'confirmed', -- Verified/accepted
    'failed',    -- Payment failed
    'reversed',  -- Chargeback/refund/reversal
    'rejected'   -- Uploaded proof or payment claim rejected
);

CREATE TABLE payroll_runs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id       UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    period_from     DATE NOT NULL,
    period_to       DATE NOT NULL,
    generated_by    UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status          payroll_status NOT NULL DEFAULT 'draft',
    total_amount    NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(branch_id, period_from, period_to)   -- prevent duplicate run for the same period
);

CREATE TABLE payroll_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payroll_run_id  UUID NOT NULL REFERENCES payroll_runs(id) ON DELETE CASCADE,
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    working_days    SMALLINT NOT NULL DEFAULT 0,    -- expected working days in the period
    present_days    SMALLINT NOT NULL DEFAULT 0,
    absent_days     SMALLINT NOT NULL DEFAULT 0,
    leave_days      SMALLINT NOT NULL DEFAULT 0,
    total_hours     NUMERIC(8,2) NOT NULL DEFAULT 0,
    hourly_rate     NUMERIC(10,2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    gross_pay       NUMERIC(12,2) NOT NULL DEFAULT 0,
    deductions      NUMERIC(12,2) NOT NULL DEFAULT 0,
    net_pay         NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(payroll_run_id, employee_id)             -- one item per employee per run
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

-- Employees
CREATE INDEX idx_employees_user_id ON employees(user_id);
CREATE INDEX idx_employees_branch_id ON employees(branch_id);
CREATE INDEX idx_employees_office_timing_id ON employees(office_timing_id);
CREATE INDEX idx_employees_manager_id ON employees(manager_id);
CREATE INDEX idx_employees_employee_code ON employees(employee_code);

-- Office timings
CREATE INDEX idx_office_timings_branch_id ON office_timings(branch_id);
CREATE INDEX idx_office_timings_is_active ON office_timings(is_active);
CREATE INDEX idx_office_timing_days_timing_id ON office_timing_days(office_timing_id);

-- Branch calendar
CREATE INDEX idx_branch_calendar_branch_id ON branch_calendar(branch_id);
CREATE INDEX idx_branch_calendar_date ON branch_calendar(date);

-- Attendance
CREATE INDEX idx_attendance_user_id ON attendance(user_id);
CREATE INDEX idx_attendance_work_date ON attendance(work_date);
CREATE INDEX idx_attendance_status ON attendance(status);

-- Payroll
CREATE INDEX idx_payroll_runs_branch_id ON payroll_runs(branch_id);
CREATE INDEX idx_payroll_runs_status ON payroll_runs(status);
CREATE INDEX idx_payroll_runs_period ON payroll_runs(period_from, period_to);
CREATE INDEX idx_payroll_items_run_id ON payroll_items(payroll_run_id);
CREATE INDEX idx_payroll_items_employee_id ON payroll_items(employee_id);

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

CREATE TRIGGER trg_employees_updated_at
    BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_office_timings_updated_at
    BEFORE UPDATE ON office_timings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_attendance_updated_at
    BEFORE UPDATE ON attendance
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_payroll_runs_updated_at
    BEFORE UPDATE ON payroll_runs
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
    '123 Main Street',
    '+1234567890',
    'main@oleron.com'
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
    'admin@oleron.com',
    '$2a$10$Ht5bbEwwJ3ExRR8o.ygn1.PMdG.JwvsQyJt.jkDrTzBO3ALAYRsbK', -- admin
    'super_admin',
    'active'
);

-- Seed default role permissions
INSERT INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete) VALUES
('super_admin',  'employee',    TRUE, TRUE,  TRUE,  TRUE),
('super_admin',  'attendance',  TRUE, TRUE,  TRUE,  TRUE),
('super_admin',  'payroll',     TRUE, TRUE,  TRUE,  TRUE),
('super_admin',  'work_order',  TRUE, TRUE,  TRUE,  TRUE),
('super_admin',  'report',      TRUE, TRUE,  TRUE,  TRUE),
('super_admin',  'settings',    TRUE, TRUE,  TRUE,  TRUE),

('regional_manager', 'employee',    TRUE, TRUE,  TRUE,  TRUE),
('regional_manager', 'attendance',  TRUE, TRUE,  TRUE,  TRUE),
('regional_manager', 'payroll',     TRUE, TRUE,  FALSE, FALSE),
('regional_manager', 'work_order',  TRUE, TRUE,  TRUE,  TRUE),
('regional_manager', 'report',      TRUE, FALSE, FALSE, FALSE),
('regional_manager', 'settings',    TRUE, TRUE,  TRUE,  FALSE),

('manager',      'employee',    TRUE, FALSE, FALSE, FALSE),
('manager',      'attendance',  TRUE, TRUE,  TRUE,  FALSE),
('manager',      'payroll',     TRUE, FALSE, FALSE, FALSE),
('manager',      'work_order',  TRUE, FALSE, FALSE, FALSE),
('manager',      'report',      TRUE, FALSE, FALSE, FALSE),

('employee',     'attendance',  TRUE, FALSE, FALSE, FALSE),
('employee',     'payroll',     TRUE, FALSE, FALSE, FALSE);

-- Seed default menus (sidebar navigation)
-- Single-page menus: leaf items link directly to a page
-- Parent menus: only used when children are genuinely different features/views

-- Top level menus (parent_id = NULL)
INSERT INTO menus (id, parent_id, label, path, resource, sort_order) VALUES
(uuid_generate_v4(), NULL, 'Dashboard',  '/dashboard',  NULL,         1),
(uuid_generate_v4(), NULL, 'Employees',  NULL,          'employee',   2),
(uuid_generate_v4(), NULL, 'Attendance', '/attendance', 'attendance', 3),
(uuid_generate_v4(), NULL, 'Payroll',    '/payroll',    'payroll',    4),
(uuid_generate_v4(), NULL, 'Work Orders','/work-orders','work_order', 5),
(uuid_generate_v4(), NULL, 'Reports',    '/reports',    'report',     6),
(uuid_generate_v4(), NULL, 'Settings',   '/settings',   'settings',   7);

-- Sub menus for Employees
INSERT INTO menus (parent_id, label, path, resource, sort_order) VALUES
((SELECT id FROM menus WHERE label = 'Employees' AND parent_id IS NULL), 'Employee List',   '/employees',         'employee', 1),
((SELECT id FROM menus WHERE label = 'Employees' AND parent_id IS NULL), 'Work Schedule',   '/employees/schedule','attendance', 2);


-- ============================================
-- REGIONAL MANAGER BRANCH ASSIGNMENTS
-- Maps a regional_manager user to the branches they oversee
-- ============================================
CREATE TABLE regional_manager_branches (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    regional_manager_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    branch_id            UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    assigned_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(regional_manager_id, branch_id)
);

CREATE INDEX idx_rm_branches_manager_id ON regional_manager_branches(regional_manager_id);
CREATE INDEX idx_rm_branches_branch_id  ON regional_manager_branches(branch_id);


-- ============================================
-- WORK ORDER MODULE
-- Regional managers upload seal/signature assets and create work orders.
-- Signature/seal fields are snapshotted on each work order for stable PDFs.
-- ============================================

CREATE TABLE regional_manager_work_order_assets (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    regional_manager_id  UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    signer_name          VARCHAR(150) NOT NULL,
    signer_title         VARCHAR(100),
    signature_url        TEXT NOT NULL,
    seal_url             TEXT NOT NULL,
    uploaded_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE work_orders (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id             UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    created_by            UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    manager_asset_id      UUID REFERENCES regional_manager_work_order_assets(id) ON DELETE SET NULL,
    work_order_no         VARCHAR(30) NOT NULL,
    work_order_date       DATE NOT NULL DEFAULT CURRENT_DATE,
    company_name          VARCHAR(150) NOT NULL DEFAULT 'Oleron.Inc',
    company_address       TEXT,
    company_phone         VARCHAR(50),
    company_fax           VARCHAR(50),
    company_email         VARCHAR(150),
    company_website       VARCHAR(150),
    company_logo_url      TEXT,
    bill_to_name          VARCHAR(150) NOT NULL,
    bill_to_address       TEXT,
    bill_to_email         VARCHAR(150),
    job_details           TEXT NOT NULL,
    signer_name           VARCHAR(150),
    signer_title          VARCHAR(100),
    signature_url         TEXT,
    seal_url              TEXT,
    currency              VARCHAR(3) NOT NULL DEFAULT 'USD',
    sub_total_amount      NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (sub_total_amount >= 0),
    total_amount          NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    status                work_order_status NOT NULL DEFAULT 'draft',
    pdf_url               TEXT,
    issued_at             TIMESTAMPTZ,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(branch_id, work_order_no)
);

CREATE TABLE work_order_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    work_order_id   UUID NOT NULL REFERENCES work_orders(id) ON DELETE CASCADE,
    line_no         SMALLINT NOT NULL CHECK (line_no > 0),
    description     TEXT NOT NULL,
    amount          NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(work_order_id, line_no)
);

CREATE TABLE work_order_invoices (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    work_order_id         UUID NOT NULL REFERENCES work_orders(id) ON DELETE RESTRICT,
    branch_id             UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    created_by            UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    invoice_no            VARCHAR(40) NOT NULL,
    invoice_date          DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date              DATE,
    invoice_title         VARCHAR(100) NOT NULL DEFAULT 'EXPORT INVOICE',
    supply_note           TEXT DEFAULT 'SUPPLY MEANT FOR EXPORT ON PAYMENT OF INTEGRATED TAX/SUPPLY MEANT FOR EXPORT UNDER BOND OR LETTER OF UNDERTAKING WITHOUT PAYMENT OF INTEGRATED TAX',
    seller_name           VARCHAR(150) NOT NULL DEFAULT 'Oleron India',
    seller_address        TEXT,
    seller_email          VARCHAR(150),
    seller_phone          VARCHAR(50),
    seller_gstin          VARCHAR(30),
    seller_logo_url       TEXT,
    bill_to_name          VARCHAR(150) NOT NULL,
    bill_to_address       TEXT,
    bill_to_email         VARCHAR(150),
    bill_to_phone         VARCHAR(50),
    bill_to_website       VARCHAR(150),
    currency              VARCHAR(3) NOT NULL DEFAULT 'USD',
    gross_amount          NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (gross_amount >= 0),
    tax_amount            NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (tax_amount >= 0),
    additional_amount     NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (additional_amount >= 0),
    total_amount          NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    paid_amount           NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (paid_amount >= 0),
    balance_amount        NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (balance_amount >= 0),
    status                invoice_status NOT NULL DEFAULT 'draft',
    payment_status        invoice_payment_status NOT NULL DEFAULT 'unpaid',
    lut_order_number      VARCHAR(100),
    arn_number            VARCHAR(100),
    notes                 TEXT,
    signer_name           VARCHAR(150),
    signer_title          VARCHAR(100),
    signature_url         TEXT,
    seal_url              TEXT,
    pdf_url               TEXT,
    issued_at             TIMESTAMPTZ,
    cancelled_at          TIMESTAMPTZ,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(branch_id, invoice_no)
);

CREATE TABLE work_order_invoice_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id      UUID NOT NULL REFERENCES work_order_invoices(id) ON DELETE CASCADE,
    work_order_item_id UUID REFERENCES work_order_items(id) ON DELETE SET NULL,
    line_no         SMALLINT NOT NULL CHECK (line_no > 0),
    description     TEXT NOT NULL,
    quantity        NUMERIC(10,2) NOT NULL DEFAULT 1 CHECK (quantity > 0),
    sac_code        VARCHAR(20),
    unit_amount     NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (unit_amount >= 0),
    tax_amount      NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (tax_amount >= 0),
    total_amount    NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(invoice_id, line_no)
);

CREATE TABLE work_order_invoice_payments (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id          UUID NOT NULL REFERENCES work_order_invoices(id) ON DELETE CASCADE,
    recorded_by         UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    payment_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    amount              NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    method              invoice_payment_method NOT NULL,
    other_method        VARCHAR(100),
    reference_no        VARCHAR(150),
    payer_name          VARCHAR(150),
    payer_account_last4 VARCHAR(10),
    bank_name           VARCHAR(150),
    status              invoice_payment_record_status NOT NULL DEFAULT 'pending',
    notes               TEXT,
    verified_by         UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    CHECK (method <> 'other' OR NULLIF(BTRIM(other_method), '') IS NOT NULL)
);

CREATE TABLE work_order_invoice_payment_statements (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id         UUID NOT NULL REFERENCES work_order_invoice_payments(id) ON DELETE CASCADE,
    uploaded_by        UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    statement_url      TEXT NOT NULL,
    original_filename  VARCHAR(255),
    file_mime_type     VARCHAR(100),
    file_size_bytes    BIGINT CHECK (file_size_bytes IS NULL OR file_size_bytes >= 0),
    notes              TEXT,
    created_at         TIMESTAMPTZ DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION refresh_work_order_invoice_payment_summary()
RETURNS TRIGGER AS $$
DECLARE
    target_invoice_id UUID;
    confirmed_total NUMERIC(12,2);
BEGIN
    target_invoice_id := COALESCE(NEW.invoice_id, OLD.invoice_id);

    SELECT COALESCE(SUM(amount), 0)
    INTO confirmed_total
    FROM work_order_invoice_payments
    WHERE invoice_id = target_invoice_id
      AND status = 'confirmed';

    UPDATE work_order_invoices
    SET paid_amount = confirmed_total,
        balance_amount = GREATEST(total_amount - confirmed_total, 0),
        payment_status = CASE
            WHEN confirmed_total = 0 THEN 'unpaid'::invoice_payment_status
            WHEN confirmed_total < total_amount THEN 'partially_paid'::invoice_payment_status
            WHEN confirmed_total = total_amount THEN 'paid'::invoice_payment_status
            ELSE 'overpaid'::invoice_payment_status
        END,
        updated_at = NOW()
    WHERE id = target_invoice_id;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION set_work_order_invoice_payment_summary()
RETURNS TRIGGER AS $$
DECLARE
    confirmed_total NUMERIC(12,2);
BEGIN
    SELECT COALESCE(SUM(amount), 0)
    INTO confirmed_total
    FROM work_order_invoice_payments
    WHERE invoice_id = NEW.id
      AND status = 'confirmed';

    NEW.paid_amount := confirmed_total;
    NEW.balance_amount := GREATEST(NEW.total_amount - confirmed_total, 0);
    NEW.payment_status := CASE
        WHEN confirmed_total = 0 THEN 'unpaid'::invoice_payment_status
        WHEN confirmed_total < NEW.total_amount THEN 'partially_paid'::invoice_payment_status
        WHEN confirmed_total = NEW.total_amount THEN 'paid'::invoice_payment_status
        ELSE 'overpaid'::invoice_payment_status
    END;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE INDEX idx_rm_work_order_assets_manager_id ON regional_manager_work_order_assets(regional_manager_id);
CREATE INDEX idx_work_orders_branch_id ON work_orders(branch_id);
CREATE INDEX idx_work_orders_created_by ON work_orders(created_by);
CREATE INDEX idx_work_orders_status ON work_orders(status);
CREATE INDEX idx_work_orders_date ON work_orders(work_order_date);
CREATE INDEX idx_work_order_items_work_order_id ON work_order_items(work_order_id);
CREATE INDEX idx_work_order_invoices_work_order_id ON work_order_invoices(work_order_id);
CREATE INDEX idx_work_order_invoices_branch_id ON work_order_invoices(branch_id);
CREATE INDEX idx_work_order_invoices_status ON work_order_invoices(status);
CREATE INDEX idx_work_order_invoices_payment_status ON work_order_invoices(payment_status);
CREATE INDEX idx_work_order_invoices_invoice_date ON work_order_invoices(invoice_date);
CREATE INDEX idx_work_order_invoice_items_invoice_id ON work_order_invoice_items(invoice_id);
CREATE INDEX idx_work_order_invoice_payments_invoice_id ON work_order_invoice_payments(invoice_id);
CREATE INDEX idx_work_order_invoice_payments_status ON work_order_invoice_payments(status);
CREATE INDEX idx_work_order_invoice_payments_method ON work_order_invoice_payments(method);
CREATE INDEX idx_work_order_invoice_payment_statements_payment_id ON work_order_invoice_payment_statements(payment_id);

CREATE TRIGGER trg_rm_work_order_assets_updated_at
    BEFORE UPDATE ON regional_manager_work_order_assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_work_orders_updated_at
    BEFORE UPDATE ON work_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_work_order_items_updated_at
    BEFORE UPDATE ON work_order_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_work_order_invoices_updated_at
    BEFORE UPDATE ON work_order_invoices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_work_order_invoices_payment_summary
    BEFORE INSERT OR UPDATE OF total_amount ON work_order_invoices
    FOR EACH ROW EXECUTE FUNCTION set_work_order_invoice_payment_summary();

CREATE TRIGGER trg_work_order_invoice_items_updated_at
    BEFORE UPDATE ON work_order_invoice_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_work_order_invoice_payments_updated_at
    BEFORE UPDATE ON work_order_invoice_payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_work_order_invoice_payments_refresh_invoice
    AFTER INSERT OR UPDATE OR DELETE ON work_order_invoice_payments
    FOR EACH ROW EXECUTE FUNCTION refresh_work_order_invoice_payment_summary();


-- ============================================
-- RECRUITMENT MODULE
-- vacancies    → open job positions per branch
-- applications → candidates who applied to a vacancy
-- interviews   → interview sessions for shortlisted candidates
-- ============================================

CREATE TYPE vacancy_status AS ENUM (
    'draft',      -- Not yet published
    'open',       -- Accepting applications
    'closed',     -- Position filled or stopped
    'cancelled'   -- Cancelled before filling
);

CREATE TYPE application_status AS ENUM (
    'applied',              -- Just submitted
    'shortlisted',          -- CV reviewed and shortlisted
    'rejected',             -- Not proceeding
    'interview_scheduled',  -- Interview booked
    'hired',                -- Converted to employee
    'withdrawn'             -- Candidate withdrew
);

CREATE TYPE interview_type AS ENUM (
    'phone',
    'video',
    'in_person'
);

CREATE TYPE interview_outcome AS ENUM (
    'pending',   -- Not yet held
    'passed',    -- Candidate progressed
    'failed',    -- Did not pass
    'no_show'    -- Candidate did not appear
);

-- Job postings per branch
CREATE TABLE vacancies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id       UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title           VARCHAR(150) NOT NULL,                -- e.g. "Senior Developer"
    department      VARCHAR(100),                         -- e.g. "Engineering"
    description     TEXT,                                 -- Full job description / JD
    requirements    TEXT,                                 -- Skills, qualifications
    positions       SMALLINT NOT NULL DEFAULT 1,          -- Number of openings
    status          vacancy_status NOT NULL DEFAULT 'draft',
    deadline        DATE,                                 -- Application deadline (optional)
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Candidate applications
CREATE TABLE applications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vacancy_id      UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    email           VARCHAR(150) NOT NULL,
    phone           VARCHAR(20),
    cv_url          TEXT,                                 -- Link to uploaded CV/resume
    cover_letter    TEXT,
    status          application_status NOT NULL DEFAULT 'applied',
    notes           TEXT,                                 -- Reviewer notes / sorting comments
    applied_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Interview sessions for shortlisted candidates
CREATE TABLE interviews (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id  UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    interviewer_id  UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,  -- branch user conducting interview
    scheduled_at    TIMESTAMPTZ NOT NULL,
    type            interview_type NOT NULL DEFAULT 'in_person',
    location        VARCHAR(200),    -- room name, address, or video link
    outcome         interview_outcome NOT NULL DEFAULT 'pending',
    feedback        TEXT,            -- interviewer feedback after the session
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);


-- Recruitment indexes
CREATE INDEX idx_vacancies_branch_id ON vacancies(branch_id);
CREATE INDEX idx_vacancies_status ON vacancies(status);
CREATE INDEX idx_applications_vacancy_id ON applications(vacancy_id);
CREATE INDEX idx_applications_status ON applications(status);
CREATE INDEX idx_applications_email ON applications(email);
CREATE INDEX idx_interviews_application_id ON interviews(application_id);
CREATE INDEX idx_interviews_interviewer_id ON interviews(interviewer_id);
CREATE INDEX idx_interviews_scheduled_at ON interviews(scheduled_at);

-- updated_at triggers for recruitment
CREATE TRIGGER trg_vacancies_updated_at
    BEFORE UPDATE ON vacancies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_applications_updated_at
    BEFORE UPDATE ON applications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_interviews_updated_at
    BEFORE UPDATE ON interviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();


-- ============================================
-- SCHEMA OVERVIEW
-- ============================================
--
-- branches          → Each organization branch (identified by code e.g. BRANCH01, office_timing_id → active schedule)
--     │
--     ├── branch_calendar → Per-date holiday/working-day overrides for payroll
--     └── users           → Staff accounts linked to a branch
--             │
--             ├── tokens      → Reset password / email verify tokens
--             ├── sessions    → Active login sessions (per device)
--             └── login_audit → All login attempts logged
--
-- role_permissions  → What each role (super_admin/admin/manager/employee) can do per resource
-- menus             → Sidebar navigation with tree structure (filtered by role_permissions)
--
-- regional_manager_work_order_assets → Uploaded manager signature/seal used for work orders
-- work_orders       → Work order header, bill-to, job, totals, signer snapshot, PDF URL
--     └── work_order_items → Work order line items (description, amount)
--     └── work_order_invoices → Export invoices issued against a work order, with payment status
--             ├── work_order_invoice_items → Invoice line items (description, qty, SAC, tax, total)
--             └── work_order_invoice_payments → Payments by any supported method
--                     └── work_order_invoice_payment_statements → Uploaded payment statements/proofs
--
-- vacancies         → Job postings per branch (title, JD, openings, status)
--     └── applications → Candidates who applied (CV, status: applied → shortlisted → hired/rejected)
--             └── interviews → Scheduled interview sessions (type, outcome, feedback)
--
-- Hire flow: vacancy(open) → application(applied) → (shortlisted) → interview(scheduled) → application(hired) → create user+employee



-- Consultant monthly timesheet submissions

CREATE TABLE IF NOT EXISTS consultant_timesheets (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id    UUID         NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    year           INT          NOT NULL CHECK (year >= 2000),
    month          INT          NOT NULL CHECK (month BETWEEN 1 AND 12),
    support_hours  NUMERIC(8,2) NOT NULL DEFAULT 0,
    overtime_hours NUMERIC(8,2) NOT NULL DEFAULT 0,
    notes          TEXT,
    details        JSONB,

    status         TEXT         NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewer_id    UUID         REFERENCES users(id),
    review_note    TEXT,
    reviewed_at    TIMESTAMPTZ,

    submitted_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (employee_id, year, month)
);

CREATE INDEX IF NOT EXISTS idx_consultant_timesheets_employee ON consultant_timesheets (employee_id);
CREATE INDEX IF NOT EXISTS idx_consultant_timesheets_status   ON consultant_timesheets (status);

CREATE TRIGGER update_consultant_timesheets_updated_at
    BEFORE UPDATE ON consultant_timesheets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();



CREATE TABLE contact_submissions (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          TEXT        NOT NULL,
  company       TEXT,
  email         TEXT        NOT NULL,
  phone         TEXT,
  category      TEXT,
  message       TEXT        NOT NULL,
  status        TEXT        NOT NULL DEFAULT 'new'
                  CHECK (status IN ('new', 'read', 'replied', 'archived')),
  ip_address    INET,
  user_agent    TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_contact_submissions_status_created
  ON contact_submissions (status, created_at DESC);

CREATE INDEX idx_contact_submissions_email
  ON contact_submissions (email);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_contact_submissions_updated_at ON contact_submissions;

CREATE TRIGGER trg_contact_submissions_updated_at
  BEFORE UPDATE ON contact_submissions
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
