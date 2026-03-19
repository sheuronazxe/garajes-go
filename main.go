package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var inputReader = bufio.NewScanner(os.Stdin)

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func loadData(path string) ([]Inquilino, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	if bom, _ := br.Peek(3); string(bom) == "\xEF\xBB\xBF" {
		br.Discard(3)
	}

	r := csv.NewReader(br)
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = 4

	_, _ = r.Read()

	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error leyendo CSV (%s): %v", path, err)
	}

	var results []Inquilino
	for i, row := range rows {
		var item Inquilino
		if err := item.FromRow(row); err != nil {
			return nil, fmt.Errorf("error en fila %d de %s: %v", i+1, path, err)
		}
		results = append(results, item)
	}

	return results, nil
}

func obtenerFecha() (time.Time, error) {
	t := time.Now()
	if t.Day() > 3 {
		t = t.AddDate(0, 1, 0)
	}

	def := fmt.Sprintf("%s %d", meses[t.Month()], t.Year())
	fmt.Printf("Introduce fecha [%s]: ", def)

	inputReader.Scan()
	input := strings.TrimSpace(inputReader.Text())
	if input == "" {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()), nil
	}

	partes := strings.Fields(input)
	if len(partes) != 2 {
		return time.Time{}, fmt.Errorf("formato incorrecto. Usa: 'Mes Año'")
	}

	nombreMes := strings.ToLower(partes[0])
	añoStr := partes[1]

	mes := time.Month(0)
	for m, n := range meses {
		if n == nombreMes {
			mes = m
			break
		}
	}

	if mes == 0 {
		return time.Time{}, fmt.Errorf("mes inválido: %s", nombreMes)
	}

	año, err := strconv.Atoi(añoStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("el año debe ser numérico")
	}

	return time.Date(año, mes, 1, 0, 0, 0, 0, t.Location()), nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		fmt.Printf("No se puede abrir el navegador en esta plataforma: %s\n", runtime.GOOS)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error al abrir el navegador: %v\n", err)
	}
}

func main() {
	inquilinos, err := loadData("inquilinos.csv")
	if err != nil {
		fatal("Error al cargar los datos: %v", err)
	}

	fecha, err := obtenerFecha()
	if err != nil {
		fatal("Error al obtener la fecha: %v", err)
	}

	fechaStr := formatMesAnyo(fecha)

	for i, item := range inquilinos {
		fmt.Printf("\033[32m%2d\033[0m %-30.30s  \033[33m%-25.25s\033[0m  %10.2f [\033[36m%s\033[0m]\n",
			i, item.Nombre, item.Plaza, item.Renta, generateID(item.Nombre, fechaStr))
	}

	fmt.Printf("¿Qué registros desea imprimir? (ej: 1,3,10) [todos]: ")
	inputReader.Scan()
	seleccion := strings.TrimSpace(inputReader.Text())

	datosFiltrados := inquilinos
	if seleccion != "" {
		datosFiltrados = nil
		for s := range strings.SplitSeq(seleccion, ",") {
			if i, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && i >= 0 && i < len(inquilinos) {
				datosFiltrados = append(datosFiltrados, inquilinos[i])
			}
		}
	}

	if len(datosFiltrados) == 0 {
		fatal("Selección no válida")
	}

	fmt.Printf("\nGenerando PDF con %d recibo(s)...\n", len(datosFiltrados))

	err = generarRecibos(datosFiltrados, fecha)
	if err != nil {
		fatal("Error al generar el PDF: %v", err)
	}

	filename := fmt.Sprintf("garajes_%s.pdf", fecha.Format("2006-01"))
	fmt.Println("\n✓ PDF generado exitosamente: " + filename)
	openBrowser(filename)
}
