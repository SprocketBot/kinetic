CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS policies (
    id BIGSERIAL PRIMARY KEY,
    role_name TEXT NOT NULL,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_policies_role_name FOREIGN KEY (role_name) REFERENCES roles(name) ON DELETE CASCADE,
    CONSTRAINT uq_policies_unique_rule UNIQUE (role_name, resource, action)
);

CREATE TABLE IF NOT EXISTS user_role_bindings (
    id BIGSERIAL PRIMARY KEY,
    subject TEXT NOT NULL,
    role_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user_role_bindings_role_name FOREIGN KEY (role_name) REFERENCES roles(name) ON DELETE CASCADE,
    CONSTRAINT uq_user_role_binding UNIQUE (subject, role_name)
);

INSERT INTO roles (name) VALUES ('admin') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name) VALUES ('observer') ON CONFLICT (name) DO NOTHING;

INSERT INTO policies (role_name, resource, action)
VALUES ('admin', 'admin.ping', 'read')
ON CONFLICT (role_name, resource, action) DO NOTHING;
