package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type DownloadTask struct {
	Artist string
	Song   string
	URL    string // Para playlists de YouTube
}

type DownloadStats struct {
	success int32
	failed  int32
	skipped int32
}

type PlaylistInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

var (
	useCookies         bool = false // Por defecto NO intentar usar cookies
	turboMode          bool = false // Modo turbo activado por defecto
	qualityMode        bool = false // Modo calidad desactivado por defecto
	audioQuality       string
	recommendedWorkers int // Número recomendado de workers según CPU
	downloadedSongs    map[string]bool
	downloadedMutex    sync.RWMutex
)

func calculateRecommendedWorkers() int {
	cpus := runtime.NumCPU()
	memoryGB := getMemoryGB()

	// Lógica más inteligente basada en recursos
	workers := cpus * 2 // Más conservador

	// Ajustar basado en memoria
	if memoryGB >= 16 {
		workers = cpus * 3
	} else if memoryGB >= 8 {
		workers = cpus * 2
	} else {
		workers = cpus
	}

	// Límites más conservadores para reducir carga
	if workers < 4 {
		workers = 4
	}
	if workers > 12 { // Reducido de 16 a 12
		workers = 12
	}

	return workers
}

func getMemoryGB() uint64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	// Esto es memoria del programa, pero nos da una idea
	return uint64(mem.Sys / 1024 / 1024 / 1024)
}

func toggleTurboMode() {
	turboMode = !turboMode
	if qualityMode {
		qualityMode = false
		fmt.Println("⚠️  Modo Calidad DESACTIVADO automáticamente al activar Modo Turbo")
	}

	if turboMode {
		fmt.Printf("🚀 Modo Turbo ACTIVADO (workers: %d, calidad: baja)\n", recommendedWorkers)
		fmt.Println("💡 Se usará la calidad más baja para máxima velocidad")
	} else {
		fmt.Println("🐢 Modo Turbo DESACTIVADO (calidad: normal)")
	}

}

func toggleQualityMode() {
	qualityMode = !qualityMode
	if turboMode {
		turboMode = false
		fmt.Println("⚠️  Modo Turbo DESACTIVADO automáticamente al activar Modo Calidad")
	}

	if qualityMode {
		fmt.Printf("🚀 Modo Calidad ACTIVADO (workers: %d, calidad: Alta)\n", recommendedWorkers)
		fmt.Println("💡 Se usará la calidad más alta para maxima calidad")
	} else {
		fmt.Println("🐢 Modo Calidad DESACTIVADO (calidad: normal)")
	}
}

func toggleResetModes() {
	turboMode = false
	qualityMode = false
	audioQuality = "5"
	fmt.Println("⚙️  Modos Desactivados. Volviendo a la configuración normal.")
}

func main() {
	fmt.Println("🎵 LeuMusic Downloader - By Leuan")
	fmt.Println("==========================================")

	// Cargar canciones descargadas al inicio
	loadDownloadedSongs()

	checkRequiredTools()
	// Calcular workers recomendados al inicio
	recommendedWorkers = calculateRecommendedWorkers()

	for {
		fmt.Println("\nOpciones:")
		fmt.Println("-------------------- LeuMusic --------------------")
		fmt.Println("1. Descargar desde archivo canciones.txt")
		fmt.Println("2. Agregar canciones manualmente")
		fmt.Println("3. Uso de cookies (", useCookies, ")")
		fmt.Println("\n")

		fmt.Println("-------------------- By: Leuan --------------------")
		fmt.Println("5. Verificar herramientas")
		fmt.Println("6. Mostrar estructura de carpetas")
		fmt.Println("7. Ver canciones descargadas")
		fmt.Println("9. Salir")
		fmt.Println("\n")

		fmt.Println("-------------------- Rendimiento --------------------")
		fmt.Printf("🔧 Workers recomendados para tu PC: %d\n", recommendedWorkers)
		fmt.Println("10. Modo Turbo (", turboMode, ")")
		fmt.Println("11. Modo Calidad (", qualityMode, ")")
		fmt.Println("12. Restablecer Configuraciones Predeterminadas")

		fmt.Print("Selecciona: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 0:
			return
		// Basic //
		case 1:
			downloadFromFile()
		case 2:
			manualInput()
		case 3:
			toggleCookies()

		// Configuration //
		case 5:
			checkRequiredTools()
		case 6:
			showFolderStructure()
		case 7:
			showDownloadedSongs()

		// Rendimiento //
		case 10:
			toggleTurboMode()
		case 11:
			toggleQualityMode()
		case 12:
			toggleResetModes()

		default:
			fmt.Println("Opción inválida")
		}
	}
}

