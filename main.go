package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const csvPath = "inquilinos.csv"

func loadData(path string) ([]Inquilino, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Inquilino{}, nil
		}
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	if bom, _ := br.Peek(3); string(bom) == "\xEF\xBB\xBF" {
		br.Discard(3)
	}

	r := csv.NewReader(br)
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1 // Allow variable fields for flexibility

	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("error leyendo cabeceras CSV (%s): %v", path, err)
	}

	// Find column indices
	nameIdx, plazaIdx, rentaIdx := 0, 1, 2
	for i, h := range headers {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "nombre" || h == "name" {
			nameIdx = i
		} else if h == "plaza" || h == "space" || h == "parking" {
			plazaIdx = i
		} else if h == "renta" || h == "price" || h == "importe" {
			rentaIdx = i
		}
	}

	var results []Inquilino
	for i := 0; ; i++ {
		row, err := r.Read()
		if err != nil {
			break
		}
		if len(row) == 0 || (len(row) == 1 && row[0] == "") {
			continue
		}

		var item Inquilino
		if nameIdx < len(row) {
			item.Nombre = strings.TrimSpace(row[nameIdx])
		}
		if plazaIdx < len(row) {
			item.Plaza = strings.TrimSpace(row[plazaIdx])
		}
		if rentaIdx < len(row) {
			renta, err := parseMoneda(row[rentaIdx])
			if err != nil {
				return nil, fmt.Errorf("error en fila %d, renta '%s': %v", i+2, row[rentaIdx], err)
			}
			item.Renta = renta
		}

		if item.Nombre != "" {
			results = append(results, item)
		}
	}

	return results, nil
}

func saveData(path string, inquilinos []Inquilino) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Write header
	if err := w.Write([]string{"Nombre", "Plaza", "Renta"}); err != nil {
		return err
	}

	// Write data
	for _, inq := range inquilinos {
		record := []string{
			inquilineNombre(inq),
			inquilinePlaza(inq),
			fmt.Sprintf("%.2f", inq.Renta),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	return w.Error()
}

func inquilineNombre(i Inquilino) string { return i.Nombre }
func inquilinePlaza(i Inquilino) string  { return i.Plaza }

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			MarginBottom(1)

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF00")).
			Padding(0, 1).
			MarginTop(1)
)

type item struct {
	index     int
	inquilino Inquilino
}

func (i item) Title() string {
	return fmt.Sprintf("%s [%s]", i.inquilino.Nombre, i.inquilino.Plaza)
}

func (i item) Description() string {
	return formatRenta(i.inquilino.Renta)
}

func (i item) FilterValue() string {
	return i.inquilino.Nombre
}

// FormMode represents the current form editing mode
type FormMode int

const (
	FormModeNone FormMode = iota
	FormModeAdd
	FormModeEdit
	FormModeDelete
)

// FormField represents which field is being edited
type FormField int

const (
	FieldNombre FormField = iota
	FieldPlaza
	FieldRenta
)

type model struct {
	list           list.Model
	fecha          time.Time
	estado         int
	seleccionados  map[int]bool
	mensaje        string
	width          int
	height         int
	itemCount      int
	
	// Form state
	formMode       FormMode
	currentField   FormField
	nameInput      textinput.Model
	plazaInput     textinput.Model
	rentaInput     textinput.Model
	editIndex      int
	
	// Data
	inquilinos     []Inquilino
}

const (
	estadoFecha = iota
	estadoSeleccion
	estadoGenerando
	estadoExito
	estadoError
)

func (m model) Init() tea.Cmd {
	return nil
}

func initTextInput(placeholder, value string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(value)
	ti.CharLimit = 100
	ti.Width = 40
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	return ti
}

func (m *model) setupFormForAdd() {
	m.formMode = FormModeAdd
	m.currentField = FieldNombre
	m.nameInput = initTextInput("Nombre del inquilino", "")
	m.plazaInput = initTextInput("Plaza (ej: A1)", "")
	m.rentaInput = initTextInput("Renta mensual (ej: 50.00)", "")
	m.nameInput.Focus()
}

