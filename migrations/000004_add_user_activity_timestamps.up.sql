ALTER TABLE users
    ADD COLUMN last_login_at DATETIME(6) NULL AFTER status,
    ADD COLUMN last_activity_at DATETIME(6) NULL AFTER last_login_at;
