CREATE SCHEMA IF NOT EXISTS dopa;
CREATE SCHEMA IF NOT EXISTS sso;
CREATE SCHEMA IF NOT EXISTS ktb;

CREATE TABLE IF NOT EXISTS dopa.citizens (
    national_id CHAR(13) PRIMARY KEY,
    title_th VARCHAR(32) NOT NULL,
    first_name_th VARCHAR(120) NOT NULL,
    last_name_th VARCHAR(120) NOT NULL,
    title_en VARCHAR(32),
    first_name_en VARCHAR(120),
    last_name_en VARCHAR(120),
    date_of_birth DATE NOT NULL,
    gender CHAR(1) NOT NULL CHECK (gender IN ('M', 'F', 'X')),
    nationality VARCHAR(64) NOT NULL DEFAULT 'Thai',
    address_line TEXT,
    province VARCHAR(120),
    district VARCHAR(120),
    subdistrict VARCHAR(120),
    postal_code CHAR(5),
    person_status VARCHAR(24) NOT NULL DEFAULT 'alive'
        CHECK (person_status IN ('alive', 'deceased', 'missing')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT citizens_national_id_digits CHECK (national_id ~ '^[0-9]{13}$')
);

CREATE TABLE IF NOT EXISTS dopa.id_cards (
    card_id BIGSERIAL PRIMARY KEY,
    national_id CHAR(13) NOT NULL REFERENCES dopa.citizens(national_id),
    laser_code VARCHAR(16),
    issued_at DATE NOT NULL,
    expired_at DATE NOT NULL,
    card_status VARCHAR(24) NOT NULL DEFAULT 'active'
        CHECK (card_status IN ('active', 'expired', 'revoked', 'lost')),
    revoked_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT id_cards_valid_period CHECK (expired_at > issued_at)
);

CREATE INDEX IF NOT EXISTS idx_dopa_id_cards_national_id
    ON dopa.id_cards(national_id);

CREATE TABLE IF NOT EXISTS sso.insured_persons (
    national_id CHAR(13) PRIMARY KEY,
    insured_status VARCHAR(24) NOT NULL DEFAULT 'insured'
        CHECK (insured_status IN ('insured', 'uninsured', 'suspended', 'terminated')),
    section VARCHAR(2) CHECK (section IN ('33', '39', '40')),
    registered_at DATE,
    employer_id VARCHAR(32),
    employer_name VARCHAR(160),
    contribution_months INTEGER NOT NULL DEFAULT 0 CHECK (contribution_months >= 0),
    latest_contribution_month DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT insured_persons_national_id_digits CHECK (national_id ~ '^[0-9]{13}$'),
    CONSTRAINT insured_section_required CHECK (
        (insured_status = 'uninsured' AND section IS NULL)
        OR (insured_status <> 'uninsured' AND section IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS sso.contributions (
    contribution_id BIGSERIAL PRIMARY KEY,
    national_id CHAR(13) NOT NULL REFERENCES sso.insured_persons(national_id),
    contribution_month DATE NOT NULL,
    employee_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (employee_amount >= 0),
    employer_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (employer_amount >= 0),
    government_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (government_amount >= 0),
    paid_at DATE,
    payment_status VARCHAR(24) NOT NULL DEFAULT 'paid'
        CHECK (payment_status IN ('paid', 'late', 'missing', 'refunded')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (national_id, contribution_month)
);

CREATE INDEX IF NOT EXISTS idx_sso_contributions_national_id_month
    ON sso.contributions(national_id, contribution_month DESC);

CREATE TABLE IF NOT EXISTS ktb.bank_accounts (
    account_id BIGSERIAL PRIMARY KEY,
    national_id CHAR(13) NOT NULL,
    bank_code CHAR(3) NOT NULL DEFAULT '006',
    branch_code VARCHAR(8),
    account_no VARCHAR(32) NOT NULL,
    account_name VARCHAR(180) NOT NULL,
    account_type VARCHAR(24) NOT NULL DEFAULT 'savings'
        CHECK (account_type IN ('savings', 'current', 'wallet')),
    account_status VARCHAR(24) NOT NULL DEFAULT 'active'
        CHECK (account_status IN ('active', 'closed', 'frozen', 'dormant')),
    balance NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    average_monthly_income NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (average_monthly_income >= 0),
    opened_at DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (bank_code, account_no),
    CONSTRAINT bank_accounts_national_id_digits CHECK (national_id ~ '^[0-9]{13}$')
);

CREATE INDEX IF NOT EXISTS idx_ktb_bank_accounts_national_id
    ON ktb.bank_accounts(national_id);

CREATE TABLE IF NOT EXISTS ktb.promptpay_registrations (
    promptpay_id BIGSERIAL PRIMARY KEY,
    national_id CHAR(13) NOT NULL,
    account_id BIGINT NOT NULL REFERENCES ktb.bank_accounts(account_id),
    proxy_type VARCHAR(24) NOT NULL
        CHECK (proxy_type IN ('national_id', 'mobile', 'ewallet')),
    proxy_value VARCHAR(32) NOT NULL,
    registration_status VARCHAR(24) NOT NULL DEFAULT 'active'
        CHECK (registration_status IN ('active', 'inactive', 'revoked')),
    registered_at DATE NOT NULL DEFAULT CURRENT_DATE,
    revoked_at DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (proxy_type, proxy_value),
    CONSTRAINT promptpay_national_id_digits CHECK (national_id ~ '^[0-9]{13}$')
);

CREATE INDEX IF NOT EXISTS idx_ktb_promptpay_national_id
    ON ktb.promptpay_registrations(national_id);