func toggleCookies() {
	useCookies = !useCookies
	fmt.Printf("🛡️  Uso de cookies: %v\n", useCookies)
	if useCookies {
		fmt.Println("💡 Se intentará usar cookies del navegador para evitar bloqueos")
	} else {
		fmt.Println("💡 Se desactivó el uso de cookies (puede haber más bloqueos)")
	}
}

func checkRequiredTools() {
	fmt.Println("\n🔍 Verificando herramientas necesarias...")

	tools := []struct {
		name string
		test string
	}{
		{"yt-dlp", "--version"},
		{"ffmpeg", "-version"},
	}

	allOk := true
	for _, tool := range tools {
		cmd := exec.Command("cmd", "/c", tool.name, tool.test)
		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ %s: NO INSTALADO\n", tool.name)
			allOk = false
		} else {
			fmt.Printf("✅ %s: OK\n", tool.name)
		}
	}

	if !allOk {
		fmt.Println("\n⚠️  Instala las herramientas faltantes:")
		fmt.Println("yt-dlp: pip install yt-dlp")
		fmt.Println("ffmpeg: winget install Gyan.FFmpeg")
	}
}

func loadDownloadedSongs() {
	downloadedMutex.Lock()
	defer downloadedMutex.Unlock()

	downloadedSongs = make(map[string]bool)
	file, err := os.Open("descargadas.txt")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			downloadedSongs[line] = true
		}
	}
}

func downloadFromFile() {
	file, err := os.Open("canciones.txt")
	if err != nil {
		fmt.Println("❌ No se encontró canciones.txt")
		fmt.Println("📝 Creando archivo de ejemplo...")
		createExampleFile()
		return
	}
	defer file.Close()

	downloadedMutex.RLock()
	loadedDownloadedSongs := make(map[string]bool)
	for k, v := range downloadedSongs {
		loadedDownloadedSongs[k] = v
	}
	downloadedMutex.RUnlock()

	var tasks []DownloadTask
	scanner := bufio.NewScanner(file)

	lineCount := 0
	validCount := 0
	skippedCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		task := parseLine(line)
		if task != nil {
			// Verificar si ya fue descargada
			key := fmt.Sprintf("%s - %s", task.Artist, task.Song)
			if loadedDownloadedSongs[key] {
				fmt.Printf("⏭️  Ya descargada: %s\n", key)
				skippedCount++
				continue
			}

			tasks = append(tasks, *task)
			validCount++
		} else {
			fmt.Printf("⚠️  Línea %d con formato inválido: %s\n", lineCount, line)
		}
	}

	if len(tasks) == 0 {
		fmt.Printf("🎯 No hay canciones nuevas para descargar. Saltadas: %d\n", skippedCount)
		return
	}

	fmt.Printf("🎶 Encontradas %d canciones nuevas (%d saltadas)\n", validCount, skippedCount)
	processDownloads(tasks)
}

func manualInput() {
	fmt.Println("\n📝 Ingresa canciones (formato: artista - canción)")
	fmt.Println("Ejemplo: pearl jam - even flow")
	fmt.Println("Escribe 'fin' para terminar")

	downloadedMutex.RLock()
	loadedDownloadedSongs := make(map[string]bool)
	for k, v := range downloadedSongs {
		loadedDownloadedSongs[k] = v
	}
	downloadedMutex.RUnlock()

	var tasks []DownloadTask
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if strings.ToLower(line) == "fin" {
			break
		}

		if line == "" {
			continue
		}

		task := parseLine(line)
		if task != nil {
			// Verificar duplicado
			key := fmt.Sprintf("%s - %s", task.Artist, task.Song)
			if loadedDownloadedSongs[key] {
				fmt.Printf("⏭️  Ya descargada: %s\n", key)
				continue
			}

			tasks = append(tasks, *task)
			fmt.Printf("✅ Agregado: %s - %s\n", task.Artist, task.Song)
		} else {
			fmt.Println("❌ Formato inválido. Usa: artista - canción")
		}
	}

	if len(tasks) > 0 {
		processDownloads(tasks)
	}
}

func parseLine(line string) *DownloadTask {
	separators := []string{" - ", " | ", " :: ", " -> "}

	for _, sep := range separators {
		if strings.Contains(line, sep) {
			parts := strings.SplitN(line, sep, 2)
			if len(parts) == 2 {
				artist := strings.TrimSpace(parts[0])
				song := strings.TrimSpace(parts[1])
				if artist != "" && song != "" {
					return &DownloadTask{Artist: artist, Song: song}
				}
			}
		}
	}

	return nil
}

