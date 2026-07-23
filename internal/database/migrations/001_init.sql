CREATE TABLE IF NOT EXISTS entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK(type IN ('person', 'organization')),
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'inactive')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role_id INTEGER NOT NULL DEFAULT 1,
    entity_id INTEGER,
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (entity_id) REFERENCES entities(id)
);

CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id INTEGER NOT NULL,
    module TEXT NOT NULL,
    action TEXT NOT NULL CHECK(action IN ('read', 'write', 'delete')),
    UNIQUE(role_id, module, action),
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL,
    original_name TEXT NOT NULL,
    path TEXT NOT NULL,
    size INTEGER NOT NULL DEFAULT 0,
    mime_type TEXT NOT NULL DEFAULT '',
    uploaded_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (uploaded_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS file_associations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id)
);

CREATE TABLE IF NOT EXISTS person_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL,
    effective_date DATE NOT NULL,
    event_type TEXT NOT NULL CHECK(event_type IN ('onboard', 'position_change', 'info_change', 'offboard')),
    payload TEXT NOT NULL,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (person_id) REFERENCES entities(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS org_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    org_id INTEGER NOT NULL,
    effective_date DATE NOT NULL,
    event_type TEXT NOT NULL CHECK(event_type IN ('establish', 'info_change', 'dissolve')),
    payload TEXT NOT NULL,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES entities(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS attendance_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL,
    date DATE NOT NULL,
    event_type TEXT NOT NULL CHECK(event_type IN (
        'normal_attendance', 'supplementary_attendance',
        'compensatory_leave', 'personal_leave', 'sick_leave',
        'annual_leave', 'statutory_leave', 'welfare_leave',
        'workday_overtime', 'holiday_overtime',
        'missing_clock', 'late', 'early_leave',
        'annual_leave_allot', 'annual_leave_carryover'
    )),
    duration REAL NOT NULL DEFAULT 1.0,
    remark TEXT DEFAULT '',
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (person_id) REFERENCES entities(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS salary_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL,
    period TEXT NOT NULL,
    event_type TEXT NOT NULL CHECK(event_type IN ('performance', 'reward_punish', 'loan_deduct', 'tax_deduct', 'other')),
    amount REAL NOT NULL DEFAULT 0,
    detail TEXT DEFAULT '',
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (person_id) REFERENCES entities(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS person_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL,
    effective_date DATE NOT NULL,
    snapshot_data TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (person_id) REFERENCES entities(id),
    FOREIGN KEY (event_id) REFERENCES person_events(id)
);

CREATE TABLE IF NOT EXISTS org_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    org_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL,
    effective_date DATE NOT NULL,
    snapshot_data TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES entities(id),
    FOREIGN KEY (event_id) REFERENCES org_events(id)
);

CREATE TABLE IF NOT EXISTS attendance_summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL,
    period TEXT NOT NULL,
    normal_attendance_days REAL NOT NULL DEFAULT 0,
    supplementary_attendance_days REAL NOT NULL DEFAULT 0,
    compensatory_leave_days REAL NOT NULL DEFAULT 0,
    personal_leave_days REAL NOT NULL DEFAULT 0,
    sick_leave_days REAL NOT NULL DEFAULT 0,
    annual_leave_days REAL NOT NULL DEFAULT 0,
    statutory_leave_days REAL NOT NULL DEFAULT 0,
    welfare_leave_days REAL NOT NULL DEFAULT 0,
    workday_overtime_days REAL NOT NULL DEFAULT 0,
    holiday_overtime_days REAL NOT NULL DEFAULT 0,
    missing_clock_count INTEGER NOT NULL DEFAULT 0,
    late_count INTEGER NOT NULL DEFAULT 0,
    early_leave_count INTEGER NOT NULL DEFAULT 0,
    annual_leave_allot REAL NOT NULL DEFAULT 0,
    annual_leave_carryover REAL NOT NULL DEFAULT 0,
    violation_count INTEGER NOT NULL DEFAULT 0,
    calculated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(person_id, period),
    FOREIGN KEY (person_id) REFERENCES entities(id)
);

CREATE TABLE IF NOT EXISTS salary_summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL,
    period TEXT NOT NULL,
    attendance_salary REAL NOT NULL DEFAULT 0,
    full_attendance_bonus REAL NOT NULL DEFAULT 0,
    overtime_salary REAL NOT NULL DEFAULT 0,
    performance_salary REAL NOT NULL DEFAULT 0,
    position_allowance REAL NOT NULL DEFAULT 0,
    meal_subsidy REAL NOT NULL DEFAULT 0,
    housing_subsidy REAL NOT NULL DEFAULT 0,
    transport_subsidy REAL NOT NULL DEFAULT 0,
    heat_subsidy REAL NOT NULL DEFAULT 0,
    insurance_compensation REAL NOT NULL DEFAULT 0,
    housing_fund_compensation REAL NOT NULL DEFAULT 0,
    social_insurance_deduct REAL NOT NULL DEFAULT 0,
    housing_fund_deduct REAL NOT NULL DEFAULT 0,
    tax_deduct REAL NOT NULL DEFAULT 0,
    loan_deduct REAL NOT NULL DEFAULT 0,
    reward_punish REAL NOT NULL DEFAULT 0,
    other_adjustments REAL NOT NULL DEFAULT 0,
    total_salary REAL NOT NULL DEFAULT 0,
    detail_data TEXT NOT NULL DEFAULT '{}',
    calculated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(person_id, period),
    FOREIGN KEY (person_id) REFERENCES entities(id)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id INTEGER,
    payload TEXT NOT NULL DEFAULT '{}',
    ip_address TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_person_events_person ON person_events(person_id, effective_date);
CREATE INDEX IF NOT EXISTS idx_org_events_org ON org_events(org_id, effective_date);
CREATE INDEX IF NOT EXISTS idx_attendance_events_person_date ON attendance_events(person_id, date);
CREATE INDEX IF NOT EXISTS idx_salary_events_person_period ON salary_events(person_id, period);
CREATE INDEX IF NOT EXISTS idx_person_snapshots_person_date ON person_snapshots(person_id, effective_date);
CREATE INDEX IF NOT EXISTS idx_file_associations_target ON file_associations(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at);
