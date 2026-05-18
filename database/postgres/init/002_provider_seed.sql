INSERT INTO dopa.citizens (
    national_id, title_th, first_name_th, last_name_th,
    title_en, first_name_en, last_name_en,
    date_of_birth, gender, address_line, province, district, subdistrict, postal_code, person_status
) VALUES
    ('1101700203451', 'นาย', 'สมชาย', 'ใจดี', 'Mr.', 'Somchai', 'Jaidee', '1991-05-20', 'M', '99/1 ถนนสุขุมวิท', 'กรุงเทพมหานคร', 'คลองเตย', 'คลองเตย', '10110', 'alive'),
    ('1101700203452', 'นางสาว', 'มาลี', 'รักไทย', 'Ms.', 'Malee', 'Rakthai', '2010-09-15', 'F', '12 หมู่ 4', 'เชียงใหม่', 'เมืองเชียงใหม่', 'สุเทพ', '50200', 'alive'),
    ('1101700203453', 'นาย', 'ประเสริฐ', 'มีสุข', 'Mr.', 'Prasert', 'Meesuk', '1980-01-10', 'M', '45 ถนนมิตรภาพ', 'นครราชสีมา', 'เมืองนครราชสีมา', 'ในเมือง', '30000', 'alive'),
    ('1101700203454', 'นาง', 'อรุณี', 'แสงทอง', 'Mrs.', 'Arunee', 'Saengthong', '1975-12-01', 'F', '7/8 ถนนราชดำเนิน', 'ขอนแก่น', 'เมืองขอนแก่น', 'ในเมือง', '40000', 'deceased')
ON CONFLICT (national_id) DO NOTHING;

INSERT INTO dopa.id_cards (
    national_id, laser_code, issued_at, expired_at, card_status, revoked_reason
) VALUES
    ('1101700203451', 'JT1234567890', '2022-01-10', '2030-01-09', 'active', NULL),
    ('1101700203452', 'JT1234567891', '2020-06-01', '2028-05-31', 'active', NULL),
    ('1101700203453', 'JT1234567892', '2016-03-15', '2024-03-14', 'expired', NULL),
    ('1101700203454', 'JT1234567893', '2021-05-20', '2029-05-19', 'revoked', 'citizen deceased')
ON CONFLICT DO NOTHING;

INSERT INTO sso.insured_persons (
    national_id, insured_status, section, registered_at, employer_id, employer_name,
    contribution_months, latest_contribution_month
) VALUES
    ('1101700203451', 'insured', '40', '2024-01-01', NULL, NULL, 12, '2026-04-01'),
    ('1101700203452', 'uninsured', NULL, NULL, NULL, NULL, 0, NULL),
    ('1101700203453', 'insured', '33', '2019-02-01', 'EMP-001', 'Siam Manufacturing Co., Ltd.', 72, '2026-04-01'),
    ('1101700203454', 'terminated', '39', '2020-01-01', NULL, NULL, 24, '2025-12-01')
ON CONFLICT (national_id) DO NOTHING;

INSERT INTO sso.contributions (
    national_id, contribution_month, employee_amount, employer_amount, government_amount, paid_at, payment_status
) VALUES
    ('1101700203451', '2026-04-01', 100.00, 0.00, 50.00, '2026-04-15', 'paid'),
    ('1101700203451', '2026-03-01', 100.00, 0.00, 50.00, '2026-03-15', 'paid'),
    ('1101700203451', '2026-02-01', 100.00, 0.00, 50.00, '2026-02-15', 'paid'),
    ('1101700203453', '2026-04-01', 750.00, 750.00, 0.00, '2026-04-15', 'paid'),
    ('1101700203453', '2026-03-01', 750.00, 750.00, 0.00, '2026-03-15', 'paid')
ON CONFLICT (national_id, contribution_month) DO NOTHING;

INSERT INTO ktb.bank_accounts (
    national_id, bank_code, branch_code, account_no, account_name,
    account_type, account_status, balance, average_monthly_income, opened_at
) VALUES
    ('1101700203451', '006', '0001', '006123456789', 'Somchai Jaidee', 'savings', 'active', 12000.00, 15000.00, '2020-01-05'),
    ('1101700203452', '006', '0002', '006123456790', 'Malee Rakthai', 'savings', 'active', 5000.00, 0.00, '2023-08-01'),
    ('1101700203453', '006', '0003', '006123456791', 'Prasert Meesuk', 'savings', 'active', 250000.00, 45000.00, '2018-02-10'),
    ('1101700203454', '006', '0004', '006123456792', 'Arunee Saengthong', 'savings', 'frozen', 8000.00, 9000.00, '2017-11-20')
ON CONFLICT (bank_code, account_no) DO NOTHING;

INSERT INTO ktb.promptpay_registrations (
    national_id, account_id, proxy_type, proxy_value, registration_status, registered_at
)
SELECT '1101700203451', account_id, 'national_id', '1101700203451', 'active', '2020-01-10'
FROM ktb.bank_accounts WHERE national_id = '1101700203451'
ON CONFLICT (proxy_type, proxy_value) DO NOTHING;

INSERT INTO ktb.promptpay_registrations (
    national_id, account_id, proxy_type, proxy_value, registration_status, registered_at
)
SELECT '1101700203452', account_id, 'mobile', '0812345678', 'active', '2023-08-05'
FROM ktb.bank_accounts WHERE national_id = '1101700203452'
ON CONFLICT (proxy_type, proxy_value) DO NOTHING;

INSERT INTO ktb.promptpay_registrations (
    national_id, account_id, proxy_type, proxy_value, registration_status, registered_at
)
SELECT '1101700203454', account_id, 'national_id', '1101700203454', 'revoked', '2020-02-20'
FROM ktb.bank_accounts WHERE national_id = '1101700203454'
ON CONFLICT (proxy_type, proxy_value) DO NOTHING;
