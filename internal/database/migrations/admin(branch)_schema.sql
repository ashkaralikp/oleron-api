-- ============================================
-- FEE TYPES MASTER  (services & procedures)
-- ============================================
CREATE TABLE fee_types (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code                VARCHAR(50)     NOT NULL UNIQUE,   -- e.g. 'REGISTRATION', 'DRESSING_S'
    name                VARCHAR(150)    NOT NULL,          -- display label on bill
    category            VARCHAR(50)     NOT NULL,          -- 'registration' | 'consultation' |
                                                           -- 'procedure' | 'dressing' | 'other'
    default_rate        NUMERIC(10,2)   NOT NULL DEFAULT 0,
    gst_percent         NUMERIC(5,2)    NOT NULL DEFAULT 0, -- e.g. 5.00, 12.00, 18.00
    is_qty_applicable   BOOLEAN         DEFAULT FALSE,     -- TRUE for dressings, procedures
    is_taxable          BOOLEAN         DEFAULT FALSE,
    is_active           BOOLEAN         DEFAULT TRUE,
    sort_order          INT             DEFAULT 0,
    created_at          TIMESTAMPTZ     DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     DEFAULT NOW()
);

-- ============================================
-- BRANCH FEE OVERRIDES
-- Branch admin can override rate AND gst_percent
-- ============================================
CREATE TABLE branch_fee_overrides (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id       UUID            NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    fee_type_id     UUID            NOT NULL REFERENCES fee_types(id) ON DELETE CASCADE,
    rate            NUMERIC(10,2)   NOT NULL,
    gst_percent     NUMERIC(5,2)    NOT NULL DEFAULT 0,
    is_active       BOOLEAN         DEFAULT TRUE,
    updated_by      UUID            REFERENCES users(id) ON DELETE SET NULL,
    updated_at      TIMESTAMPTZ     DEFAULT NOW(),
    UNIQUE(branch_id, fee_type_id)
);

-- ============================================
-- LAB TESTS MASTER
-- ============================================
CREATE TABLE lab_tests (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code                VARCHAR(50)     NOT NULL UNIQUE,   -- 'CBC', 'LFT', 'HBA1C'
    name                VARCHAR(150)    NOT NULL,
    sample_type         VARCHAR(50),                       -- 'blood' | 'urine' | 'stool' | 'swab'
    default_rate        NUMERIC(10,2)   NOT NULL DEFAULT 0,
    gst_percent         NUMERIC(5,2)    NOT NULL DEFAULT 0,
    outsource_cost      NUMERIC(10,2)   DEFAULT 0,         -- cost if sent to external lab
    turnaround_hours    INT             DEFAULT 24,
    urgent_surcharge    NUMERIC(10,2)   DEFAULT 0,
    is_panel            BOOLEAN         DEFAULT FALSE,
    is_outsourced       BOOLEAN         DEFAULT FALSE,
    is_taxable          BOOLEAN         DEFAULT FALSE,
    is_active           BOOLEAN         DEFAULT TRUE,
    sort_order          INT             DEFAULT 0,
    created_at          TIMESTAMPTZ     DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     DEFAULT NOW()
);

-- ============================================
-- LAB TEST COMPONENTS  (panel → sub-tests)
-- ============================================
CREATE TABLE lab_test_components (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    panel_id        UUID NOT NULL REFERENCES lab_tests(id) ON DELETE CASCADE,
    component_id    UUID NOT NULL REFERENCES lab_tests(id) ON DELETE RESTRICT,
    sort_order      INT  DEFAULT 0,
    UNIQUE(panel_id, component_id)
);

-- ============================================
-- BRANCH LAB PRICE OVERRIDES
-- ============================================
CREATE TABLE branch_lab_price_overrides (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id       UUID            NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    lab_test_id     UUID            NOT NULL REFERENCES lab_tests(id) ON DELETE CASCADE,
    rate            NUMERIC(10,2)   NOT NULL,
    gst_percent     NUMERIC(5,2)    NOT NULL DEFAULT 0,
    is_active       BOOLEAN         DEFAULT TRUE,
    updated_by      UUID            REFERENCES users(id) ON DELETE SET NULL,
    updated_at      TIMESTAMPTZ     DEFAULT NOW(),
    UNIQUE(branch_id, lab_test_id)
);

