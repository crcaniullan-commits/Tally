import { useFocusEffect } from 'expo-router';
import { useCallback, useState } from 'react';
import { useAlerta } from '../../shared/components/Confirmar';
import { useAviso } from '../../shared/components/Toast';
import { traducirError } from '../auth/errores';
import { useSesion } from '../auth/SesionContext';
import { categoriaTieneMovimientos } from '../movimientos/movimientos.service';
import { Categoria, categoriasPredeterminadas, combinarCategorias, TipoMovimiento } from './categorias';
import { crearCategoria, eliminarCategoria, listarCategorias, renombrarCategoria } from './categorias.service';
import { EstiloCategoria, guardarEstilo, leerEstilos } from './estilosPropios';
import { guardarOrden, leerOrden, leerVistaCategorias, ordenar, OrdenCategorias } from './ordenCategorias';

// Toda la gestión de categorías de Inicio (Historia 003): cargar, crear, editar, eliminar y ordenar.
// La pantalla solo muestra los datos y llama a estas funciones.
export function useCategorias() {
  const { token, empresa, manejarError } = useSesion();
  const avisar = useAviso();
  const alerta = useAlerta();
  const [categorias, setCategorias] = useState<Categoria[]>(categoriasPredeterminadas);
  const [orden, setOrden] = useState<OrdenCategorias>({ ingreso: [], egreso: [] });
  const [vistaGrilla, setVistaGrilla] = useState(false); // Preferencia de Configuración

  const cargar = useCallback(async () => {
    if (!token || !empresa) return;
    try {
      const [cats, estilos] = await Promise.all([listarCategorias(empresa.id, token), leerEstilos(empresa.id)]);
      setCategorias(combinarCategorias(cats, estilos));
    } catch (error) {
      manejarError(error);
    }
  }, [token, empresa, manejarError]);

  // Cada vez que se vuelve a la pantalla se recargan categorías, orden y preferencia de vista
  useFocusEffect(useCallback(() => {
    cargar();
    leerVistaCategorias().then((v) => setVistaGrilla(v === 'grilla'));
    if (empresa) leerOrden(empresa.id).then(setOrden);
  }, [cargar, empresa]));

  // Categorías de un tipo, en el orden elegido por el usuario ('ambos' = propias antiguas sin tipo)
  const delTipo = (tipo: TipoMovimiento) => ordenar(categorias.filter((c) => c.tipo === tipo || c.tipo === 'ambos'), orden[tipo]);

  // Crea (editando = null) o edita una categoría. Devuelve la categoría guardada, o null si falló.
  async function guardar(editando: Categoria | null, nombre: string, estilo: EstiloCategoria, tipo: TipoMovimiento): Promise<Categoria | null> {
    if (!token || !empresa) return null;
    try {
      if (editando) {
        // Propia: el nombre se cambia en el backend. Predeterminada: el nombre visible se guarda en el celular.
        if (editando.propia && nombre !== editando.nombre) await renombrarCategoria(editando.id, nombre, empresa.id, token);
        const tipoGuardado = editando.propia ? (editando.tipo === 'ambos' ? tipo : editando.tipo) : undefined;
        await guardarEstilo(empresa.id, editando.clave, {
          icono: estilo.icono, color: estilo.color, colorSuave: estilo.colorSuave,
          ...(editando.propia ? { tipo: tipoGuardado } : { nombre }),
        });
        const editada: Categoria = { ...editando, nombre, icono: estilo.icono, color: estilo.color, colorSuave: estilo.colorSuave, tipo: tipoGuardado ?? editando.tipo };
        setCategorias((actuales) => actuales.map((x) => (x.clave === editando.clave ? editada : x)));
        avisar('¡Categoría actualizada!');
        return editada;
      }
      const creada = await crearCategoria(nombre, empresa.id, token);
      await guardarEstilo(empresa.id, creada.id, { ...estilo, tipo });
      const nueva: Categoria = {
        id: creada.id, nombre: creada.nombre, clave: creada.id, propia: true, tipo,
        icono: estilo.icono, color: estilo.color, colorSuave: estilo.colorSuave,
      };
      setCategorias((actuales) => [...actuales, nueva]);
      avisar('¡Categoría creada!');
      return nueva;
    } catch (error) {
      manejarError(error);
      alerta('No se pudo guardar la categoría', traducirError(error));
      return null;
    }
  }

  // Elimina una categoría propia solo si no tiene registros. Devuelve true si se eliminó.
  async function eliminar(c: Categoria): Promise<boolean> {
    if (!token || !empresa) return false;
    try {
      if (await categoriaTieneMovimientos(c.id, empresa.id, token)) {
        alerta('No se puede eliminar', `"${c.nombre}" tiene registros. Puedes cambiarle el nombre, pero no eliminarla.`);
        return false;
      }
      await eliminarCategoria(c.id, empresa.id, token);
      setCategorias((actuales) => actuales.filter((x) => x.id !== c.id));
      avisar('Categoría eliminada', 'trash');
      return true;
    } catch (error) {
      manejarError(error);
      alerta('No se pudo eliminar', traducirError(error));
      return false;
    }
  }

  // Guarda el nuevo orden de un tipo (arrastrar y soltar)
  function reordenar(tipo: TipoMovimiento, claves: string[]) {
    if (!empresa) return;
    const nuevo = { ...orden, [tipo]: claves };
    setOrden(nuevo);
    guardarOrden(empresa.id, nuevo);
    avisar('Orden guardado', 'grip');
  }

  return { categorias, delTipo, cargar, guardar, eliminar, reordenar, vistaGrilla };
}
