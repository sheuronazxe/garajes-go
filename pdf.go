package main

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"garajes/fonts"
)

const (
	AnchoPagina      = 210.0
	recibosPorPagina = 5
	alturaRecibo     = 297.0 / float64(recibosPorPagina)
	margen           = 15.0
)

func newPDF() (*fpdf.Fpdf, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)

	type fontDef struct {
		family, style, file string
	}

	fontsToLoad := []fontDef{
		{"Roboto", "", "Roboto-Regular.ttf"},
		{"Roboto", "B", "Roboto-Bold.ttf"},
		{"Roboto-Light", "", "Roboto-Light.ttf"},
	}

	for _, f := range fontsToLoad {
		bytes, err := fonts.FS.ReadFile(f.file)
		if err != nil {
			return nil, fmt.Errorf("error cargando fuente %s: %w", f.file, err)
		}
		pdf.AddUTF8FontFromBytes(f.family, f.style, bytes)
	}

	return pdf, nil
}

func generarRecibos(datos []Inquilino, fecha time.Time) error {
	pdf, err := newPDF()
	if err != nil {
		return fmt.Errorf("error creando PDF: %w", err)
	}

	pdf.SetMargins(margen, 10, margen)
	pdf.SetCellMargin(0)
	pdf.SetDrawColor(160, 160, 160)

	fechaStr := formatMesAnyo(fecha)

	slices.SortFunc(datos, func(a, b Inquilino) int {
		return strings.Compare(a.Nombre, b.Nombre)
	})

	totalPaginas := (len(datos) + recibosPorPagina - 1) / recibosPorPagina
	if totalPaginas > 0 {
		paginas := make([][]Inquilino, totalPaginas)
		for i, registro := range datos {
			pagina := i % totalPaginas
			paginas[pagina] = append(paginas[pagina], registro)
		}

		for _, registrosPagina := range paginas {
			pdf.AddPage()
			for posicion, registro := range registrosPagina {
				yOffset := float64(posicion) * alturaRecibo
				dibujarRecibo(pdf, registro, fechaStr, yOffset)
			}
		}
	}

	filename := fmt.Sprintf("garajes_%s.pdf", fecha.Format("2006-01"))
	return pdf.OutputFileAndClose(filename)
}

func dibujarRecibo(pdf *fpdf.Fpdf, inquilino Inquilino, fecha string, y float64) {
	id := generateID(inquilino.Nombre, fecha)

	// Marcas de corte
	pdf.SetLineWidth(0.3)
	pdf.Line(0, y+alturaRecibo, 10, y+alturaRecibo)
	pdf.Line(AnchoPagina-10, y+alturaRecibo, AnchoPagina, y+alturaRecibo)

	// Separadores
	pdf.Line(margen, y+20, AnchoPagina-margen, y+20)
	pdf.Line(margen, y+45, AnchoPagina-margen, y+45)

	// Título principal
	pdf.SetY(y + 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Roboto-Light", "", 18)
	pdf.CellFormat(110, 7, "RECIBO DE GARAJE", "0", 0, "LA", false, 0, "")

	// Fecha en negrita
	pdf.SetFont("Roboto", "B", 14)
	pdf.CellFormat(70, 7, fecha, "0", 0, "RA", false, 0, "")

	// Etiquetas de formato
	pdf.SetTextColor(120, 120, 120)
	pdf.SetFont("Roboto-Light", "", 9)

	pdf.Ln(16)
	pdf.CellFormat(90, 6, "INQUILINO", "0", 0, "L", false, 0, "")
	pdf.CellFormat(45, 6, "PLAZA", "0", 0, "C", false, 0, "")
	pdf.CellFormat(45, 6, "IMPORTE", "0", 0, "R", false, 0, "")

	// Valores de los datos
	pdf.Ln(8)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Roboto", "", 16)

	pdf.CellFormat(90, 6, inquilino.Nombre, "0", 0, "L", false, 0, "")
	pdf.CellFormat(45, 6, inquilino.Plaza, "0", 0, "C", false, 0, "")
	pdf.CellFormat(45, 6, formatRenta(inquilino.Renta), "0", 0, "R", false, 0, "")

	// Referencia recibo
	pdf.Ln(14)
	pdf.SetTextColor(120, 120, 120)
	pdf.SetFont("Roboto-Light", "", 9)
	pdf.CellFormat(180, 7, "Referencia: "+id, "0", 0, "RT", false, 0, "")
}