-- ============================================
-- INDEXES
-- ============================================
CREATE INDEX idx_fee_types_code       ON fee_types(code);
CREATE INDEX idx_fee_types_category   ON fee_types(category);
CREATE INDEX idx_lab_tests_code       ON lab_tests(code);
CREATE INDEX idx_lab_test_comp_panel  ON lab_test_components(panel_id);
CREATE INDEX idx_bfo_branch           ON branch_fee_overrides(branch_id);
CREATE INDEX idx_blpo_branch          ON branch_lab_price_overrides(branch_id);

-- ============================================
-- TRIGGERS
-- ============================================
CREATE TRIGGER trg_fee_types_updated_at
    BEFORE UPDATE ON fee_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_lab_tests_updated_at
    BEFORE UPDATE ON lab_tests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================
-- SEED — FEE TYPES  (matching your screenshot)
-- ============================================
INSERT INTO fee_types (code, name, category, default_rate, gst_percent, is_qty_applicable, is_taxable) VALUES
('REGISTRATION',          'Registration',               'registration',  100,  0,     FALSE, FALSE),
('CONSULTATION_GEN',      'Consultation (general)',     'consultation',  300,  0,     FALSE, FALSE),
('CONSULTATION_EYE',      'Consultation (eye)',         'consultation',  350,  0,     FALSE, FALSE),
('FOLLOW_UP',             'Follow-up consultation',     'consultation',  150,  0,     FALSE, FALSE),
('DRESSING_S',            'Dressing (S)',               'dressing',       15,  0,     TRUE,  FALSE),
('DRESSING_M',            'Dressing (M)',               'dressing',       16,  0,     TRUE,  FALSE),
('DRESSING_L',            'Dressing (L)',               'dressing',       17,  0,     TRUE,  FALSE),
('DRESSING_CHARGE',       'Dressing charge',            'dressing',       15,  0,     TRUE,  FALSE),
('BLADDER_WASH',          'Bladder wash',               'procedure',      44,  0,     TRUE,  FALSE),
('FOREIGN_BODY_REMOVAL',  'Foreign body removal',       'procedure',      30,  0,     FALSE, FALSE),
('HOME_CARE_DR',          'Home care Dr charge',        'procedure',      31,  0,     FALSE, FALSE),
('HOME_CARE_CHARGE_DR',   'Home care charge Dr',        'procedure',      20,  0,     FALSE, FALSE),
('I_AND_D',               'I & D',                      'procedure',      22,  0,     FALSE, FALSE),
('PROCEDURE',             'Procedure',                  'procedure',       0,  0,     FALSE, FALSE);

-- ============================================
-- SEED — LAB TESTS
-- ============================================
INSERT INTO lab_tests (code, name, sample_type, default_rate, gst_percent, turnaround_hours, is_panel) VALUES
('BLOOD_SUGAR_F',  'Blood sugar (fasting)',    'blood',  80,  0, 4,  FALSE),
('BLOOD_SUGAR_PP', 'Blood sugar (post-meal)',  'blood',  80,  0, 4,  FALSE),
('HBA1C',          'HbA1c',                   'blood',  350, 0, 24, FALSE),
('URINE_RE',       'Urine routine exam',       'urine',  120, 0, 4,  FALSE),
('HB',             'Haemoglobin',             'blood',  0,   0, 4,  FALSE),
('WBC',            'WBC count',               'blood',  0,   0, 4,  FALSE),
('PLATELETS',      'Platelet count',          'blood',  0,   0, 4,  FALSE),
('CBC',            'Complete blood count',    'blood',  250, 0, 6,  TRUE),
('LFT',            'Liver function test',     'blood',  500, 0, 24, TRUE),
('RFT',            'Renal function test',     'blood',  450, 0, 24, TRUE),
('LIPID',          'Lipid profile',           'blood',  450, 0, 24, TRUE);