func (m *model) setupFormForEdit(idx int) {
	if idx < 0 || idx >= len(m.inquilinos) {
		return
	}
	m.formMode = FormModeEdit
	m.editIndex = idx
	m.currentField = FieldNombre
	
	inq := m.inquilinos[idx]
	m.nameInput = initTextInput("Nombre del inquilino", inq.Nombre)
	m.plazaInput = initTextInput("Plaza (ej: A1)", inq.Plaza)
	m.rentaInput = initTextInput("Renta mensual (ej: 50.00)", fmt.Sprintf("%.2f", inq.Renta))
	m.nameInput.Focus()
}

func (m *model) getCurrentInput() *textinput.Model {
	switch m.currentField {
	case FieldNombre:
		return &m.nameInput
	case FieldPlaza:
		return &m.plazaInput
	case FieldRenta:
		return &m.rentaInput
	}
	return &m.nameInput
}

func (m *model) nextField() {
	m.currentField++
	if m.currentField > FieldRenta {
		m.currentField = FieldNombre
	}
	
	// Reset focus
	m.nameInput.Blur()
	m.plazaInput.Blur()
	m.rentaInput.Blur()
	
	// Set focus on current field
	switch m.currentField {
	case FieldNombre:
		m.nameInput.Focus()
	case FieldPlaza:
		m.plazaInput.Focus()
	case FieldRenta:
		m.rentaInput.Focus()
	}
}

func (m *model) prevField() {
	m.currentField--
	if m.currentField < FieldNombre {
		m.currentField = FieldRenta
	}
	
	// Reset focus
	m.nameInput.Blur()
	m.plazaInput.Blur()
	m.rentaInput.Blur()
	
	// Set focus on current field
	switch m.currentField {
	case FieldNombre:
		m.nameInput.Focus()
	case FieldPlaza:
		m.plazaInput.Focus()
	case FieldRenta:
		m.rentaInput.Focus()
	}
}

func (m *model) saveCurrentForm() error {
	nombre := strings.TrimSpace(m.nameInput.Value())
	plaza := strings.TrimSpace(m.plazaInput.Value())
	rentaStr := strings.TrimSpace(m.rentaInput.Value())
	
	if nombre == "" {
		return fmt.Errorf("el nombre es obligatorio")
	}
	if plaza == "" {
		return fmt.Errorf("la plaza es obligatoria")
	}
	
	renta, err := parseMoneda(rentaStr)
	if err != nil {
		return fmt.Errorf("renta inválida: %v", err)
	}
	
	switch m.formMode {
	case FormModeAdd:
		m.inquilinos = append(m.inquilinos, Inquilino{
			Nombre: nombre,
			Plaza:  plaza,
			Renta:  renta,
		})
	case FormModeEdit:
		if m.editIndex >= 0 && m.editIndex < len(m.inquilinos) {
			m.inquilinos[m.editIndex] = Inquilino{
				Nombre: nombre,
				Plaza:  plaza,
				Renta:  renta,
			}
		}
	}
	
	// Save to CSV
	if err := saveData(csvPath, m.inquilinos); err != nil {
		return fmt.Errorf("error guardando datos: %v", err)
	}
	
	return nil
}

func (m *model) deleteCurrentItem() error {
	idx := m.list.Index()
	if idx < 0 || idx >= len(m.inquilinos) {
		return fmt.Errorf("índice inválido")
	}
	
	// Remove from slice
	m.inquilinos = append(m.inquilinos[:idx], m.inquilinos[idx+1:]...)
	
	// Save to CSV
	if err := saveData(csvPath, m.inquilinos); err != nil {
		return fmt.Errorf("error guardando datos: %v", err)
	}
	
	return nil
}

