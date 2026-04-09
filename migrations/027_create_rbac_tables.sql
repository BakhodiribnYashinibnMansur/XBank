-- +goose Up
CREATE TABLE rbac_roles (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT        NOT NULL DEFAULT '',
    is_system   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE rbac_permissions (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    resource    VARCHAR(100) NOT NULL,
    action      VARCHAR(50)  NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (resource, action)
);

CREATE TABLE rbac_policies (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id       UUID        NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
    permission_id UUID        NOT NULL REFERENCES rbac_permissions(id) ON DELETE CASCADE,
    scope         VARCHAR(20) NOT NULL DEFAULT 'all',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (role_id, permission_id)
);

CREATE INDEX idx_rbac_policies_role_id ON rbac_policies(role_id);

-- Seed system roles
INSERT INTO rbac_roles (name, description, is_system) VALUES
    ('CUSTOMER', 'Default customer role', TRUE),
    ('TELLER',   'Bank teller role',      TRUE),
    ('ADMIN',    'Administrator role',     TRUE);

-- Seed permissions (resource:action pairs)
INSERT INTO rbac_permissions (resource, action) VALUES
    ('accounts',  'read'),  ('accounts',  'write'), ('accounts',  'delete'),
    ('transfers', 'read'),  ('transfers', 'write'),
    ('cards',     'read'),  ('cards',     'write'),
    ('users',     'read'),  ('users',     'write'), ('users',     'delete'),
    ('kyc',       'read'),  ('kyc',       'write'), ('kyc',       'approve'),
    ('admin',     'read'),  ('admin',     'write'),
    ('reports',   'read');

-- ADMIN gets everything with scope=all
INSERT INTO rbac_policies (role_id, permission_id, scope)
SELECT r.id, p.id, 'all'
FROM rbac_roles r, rbac_permissions p
WHERE r.name = 'ADMIN';

-- TELLER gets operational read/write with scope=all
INSERT INTO rbac_policies (role_id, permission_id, scope)
SELECT r.id, p.id, 'all'
FROM rbac_roles r, rbac_permissions p
WHERE r.name = 'TELLER'
  AND p.resource IN ('accounts', 'transfers', 'cards', 'kyc', 'users')
  AND p.action IN ('read', 'write');

-- CUSTOMER gets own accounts/transfers/cards read
INSERT INTO rbac_policies (role_id, permission_id, scope)
SELECT r.id, p.id, 'own'
FROM rbac_roles r, rbac_permissions p
WHERE r.name = 'CUSTOMER'
  AND p.resource IN ('accounts', 'transfers', 'cards')
  AND p.action = 'read';

-- CUSTOMER gets own accounts/transfers/cards write
INSERT INTO rbac_policies (role_id, permission_id, scope)
SELECT r.id, p.id, 'own'
FROM rbac_roles r, rbac_permissions p
WHERE r.name = 'CUSTOMER'
  AND p.resource IN ('accounts', 'transfers', 'cards')
  AND p.action = 'write';

-- +goose Down
DROP TABLE IF EXISTS rbac_policies;
DROP TABLE IF EXISTS rbac_permissions;
DROP TABLE IF EXISTS rbac_roles;
