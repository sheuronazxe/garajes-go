package main

import (
	"fmt"
	"hash/adler32"
	"math"
	"strconv"
	"strings"
	"time"
)

// Inquilino representa los datos de un inquilino de garaje.
type Inquilino struct {
	Nombre string  // Nombre del inquilino
	Plaza  string  // Número o identificador de la plaza
	Renta  float64 // Renta mensual en euros
}

func (inquilino *Inquilino) FromRow(linea []string) error {
	inquilino.Nombre = linea[0]
	inquilino.Plaza = linea[1]

	renta, err := parseMoneda(linea[2])
	if err != nil {
		return fmt.Errorf("error al convertir renta '%s': %v", linea[2], err)
	}
	inquilino.Renta = renta

	return nil
}

func parseMoneda(s string) (float64, error) {
	cleaned := strings.NewReplacer("€", "", ".", "", ",", ".").Replace(s)
	return strconv.ParseFloat(strings.TrimSpace(cleaned), 64)
}

func formatRenta(val float64) string {
	val = math.Round(val*100) / 100
	signo := ""
	if val < 0 {
		signo = "-"
		val = -val
	}
	s := fmt.Sprintf("%.2f", val)
	s = strings.ReplaceAll(s, ".", ",")
	parts := strings.Split(s, ",")
	intPart, decPart := parts[0], parts[1]

	if len(intPart) > 3 {
		var result []byte
		for i, c := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result = append(result, '.')
			}
			result = append(result, byte(c))
		}
		intPart = string(result)
	}

	return signo + intPart + "," + decPart + " €"
}

func generateID(nombre, fecha string) string {
	data := []byte(nombre + fecha)
	return fmt.Sprintf("%08x", adler32.Checksum(data))
}

var meses = map[time.Month]string{
	time.January: "enero", time.February: "febrero", time.March: "marzo",
	time.April: "abril", time.May: "mayo", time.June: "junio",
	time.July: "julio", time.August: "agosto", time.September: "septiembre",
	time.October: "octubre", time.November: "noviembre", time.December: "diciembre",
}

func formatMesAnyo(fecha time.Time) string {
	return fmt.Sprintf("%s %d", meses[fecha.Month()], fecha.Year())
}
