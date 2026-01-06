-- Пароль по умолчанию: admin123
-- Email: admin@test.com
DO $$
DECLARE
    admin_exists BOOLEAN;
BEGIN
    SELECT EXISTS(SELECT 1 FROM users WHERE is_admin = true) INTO admin_exists;
    
    IF NOT admin_exists THEN
        INSERT INTO users (email, first_name, last_name, password_hash, is_landlord, is_admin, created_at, updated_at)
        VALUES (
            'admin@rentapp.com',
            'Admin',
            'User',
            '$2a$12$lirb43ksbqPL/mLOSGmCMOnrJQWtDeeWl.frct8xnueZd5dX1suwW', -- bcrypt hash для 'admin123'
            false,
            true,
            NOW(),
            NOW()
        );
    END IF;
END $$;

