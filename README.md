# 🏠 Generador de Recibos de Garaje

Una aplicación TUI (Terminal User Interface) moderna para generar recibos de garaje en PDF, construida con [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Características

- ✅ **Interfaz TUI interactiva** con navegación por teclado
- ✅ **Añadir inquilinos** desde la interfaz
- ✅ **Editar inquilinos** existentes
- ✅ **Eliminar inquilinos** con confirmación
- ✅ **Selección múltiple** para generar recibos específicos
- ✅ **Filtrado** por nombre de inquilino
- ✅ **Generación de PDF** profesional con diseño limpio
- ✅ **Persistencia de datos** en CSV

## Instalación

### Requisitos
- Go 1.19 o superior

### Compilación
```bash
go build -ldflags "-s -w" -o garajes .
```

## Uso

### Ejecutar
```bash
./garajes
```

### Controles

#### Navegación Principal
| Tecla | Acción |
|-------|--------|
| `↑` / `↓` | Navegar por la lista |
| `/` | Filtrar por nombre |
| `Espacio` | Marcar/desmarcar inquilino |
| `A` (mayúscula) | Seleccionar todos/ninguno |
| `Enter` | Generar PDF |
| `q` o `Ctrl+C` | Salir |

#### Gestión de Inquilinos
| Tecla | Acción |
|-------|--------|
| `a` | **Añadir** nuevo inquilino |
| `e` | **Editar** inquilino seleccionado |
| `d` | **Eliminar** inquilino seleccionado |

#### Formularios (Añadir/Editar)
| Tecla | Acción |
|-------|--------|
| `Tab` | Siguiente campo |
| `Shift+Tab` | Campo anterior |
| `Enter` | Guardar |
| `Esc` | Cancelar |

#### Eliminación
| Tecla | Acción |
|-------|--------|
| `Y` | Confirmar eliminación |
| `N` / `Esc` | Cancelar |

## Formato del CSV

El archivo `inquilinos.csv` debe tener el siguiente formato:

```csv
Nombre,Plaza,Renta
Juan Pérez,A1,50.00
María García,B2,75.50
Carlos López,C3,60.00
```

### Columnas soportadas
- **Nombre**: Nombre del inquilino (obligatorio)
- **Plaza**: Identificador de la plaza (obligatorio)
- **Renta**: Importe mensual en euros (obligatorio)

El programa es flexible con los nombres de las columnas y acepta variaciones como:
- Nombre: `nombre`, `name`
- Plaza: `plaza`, `space`, `parking`
- Renta: `renta`, `price`, `importe`

## Estructura del Proyecto

```
garajes/
├── main.go          # Interfaz TUI y lógica principal
├── models.go        # Modelo de datos y utilidades
├── pdf.go           # Generación de PDF
├── fonts/           # Fuentes para el PDF
│   ├── Roboto-Regular.ttf
│   ├── Roboto-Bold.ttf
│   └── Roboto-Light.ttf
├── fonts.go         # Embed de fuentes
├── inquilinos.csv   # Datos de inquilinos
└── README.md        # Este archivo
```

## Mejoras Sugeridas

Aquí tienes algunas ideas para mejorar aún más la aplicación:

### Funcionalidad
1. **Búsqueda avanzada**: Añadir búsqueda por plaza o por importe
2. **Historial**: Mantener un historial de PDFs generados
3. **Plantillas personalizables**: Permitir editar el diseño del recibo
4. **Exportación a Excel**: Además de PDF, permitir exportar a XLSX
5. **Recordatorios**: Integración con email para enviar recordatorios de pago
6. **Estadísticas**: Dashboard con ingresos totales, morosidad, etc.
7. **Multi-usuario**: Soporte para múltiples propiedades/edificios
8. **Backup automático**: Copias de seguridad periódicas del CSV

### Interfaz
1. **Tema configurable**: Permitir cambiar colores y estilos
2. **Atajos personalizables**: Configurar teclas según preferencia
3. **Vista de detalles**: Mostrar información completa del inquilino
4. **Ordenamiento**: Ordenar por nombre, plaza o importe
5. **Paginación**: Mejor manejo de listas largas

### Técnica
1. **Base de datos SQLite**: En lugar de CSV para mejor rendimiento
2. **API REST**: Para integración con otras aplicaciones
3. **Versión web**: Usar Bubble Tea para web con Bubbles
4. **Tests unitarios**: Aumentar cobertura de tests
5. **CI/CD**: Pipeline de integración continua

## Licencia

Este proyecto es de código abierto. Siéntete libre de modificarlo y distribuirlo.

## Autor

Desarrollado con ❤️ usando Go y Bubble Tea.
