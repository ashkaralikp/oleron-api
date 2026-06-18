-- ============================================
-- MIGRATION: Add 'accounts' user role
-- Run this against the live database once.
-- ============================================

-- 1. Extend the enum (must be committed before the new value can be used)
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'accounts';

-- Commit the enum change so the new value is visible to the next statement.
COMMIT;

-- 2. Role permissions — accounts can view/manage work orders and invoices,
--    view payroll, and view reports. No employee, attendance, or settings access.
INSERT INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete) VALUES
('accounts', 'work_order', TRUE, FALSE, TRUE,  FALSE),
('accounts', 'payroll',    TRUE, FALSE, FALSE, FALSE),
('accounts', 'report',     TRUE, FALSE, FALSE, FALSE)
ON CONFLICT (role, resource) DO NOTHING;
