DELETE FROM categories
WHERE empresa_id IS NULL
  AND tipo = 'gasto'
  AND nombre IN ('Insumos', 'Comunes', 'Pagos');
