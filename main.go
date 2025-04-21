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

// Declaramos config como variable global
var config Config
var tpl = template.Must(template.ParseFiles("index.html"))

// Primero, creamos una estructura para almacenar la configuración
type Config struct {
	Puerto    string
	ImagenDir string
}

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


// randomImagePaths escanea el directorio especificado buscando archivos de imagen con extensiones .png, .jpg o .jpeg,
// los mezcla aleatoriamente y devuelve un slice que contiene las rutas completas de un subconjunto de las imágenes encontradas.
// Si el número solicitado es mayor que la cantidad de imágenes disponibles, se ajusta para devolver todas las imágenes disponibles.
// La función devuelve un error si no se puede leer el directorio o si no se encuentran archivos de imagen.

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

// Modificamos parseFlags para que devuelva la estructura Config
func parseFlags() Config {
	config := Config{}

	// Definimos los flags - cambiamos el valor por defecto y el mensaje
	flag.StringVar(&config.Puerto, "p", "8000", "número de puerto (ejemplo: -p 8000)")
	flag.StringVar(&config.ImagenDir, "i", "./", "directorio de imágenes (ejemplo: -i ./imagenes)")

	// Parseamos los flags
	flag.Parse()

	// Añadimos solo ":" al puerto
	config.Puerto = ":" + config.Puerto

	return config
}

// indexHandler obtiene ahora cuatro imágenes al azar y pasa la información a la plantilla.
// indexHandler funciona como el manejador HTTP para la ruta principal.
//
// Realiza los siguientes pasos:
//  1. Recupera el nombre del host de la máquina que ejecuta el servidor. Si ocurre un error,
//     el nombre del host se establece por defecto a "desconocido".
//  2. Obtiene una cantidad determinada (4) de rutas aleatorias de archivos de imagen del directorio "imagenes"
//     utilizando la función randomImagePaths. Si la obtención falla,
//     se utiliza una ruta de imagen predeterminada ("imagenes/imagen.png").
//  3. Para cada ruta de imagen obtenida, realiza lo siguiente:
//     - Determina el tipo MIME extrayendo la extensión del archivo (sin el punto).
//     - Codifica los datos de la imagen en base64 utilizando la función imageData.
//     - Extrae el nombre del archivo de la imagen.
//  4. Construye una estructura de datos que contiene el nombre del host y un slice con la información de las imágenes.
//  5. Ejecuta una plantilla (tpl) utilizando el objeto de datos construido, escribiendo el contenido renderizado
//     en la respuesta HTTP.
//
// Si la ejecución de la plantilla falla, el manejador responde con un error HTTP 500.
// indexHandler atiende las solicitudes HTTP para la página de inicio.
//
// Realiza las siguientes operaciones:
//   - Recupera el nombre del host de la máquina, utilizando "desconocido" como valor por defecto en caso de error.
//   - Obtiene las rutas de 4 imágenes aleatorias del directorio "imagenes" mediante la función randomImagePaths.
//     Si se produce un error durante esta operación, se registra el error y se establece una ruta de imagen predeterminada.
//   - Para cada ruta de imagen obtenida, realiza lo siguiente:
//   - Se deduce el tipo MIME extrayendo la extensión del archivo (sin el punto).
//   - Se codifican los datos de la imagen en una cadena base64 utilizando la función imageData.
//   - Se recopilan el tipo MIME de la imagen, los datos codificados en base64 y el nombre del archivo
//     en una estructura ImageInfo.
//   - Se agregan las estructuras ImageInfo a un slice y se empaquetan junto con el nombre del host
//     en una estructura de datos.
//   - Se ejecuta la plantilla utilizando los datos ensamblados para renderizar la respuesta, respondiendo con un error HTTP 500
//     si la ejecución de la plantilla falla.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Se obtiene el nombre del equipo.
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "desconocido"
	}

	// Se obtienen 4 rutas de imagen aleatorias desde la carpeta configurada
	paths, err := randomImagePaths(config.ImagenDir, 4)
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

// Modificamos main para usar la nueva configuración
func main() {
	// Inicializamos la configuración global
	config = parseFlags()
	fmt.Printf("Servidor corriendo en puerto %s\nUsando directorio de imágenes: %s\n",
		config.Puerto, config.ImagenDir)

	// Se configuran las rutas de los handlers
	mux := configureRoutes()

	// Se inicia el servidor web
	log.Fatal(http.ListenAndServe(config.Puerto, mux))
}
