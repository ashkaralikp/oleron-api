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
    'admin',          -- Branch level full access (Manager)
    'manager',        -- Manages employees within a branch
    'employee'        -- Punches in/out; views own attendance and salary
    'consultant'       -- External consultant with limited access (e.g. payroll only)
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
    hourly_rate      NUMERIC(10,2),
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
('super_admin',  'report',      TRUE, TRUE,  TRUE,  TRUE),
('super_admin',  'settings',    TRUE, TRUE,  TRUE,  TRUE),

('admin',        'employee',    TRUE, TRUE,  TRUE,  TRUE),
('admin',        'attendance',  TRUE, TRUE,  TRUE,  TRUE),
('admin',        'payroll',     TRUE, TRUE,  FALSE, FALSE),
('admin',        'report',      TRUE, FALSE, FALSE, FALSE),
('admin',        'settings',    TRUE, TRUE,  TRUE,  FALSE),

('manager',      'employee',    TRUE, FALSE, FALSE, FALSE),
('manager',      'attendance',  TRUE, TRUE,  TRUE,  FALSE),
('manager',      'payroll',     TRUE, FALSE, FALSE, FALSE),
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
(uuid_generate_v4(), NULL, 'Reports',    '/reports',    'report',     5),
(uuid_generate_v4(), NULL, 'Settings',   '/settings',   'settings',   6);

-- Sub menus for Employees
INSERT INTO menus (parent_id, label, path, resource, sort_order) VALUES
((SELECT id FROM menus WHERE label = 'Employees' AND parent_id IS NULL), 'Employee List',   '/employees',         'employee', 1),
((SELECT id FROM menus WHERE label = 'Employees' AND parent_id IS NULL), 'Work Schedule',   '/employees/schedule','attendance', 2);


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
-- vacancies         → Job postings per branch (title, JD, openings, status)
--     └── applications → Candidates who applied (CV, status: applied → shortlisted → hired/rejected)
--             └── interviews → Scheduled interview sessions (type, outcome, feedback)
--
-- Hire flow: vacancy(open) → application(applied) → (shortlisted) → interview(scheduled) → application(hired) → create user+employee