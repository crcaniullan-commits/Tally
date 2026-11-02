-- Eliminar categorías globales
DELETE FROM categories WHERE nombre IN ('Ventas', 'Servicios', 'Pedidos') AND empresa_id IS NULL;