func processDownloads(tasks []DownloadTask) {
	fmt.Printf("\n🚀 Iniciando descarga TURBO de %d canciones...\n", len(tasks))
	startTime := time.Now()

	var stats DownloadStats

	// Configuración TURBO - más goroutines para I/O bound
	concurrentWorkers := calculateOptimalWorkers()
	fmt.Printf("⚡ Usando %d workers concurrentes\n", concurrentWorkers)

	semaphore := make(chan struct{}, concurrentWorkers)
	var wg sync.WaitGroup

	// Canal para resultados
	results := make(chan bool, len(tasks))

	for i, task := range tasks {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(t DownloadTask, index int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			var success bool
			if t.URL != "" {
				// Descargar desde URL (playlist o artista)
				success = downloadFromURL(t.URL, t.Artist, index+1, len(tasks))
			} else {
				// Búsqueda normal
				success = downloadSongTurbo(t.Artist, t.Song, index+1, len(tasks))
			}

			results <- success

			if success {
				// Registrar en descargadas.txt
				if t.URL == "" {
					markAsDownloaded(t.Artist, t.Song)
				}
				atomic.AddInt32(&stats.success, 1)
			} else {
				atomic.AddInt32(&stats.failed, 1)
			}
		}(task, i)
	}

	// Mostrar progreso en tiempo real
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				success := atomic.LoadInt32(&stats.success)
				failed := atomic.LoadInt32(&stats.failed)
				completed := success + failed
				progress := float64(completed) / float64(len(tasks)) * 100
				fmt.Printf("\r📊 Progreso: %.1f%% (%d/%d) - ✅%d ❌%d",
					progress, completed, len(tasks), success, failed)
			case <-done:
				return
			}
		}
	}()

	wg.Wait()
	close(results)
	done <- true

	duration := time.Since(startTime)

	fmt.Printf("\n\n🎊 Descargas completadas en %v!\n", duration.Round(time.Second))
	fmt.Printf("✅ Éxitos: %d\n", stats.success)
	fmt.Printf("❌ Fallos: %d\n", stats.failed)

	if stats.success > 0 {
		fmt.Printf("📝 Registradas en: descargadas.txt\n")
		fmt.Println("\n📁 Estructura actual de carpetas:")
		showFolderStructure()
	}
}

func calculateOptimalWorkers() int {
	if turboMode {
		return recommendedWorkers
	}
	// Modo normal: más conservador
	cpus := runtime.NumCPU()
	workers := cpus * 2 // Reducido de 3 a 2
	if workers < 4 {
		workers = 4
	}
	if workers > 10 { // Reducido de 16 a 10
		workers = 10
	}
	return workers
}

func downloadSongTurbo(artist, song string, current, total int) bool {
	artistFolder := sanitizeFolderName(artist)

	if _, err := os.Stat(artistFolder); os.IsNotExist(err) {
		os.MkdirAll(artistFolder, 0755)
	}

	// Verificar duplicados en disco
	if isAlreadyDownloaded(artist, song, artistFolder) {
		fmt.Printf("   ⏭️  Ya existe en disco: %s - %s\n", artist, song)
		markAsDownloaded(artist, song)
		return true
	}

	searchQuery := fmt.Sprintf("%s %s", artist, song)
	outputTemplate := filepath.Join(artistFolder, "%(title)s.%(ext)s")

	return executeDownload(searchQuery, outputTemplate, "ytsearch1:")
}

func downloadFromURL(url, artist string, current, total int) bool {
	artistFolder := sanitizeFolderName(artist)

	if _, err := os.Stat(artistFolder); os.IsNotExist(err) {
		os.MkdirAll(artistFolder, 0755)
	}

	outputTemplate := filepath.Join(artistFolder, "%(title)s.%(ext)s")

	return executeDownload(url, outputTemplate, "")
}

func executeDownload(query, outputTemplate, searchPrefix string) bool {

	// Determinar calidad basada en modo
	audioQuality := "5"
	if turboMode {
		audioQuality = "9" // Calidad más baja, más rápido
	}
	if qualityMode {
		audioQuality = "1" // Calidad más baja, más rápido
	}

	args := []string{
		"/c", "yt-dlp",
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", audioQuality,
		"--embed-thumbnail",
		"--add-metadata",
		"--no-overwrites",
		"--no-playlist",
		"--socket-timeout", "30",
		"--retries", "3",
		"--fragment-retries", "3",
		"--ignore-errors", // Ignorar errores y continuar
		"--no-warnings",   // Suprimir warnings
		"--output", outputTemplate,
	}

	// Agregar cookies si está activado
	if useCookies {
		args = append(args, "--cookies-from-browser", "chrome")
	}

	// Agregar prefijo de búsqueda si es necesario
	if searchPrefix != "" {
		args = append(args, searchPrefix+query)
	} else {
		args = append(args, query)
	}

	cmd := exec.Command("cmd", args...)

	// Ejecutar y capturar output para debugging en caso de error
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Filtrar warnings comunes que no son errores fatales
		outputStr := string(output)

		// Si contiene "Finished downloading playlist", considerar como éxito
		if strings.Contains(outputStr, "Finished downloading playlist") {
			return true
		}

		// Ignorar warnings específicos
		if containsOnlyWarnings(outputStr) {
			return true
		}

		// Mostrar error real
		if len(outputStr) > 0 {
			// Extraer solo la parte importante del error
			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "ERROR:") {
					fmt.Printf("   🔥 Error: %s\n", strings.TrimPrefix(line, "ERROR:"))
					break
				}
			}
		}
		return false
	}

	return true
}