func (m *model) refreshList() {
	var items []list.Item
	for i, inq := range m.inquilinos {
		items = append(items, item{index: i, inquilino: inq})
	}
	m.list.SetItems(items)
	m.itemCount = len(m.inquilinos)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.height > 20 {
			m.list.SetHeight(m.height - 10)
		}
		return m, nil

	case tea.KeyMsg:
		// Handle form mode keys
		if m.formMode != FormModeNone {
			switch msg.String() {
			case "ctrl+c", "q":
				if m.formMode == FormModeDelete {
					// Cancel delete
					m.formMode = FormModeNone
					return m, nil
				}
				return m, tea.Quit
			
			case "esc":
				// Cancel form
				m.formMode = FormModeNone
				m.nameInput.Blur()
				m.plazaInput.Blur()
				m.rentaInput.Blur()
				return m, nil
			
			case "tab":
				m.nextField()
				return m, nil
			
			case "shift+tab":
				m.prevField()
				return m, nil
			
			case "enter":
				// Save form
				if err := m.saveCurrentForm(); err != nil {
					m.mensaje = err.Error()
					m.estado = estadoError
					m.formMode = FormModeNone
					m.nameInput.Blur()
					m.plazaInput.Blur()
					m.rentaInput.Blur()
					return m, nil
				}
				m.refreshList()
				m.formMode = FormModeNone
				m.nameInput.Blur()
				m.plazaInput.Blur()
				m.rentaInput.Blur()
				m.estado = estadoSeleccion
				m.mensaje = ""
				return m, nil
			}
			
			// Update current input field
			input := m.getCurrentInput()
			var cmd tea.Cmd
			*input, cmd = input.Update(msg)
			return m, cmd
		}
		
		// Handle delete confirmation
		if m.formMode == FormModeDelete {
			switch msg.String() {
			case "y", "Y":
				if err := m.deleteCurrentItem(); err != nil {
					m.mensaje = err.Error()
					m.estado = estadoError
				} else {
					m.refreshList()
					m.mensaje = "Inquilino eliminado"
				}
				m.formMode = FormModeNone
				return m, nil
			case "n", "N", "esc":
				m.formMode = FormModeNone
				return m, nil
			}
			return m, nil
		}
		
		// Normal mode keys
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			if m.estado == estadoFecha {
				m.estado = estadoSeleccion
				return m, nil
			} else if m.estado == estadoSeleccion {
				m.estado = estadoGenerando
				return m, generarRecibosCmd(m)
			}
		
		case "a":
			// Add new tenant
			m.setupFormForAdd()
			return m, nil
		
		case "e":
			// Edit selected tenant
			idx := m.list.Index()
			if idx >= 0 && idx < len(m.inquilinos) {
				m.setupFormForEdit(idx)
			}
			return m, nil
		
		case "d":
			// Delete selected tenant (with confirmation)
			idx := m.list.Index()
			if idx >= 0 && idx < len(m.inquilinos) {
				m.formMode = FormModeDelete
				m.editIndex = idx
			}
			return m, nil

		case " ":
			if m.estado == estadoSeleccion {
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.inquilinos) {
					if m.seleccionados[idx] {
						delete(m.seleccionados, idx)
					} else {
						m.seleccionados[idx] = true
					}
				}
				return m, nil
			}
		
		case "A":
			if m.estado == estadoSeleccion {
				if len(m.seleccionados) == m.itemCount {
					m.seleccionados = make(map[int]bool)
				} else {
					m.seleccionados = make(map[int]bool)
					for i := 0; i < m.itemCount; i++ {
						m.seleccionados[i] = true
					}
				}
				return m, nil
			}
		}

	case fechaMsg:
		m.fecha = time.Time(msg)
		m.estado = estadoSeleccion
		return m, nil

	case generacionMsg:
		if msg.err != nil {
			m.estado = estadoError
			m.mensaje = msg.err.Error()
		} else {
			m.estado = estadoExito
			m.mensaje = msg.filename
		}
		return m, nil
	}

	var cmd tea.Cmd
	if m.estado == estadoSeleccion && m.formMode == FormModeNone {
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🏠 Generador de Recibos de Garaje") + "\n\n")

	// Show form if in add/edit mode
	if m.formMode == FormModeAdd || m.formMode == FormModeEdit {
		modeStr := "Añadir"
		if m.formMode == FormModeEdit {
			modeStr = "Editar"
		}
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render(fmt.Sprintf("%s Inquilino", modeStr)) + "\n\n")
		
		fields := []struct{
			label string
			input textinput.Model
		}{
			{"Nombre:", m.nameInput},
			{"Plaza:", m.plazaInput},
			{"Renta:", m.rentaInput},
		}
		
		for i, f := range fields {
			isCurrent := (i == int(m.currentField))
			style := lipgloss.NewStyle()
			if isCurrent {
				style = style.Foreground(lipgloss.Color("#00FF00")).Bold(true)
			} else {
				style = style.Foreground(lipgloss.Color("#AAAAAA"))
			}
			b.WriteString(style.Render(f.label + " "))
			b.WriteString(inputStyle.Render(f.input.View()))
			b.WriteString("\n")
		}
		
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("[Tab] Siguiente campo  [Shift+Tab] Anterior  [Enter] Guardar  [Esc] Cancelar") + "\n")
		return b.String()
	}

	// Show delete confirmation
	if m.formMode == FormModeDelete {
		b.WriteString(errorStyle.Render("⚠️  ¿Eliminar inquilino?") + "\n\n")
		if m.editIndex >= 0 && m.editIndex < len(m.inquilinos) {
			inq := m.inquilinos[m.editIndex]
			b.WriteString(fmt.Sprintf("Nombre: %s\n", inq.Nombre))
			b.WriteString(fmt.Sprintf("Plaza: %s\n", inq.Plaza))
			b.WriteString(fmt.Sprintf("Renta: %s\n", formatRenta(inq.Renta)))
		}
		b.WriteString("\n")
		b.WriteString(successStyle.Render("[Y] Sí, eliminar  ") + errorStyle.Render("[N/ESC] Cancelar") + "\n")
		return b.String()
	}

	switch m.estado {
	case estadoFecha:
		b.WriteString("Introduce la fecha (Mes Año) o presiona Enter para usar la fecha por defecto:\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("Por defecto: %s)", formatMesAnyo(time.Now()))))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("(Esta funcionalidad requiere entrada de texto - usa la versión anterior para esta opción)\n"))
		b.WriteString("\nPresiona Enter para continuar con la fecha por defecto...")

	case estadoSeleccion:
		count := len(m.seleccionados)
		b.WriteString(dateStyle.Render(formatMesAnyo(m.fecha)) + "\n\n")
		b.WriteString(fmt.Sprintf("Selecciona inquilinos: [espacio] marcar/desmarcar, [a] todos, [enter] generar\n"))
		b.WriteString(fmt.Sprintf("Seleccionados: %d/%d\n\n", count, m.itemCount))
		b.WriteString(helpStyle.Render("[a] Añadir  [e] Editar  [d] Eliminar  [/] Filtrar  [q] Salir\n\n"))
		b.WriteString(m.list.View())

	case estadoGenerando:
		b.WriteString("Generando PDF...\n")

	case estadoExito:
		b.WriteString(successStyle.Render("✓ " + m.mensaje) + "\n\n")
		b.WriteString("Presiona 'q' para salir.\n")

	case estadoError:
		b.WriteString(errorStyle.Render("✗ Error: "+m.mensaje) + "\n\n")
		b.WriteString("Presiona 'q' para salir.\n")
	}

	return b.String()
}

