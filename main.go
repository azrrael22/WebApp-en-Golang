package main

import (
	"encoding/base64" // Se utiliza para codificar la imagen en formato base64.
	"flag"            // Para parsear flags de la línea de comandos.
	"fmt"             // Para imprimir mensajes en la consola.
	"html/template"   // Para renderizar las plantillas HTML.
	"log"             // Para registrar mensajes de error.
	"math/rand"       // Para seleccionar un elemento al azar.
	"net/http"        // Para manejar solicitudes HTTP.
	"os"              // Para trabajar con operaciones de archivos y obtener información del sistema.
	"path/filepath"   // Para construir rutas de archivo de forma segura.
	"strings"         // Para trabajar con cadenas.
)

// Se parsea y carga la plantilla HTML desde el archivo index.html.
var tpl = template.Must(template.ParseFiles("index.html"))

// imageData lee el contenido de la imagen ubicada en 'path', la codifica en base64 y retorna la cadena resultante.
func imageData(path string) string {
	// Se lee el archivo de la imagen.
	data, err := os.ReadFile(path)
	if err != nil {
		// Se finaliza la ejecución si ocurre algún error al leer la imagen.
		log.Fatalf("Error al leer la imagen: %v", err)
	}
	// Se codifica la imagen a base64 y se retorna la cadena.
	return base64.StdEncoding.EncodeToString(data)
}

// randomImagePath busca en el directorio dado una imagen (formatos: .png, .jpg, .jpeg) y retorna su ruta al azar.
func randomImagePath(dir string) (string, error) {
	// Listar los archivos en el directorio.
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var images []string
	// Filtrar sólo los archivos con extensiones válidas.
	for _, file := range files {
		if !file.IsDir() {
			name := strings.ToLower(file.Name())
			if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") {
				images = append(images, file.Name())
			}
		}
	}

	if len(images) == 0 {
		return "", fmt.Errorf("no se encontraron imágenes en el directorio %s", dir)
	}

	// Semilla para el generador aleatorio.
	index := rand.Intn(len(images))
	// Construir la ruta completa.
	return filepath.Join(dir, images[index]), nil
}

// parseFlags se encarga de parsear los flags de línea de comandos y retorna el puerto sobre el que se ejecutará el servidor.
// Por defecto, se utiliza "localhost:8000".
func parseFlags() string {
	var puerto string
	flag.StringVar(&puerto, "p", "localhost:8000", "número de puerto")
	flag.Parse()
	return puerto
}

// indexHandler maneja la ruta raíz ("/") y renderiza la plantilla index.html.
// Se encarga de pasar a la plantilla el nombre del equipo (hostname) y la imagen codificada.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Se obtiene el nombre del equipo en el que se está ejecutando la aplicación.
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "desconocido" // Valor por defecto si ocurre algún error al obtener el hostname.
	}

	// Se obtiene una ruta de imagen aleatoria desde la carpeta "imagenes".
	imgPath, err := randomImagePath("imagenes")
	if err != nil {
		log.Printf("Error: %v", err)
		imgPath = "imagenes/imagen.png" // Ruta por defecto si ocurre un error.
	}

	// Detectar la extensión (sin el punto) para establecer el tipo MIME
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(imgPath)), ".")

	// Se crea la estructura que contiene los datos que se pasarán a la plantilla.
	data := struct {
		Host      string // Nombre del equipo.
		Mime      string // Tipo MIME (png, jpg, jpeg).
		ImageData string // Cadena de la imagen codificada en base64.
	}{
		Host:      hostname,
		Mime:      ext,
		ImageData: imageData(imgPath),
	}

	// Se ejecuta la plantilla enviando los datos. En caso de error, se envía una respuesta HTTP de error.
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// configureRoutes configura las rutas y sus respectivos handlers para el servidor web.
// Aquí se enlaza la ruta "/" al handler indexHandler.
func configureRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	return mux
}

// func main arranca el servidor web
func main() {
	// Se obtiene el puerto a utilizar a partir de los flags.
	puerto := parseFlags()
	fmt.Println("Servidor corriendo en puerto", puerto)

	// Se configuran las rutas de los handlers.
	mux := configureRoutes()

	// Se inicia el servidor web. En caso de fallo, se loguea el error y se finaliza el programa.
	log.Fatal(http.ListenAndServe(puerto, mux))
}
