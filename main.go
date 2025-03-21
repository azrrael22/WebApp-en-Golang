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

// randomImagePaths busca en el directorio dado todas las imágenes (formatos: .png, .jpg, .jpeg),
// las mezcla de forma aleatoria y retorna un slice con 'count' rutas.
func randomImagePaths(dir string, count int) ([]string, error) {
	// Listar los archivos en el directorio.
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var images []string
	// Filtrar únicamente los archivos con extensiones válidas.
	for _, file := range files {
		if !file.IsDir() {
			name := strings.ToLower(file.Name())
			if strings.HasSuffix(name, ".png") ||
				strings.HasSuffix(name, ".jpg") ||
				strings.HasSuffix(name, ".jpeg") {
				images = append(images, file.Name())
			}
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no se encontraron imágenes en el directorio %s", dir)
	}

	// Mezclar aleatoriamente el slice.
	rand.Shuffle(len(images), func(i, j int) {
		images[i], images[j] = images[j], images[i]
	})

	// Si se solicita más imágenes de las disponibles, ajustar count.
	if count > len(images) {
		count = len(images)
	}

	// Construir la ruta completa para cada imagen seleccionada.
	var paths []string
	for _, filename := range images[:count] {
		paths = append(paths, filepath.Join(dir, filename))
	}

	return paths, nil
}

// parseFlags se encarga de parsear los flags de línea de comandos y retorna el puerto sobre el que se ejecutará el servidor.
// Por defecto, se utiliza "localhost:8000".
func parseFlags() string {
	var puerto string
	flag.StringVar(&puerto, "p", "localhost:8000", "número de puerto")
	flag.Parse()
	return puerto
}

// indexHandler ahora obtiene cuatro imágenes al azar y pasa la información a la plantilla.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Se obtiene el nombre del equipo.
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "desconocido"
	}

	// Se obtienen 4 rutas de imagen aleatorias desde la carpeta "imagenes".
	paths, err := randomImagePaths("imagenes", 4)
	if err != nil {
		log.Printf("Error: %v", err)
		// Opcional: definir ruta por defecto si ocurre un error.
		paths = []string{"imagenes/imagen.png"}
	}

	// Definir una estructura para contener la información de cada imagen.
	type ImageInfo struct {
		Mime      string // Extensión sin el punto, para el tipo MIME.
		ImageData string // Cadena codificada en base64.
		ImageName string // Nombre del archivo.
	}

	var images []ImageInfo
	for _, p := range paths {
		// Detectar la extensión sin el punto.
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(p)), ".")
		images = append(images, ImageInfo{
			Mime:      ext,
			ImageData: imageData(p),
			ImageName: filepath.Base(p),
		})
	}

	// Se crea la estructura que se pasará a la plantilla.
	data := struct {
		Host   string      // Nombre del equipo.
		Images []ImageInfo // Slice con la información de las imágenes.
	}{
		Host:   hostname,
		Images: images,
	}

	// Se ejecuta la plantilla enviando los datos.
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