func containsOnlyWarnings(output string) bool {
	// Lista de warnings que podemos ignorar

	lines := strings.Split(output, "\n")
	hasRealError := false

	for _, line := range lines {
		if strings.HasPrefix(line, "ERROR:") {
			// Verificar si es un error de "bot" que podemos ignorar con cookies
			if strings.Contains(line, "Sign in to confirm you're not a bot") {
				if useCookies {
					// Con cookies activadas, este error no debería ocurrir
					hasRealError = true
				}
				// Sin cookies, podemos ignorarlo y continuar
				continue
			}
			hasRealError = true
			break
		}
	}

	return !hasRealError
}

func isAlreadyDownloaded(artist, song, folder string) bool {
	downloadedMutex.RLock()
	defer downloadedMutex.RUnlock()

	key := fmt.Sprintf("%s - %s", artist, song)
	if downloadedSongs[key] {
		return true
	}

	// También verificar en archivos existentes
	files, _ := filepath.Glob(filepath.Join(folder, "*.mp3"))
	cleanSong := strings.ToLower(song)

	for _, file := range files {
		filename := strings.ToLower(filepath.Base(file))
		if strings.Contains(filename, cleanSong) {
			return true
		}
	}

	return false
}

func markAsDownloaded(artist, song string) {
	entry := fmt.Sprintf("%s - %s\n", artist, song)

	downloadedMutex.Lock()
	defer downloadedMutex.Unlock()

	// Actualizar mapa en memoria
	downloadedSongs[entry] = true

	// Escribir en archivo
	file, err := os.OpenFile("descargadas.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error abriendo descargadas.txt: %v", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(entry); err != nil {
		log.Printf("Error escribiendo en descargadas.txt: %v", err)
	}
}

func showDownloadedSongs() {
	downloadedMutex.RLock()
	defer downloadedMutex.RUnlock()

	if len(downloadedSongs) == 0 {
		fmt.Println("📝 Aún no hay canciones descargadas")
		return
	}

	fmt.Println("\n📋 Canciones ya descargadas:")
	fmt.Println("==========================")

	count := 0
	for song := range downloadedSongs {
		fmt.Printf("✅ %s\n", song)
		count++
	}

	fmt.Printf("\nTotal: %d canciones descargadas\n", count)
}

func sanitizeFolderName(name string) string {
	// Eliminar caracteres inválidos para nombres de carpeta
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = reg.ReplaceAllString(name, "_")

	// También eliminar puntos al final y espacios extra
	name = strings.TrimRight(name, ". ")
	name = strings.TrimSpace(name)

	// Limitar longitud
	if len(name) > 50 {
		name = name[:50]
	}

	if name == "" {
		name = "CarpetaSinNombre"
	}

	return name
}

func showFolderStructure() {
	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("❌ Error leyendo directorio:", err)
		return
	}

	foundFolders := false
	totalSongs := 0

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			foundFolders = true
			mp3Files, _ := filepath.Glob(filepath.Join(entry.Name(), "*.mp3"))
			fmt.Printf("📁 %s/ (%d canciones)\n", entry.Name(), len(mp3Files))
			totalSongs += len(mp3Files)

			for i, file := range mp3Files {
				if i < 2 { // Mostrar solo 2 canciones por carpeta
					fmt.Printf("   🎵 %s\n", filepath.Base(file))
				} else if i == 2 && len(mp3Files) > 3 {
					fmt.Printf("   ... y %d más\n", len(mp3Files)-2)
					break
				}
			}
		}
	}

	if foundFolders {
		fmt.Printf("\n🎯 Total de canciones en disco: %d\n", totalSongs)
	} else {
		fmt.Println("   (no hay carpetas de artistas aún)")
	}
}

func createExampleFile() {
	exampleContent := `# Archivo de canciones para descargar
# Formato: artista - canción

pearl jam - even flow
justin bieber - baby
queen - bohemian rhapsody
metallica - enter sandman
adele - rolling in the deep
coldplay - yellow
ed sheeran - shape of you
taylor swift - shake it off
nirvana - smells like teen spirit
oasis - wonderwall`

	err := os.WriteFile("canciones.txt", []byte(exampleContent), 0644)
	if err != nil {
		fmt.Println("Error creando archivo de ejemplo:", err)
		return
	}

	fmt.Println("✅ Archivo canciones.txt creado con ejemplos")
}