type fechaMsg time.Time

type generacionMsg struct {
	filename string
	err      error
}

func generarRecibosCmd(m model) tea.Cmd {
	return func() tea.Msg {
		items := m.list.Items()
		var datosFiltrados []Inquilino

		if len(m.seleccionados) == 0 {
			for _, it := range items {
				item := it.(item)
				datosFiltrados = append(datosFiltrados, item.inquilino)
			}
		} else {
			for i, it := range items {
				if m.seleccionados[i] {
					item := it.(item)
					datosFiltrados = append(datosFiltrados, item.inquilino)
				}
			}
		}

		if len(datosFiltrados) == 0 {
			return generacionMsg{err: fmt.Errorf("no hay inquilinos seleccionados")}
		}

		err := generarRecibos(datosFiltrados, m.fecha)
		if err != nil {
			return generacionMsg{err: err}
		}

		filename := fmt.Sprintf("garajes_%s.pdf", m.fecha.Format("2006-01"))
		return generacionMsg{filename: filename}
	}
}

func main() {
	inquilinos, err := loadData(csvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al cargar los datos: %v\n", err)
		os.Exit(1)
	}

	defaultDate := time.Now()
	if defaultDate.Day() > 3 {
		defaultDate = defaultDate.AddDate(0, 1, 0)
	}
	defaultDate = time.Date(defaultDate.Year(), defaultDate.Month(), 1, 0, 0, 0, 0, defaultDate.Location())

	var items []list.Item
	for i, inq := range inquilinos {
		items = append(items, item{index: i, inquilino: inq})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Inquilinos"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	m := model{
		list:          l,
		fecha:         defaultDate,
		estado:        estadoFecha,
		seleccionados: make(map[int]bool),
		itemCount:     len(inquilinos),
		inquilinos:    inquilinos,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
