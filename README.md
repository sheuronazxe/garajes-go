El programa lee los datos del fichero inquilinos.csv con el siguiente formato:

NOMBRE,PLAZA,RENTA,CONTACTO

Se puede exportar la hoja con los datos de garajes de Google Sheet directamente en formato .csv con comas, copiarlo en la carpeta del programa y ponerle de nombre datos.csv

Compilado con:
go build -ldflags "-s -w"
