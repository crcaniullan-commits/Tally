-- Categorías globales de gastos: son de tipo 'gasto' y se pueden usar en los
-- gastos de cualquier empresa (empresa_id NULL = global).
INSERT INTO categories (nombre, empresa_id, tipo) VALUES
    ('Insumos', NULL, 'gasto'),
    ('Comunes', NULL, 'gasto'),
    ('Pagos', NULL, 'gasto');
