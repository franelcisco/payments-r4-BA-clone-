package ipfy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// GetIPInfo retrieves the public IP address of the server.
func GetIPInfo() {
	// URL del servicio que devuelve la IP pública
	url := "https://api64.ipify.org?format=json"

	// Realizar una solicitud GET al servicio
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error al realizar la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close() // Asegurarse de cerrar el cuerpo de la respuesta

	// Leer el cuerpo de la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error al leer la respuesta: %v", err)
	}

	// Convertir el cuerpo a string y limpiar espacios en blanco
	ipAddress := strings.TrimSpace(string(body))

	// Imprimir la dirección IP pública
	fmt.Printf("La dirección IP pública del servidor es: %s\n", ipAddress)
}
