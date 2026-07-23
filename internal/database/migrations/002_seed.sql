INSERT OR IGNORE INTO roles (name, description) VALUES ('admin', '系统管理员，拥有所有权限');

INSERT OR IGNORE INTO role_permissions (role_id, module, action)
SELECT r.id, m.module, m.action
FROM roles r
CROSS JOIN (
    SELECT 'person' AS module, 'read' AS action UNION ALL
    SELECT 'person', 'write' UNION ALL
    SELECT 'person', 'delete' UNION ALL
    SELECT 'organization', 'read' UNION ALL
    SELECT 'organization', 'write' UNION ALL
    SELECT 'organization', 'delete' UNION ALL
    SELECT 'attendance', 'read' UNION ALL
    SELECT 'attendance', 'write' UNION ALL
    SELECT 'attendance', 'delete' UNION ALL
    SELECT 'salary', 'read' UNION ALL
    SELECT 'salary', 'write' UNION ALL
    SELECT 'salary', 'delete' UNION ALL
    SELECT 'file', 'read' UNION ALL
    SELECT 'file', 'write' UNION ALL
    SELECT 'file', 'delete' UNION ALL
    SELECT 'audit', 'read' UNION ALL
    SELECT 'audit', 'write' UNION ALL
    SELECT 'audit', 'delete'
) m
WHERE r.name = 'admin';

INSERT OR IGNORE INTO entities (type, name) VALUES ('organization', '默认组织');
