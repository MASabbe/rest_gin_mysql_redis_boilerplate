-- Phase 3: Role-Based Access Control (RBAC) Down Migration
-- Drops RBAC tables in reverse referential dependency order

DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
