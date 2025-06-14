-- Insertar usuario administrador (Rector)
INSERT INTO users (id, email, name, password, role, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'rector@colegio.edu',
    'Rector del Colegio',
    crypt('rector123', gen_salt('bf')),
    'administrador',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) ON CONFLICT (email) DO NOTHING; 