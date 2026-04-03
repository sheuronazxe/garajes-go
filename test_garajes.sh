#!/bin/bash
# Script para probar la aplicación garajes

echo "Probando aplicación de garajes..."
echo ""

# Crear un script expect para simular interacción
cat > /tmp/test_interact.exp << 'EXPECT'
#!/usr/bin/expect -f
set timeout 5

spawn ./garajes

# Esperar a que aparezca el prompt de fecha
expect {
    "MM AAAA" {
        send "12 2024\r"
    }
    timeout {
        puts "Timeout esperando input de fecha"
        exit 1
    }
}

# Esperar un poco para ver la lista
sleep 2

# Enviar 'q' para salir
send "q"

expect eof
EXPECT

chmod +x /tmp/test_interact.exp

# Ejecutar prueba si expect está disponible
if command -v expect &> /dev/null; then
    echo "Ejecutando prueba interactiva..."
    /tmp/test_interact.exp
else
    echo "Expect no disponible, ejecutando prueba básica..."
    # Prueba básica: solo verificar que arranca
    timeout 3 ./garajes < /dev/null || true
fi

echo ""
echo "Prueba completada"
