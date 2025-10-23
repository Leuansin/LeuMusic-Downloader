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
	URL    string
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
	useCookies         bool = false
	turboMode          bool = false
	qualityMode        bool = false
	audioQuality       string
	recommendedWorkers int
	downloadedSongs    map[string]bool
	downloadedMutex    sync.RWMutex
)

func calculateRecommendedWorkers() int {
	cpus := runtime.NumCPU()
	memoryGB := getMemoryGB()

	workers := cpus * 2

	if memoryGB >= 16 {
		workers = cpus * 3
	} else if memoryGB >= 8 {
		workers = cpus * 2
	} else {
		workers = cpus
	}

	if workers < 4 {
		workers = 4
	}
	if workers > 12 {
		workers = 12
	}

	return workers
}

func getMemoryGB() uint64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return uint64(mem.Sys / 1024 / 1024 / 1024)
}

func toggleTurboMode() {
	turboMode = !turboMode
	if qualityMode {
		qualityMode = false
		fmt.Println("⚠️  Quality Mode automatically disabled when enabling Turbo Mode")
	}

	if turboMode {
		fmt.Printf("🚀 Turbo Mode ACTIVATED (workers: %d, quality: low)\n", recommendedWorkers)
		fmt.Println("💡 Using lowest quality for maximum speed")
	} else {
		fmt.Println("🐢 Turbo Mode DEACTIVATED (quality: normal)")
	}
}

func toggleQualityMode() {
	qualityMode = !qualityMode
	if turboMode {
		turboMode = false
		fmt.Println("⚠️  Turbo Mode automatically disabled when enabling Quality Mode")
	}

	if qualityMode {
		fmt.Printf("🎵 Quality Mode ACTIVATED (workers: %d, quality: high)\n", recommendedWorkers)
		fmt.Println("💡 Using highest quality for best audio")
	} else {
		fmt.Println("🐢 Quality Mode DEACTIVATED (quality: normal)")
	}
}

func toggleResetModes() {
	turboMode = false
	qualityMode = false
	audioQuality = "5"
	fmt.Println("⚙️  All modes deactivated. Returning to normal configuration.")
}

func main() {
	fmt.Println("🎵 LeuMusic Downloader - By Leuan")
	fmt.Println("==========================================")

	loadDownloadedSongs()
	checkRequiredTools()
	recommendedWorkers = calculateRecommendedWorkers()

	for {
		fmt.Println("\nOptions:")
		fmt.Println("-------------------- LeuMusic --------------------")
		fmt.Println("1. Download from songs.txt file")
		fmt.Println("2. Add songs manually")
		fmt.Println("3. Use cookies (", useCookies, ")")
		fmt.Println("\n")

		fmt.Println("-------------------- By: Leuan --------------------")
		fmt.Println("5. Check tools")
		fmt.Println("6. Show folder structure")
		fmt.Println("7. View downloaded songs")
		fmt.Println("9. Exit")
		fmt.Println("\n")

		fmt.Println("-------------------- Performance --------------------")
		fmt.Printf("🔧 Recommended workers for your PC: %d\n", recommendedWorkers)
		fmt.Println("10. Turbo Mode (", turboMode, ")")
		fmt.Println("11. Quality Mode (", qualityMode, ")")
		fmt.Println("12. Reset to Default Settings")

		fmt.Print("Select: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 0:
			return
		case 1:
			downloadFromFile()
		case 2:
			manualInput()
		case 3:
			toggleCookies()
		case 5:
			checkRequiredTools()
		case 6:
			showFolderStructure()
		case 7:
			showDownloadedSongs()
		case 10:
			toggleTurboMode()
		case 11:
			toggleQualityMode()
		case 12:
			toggleResetModes()
		default:
			fmt.Println("Invalid option")
		}
	}
}

func toggleCookies() {
	useCookies = !useCookies
	fmt.Printf("🛡️  Use cookies: %v\n", useCookies)
	if useCookies {
		fmt.Println("💡 Will try to use browser cookies to avoid blocks")
	} else {
		fmt.Println("💡 Cookies deactivated (may have more blocks)")
	}
}

func checkRequiredTools() {
	fmt.Println("\n🔍 Checking required tools...")

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
			fmt.Printf("❌ %s: NOT INSTALLED\n", tool.name)
			allOk = false
		} else {
			fmt.Printf("✅ %s: OK\n", tool.name)
		}
	}

	if !allOk {
		fmt.Println("\n⚠️  Install missing tools:")
		fmt.Println("yt-dlp: pip install yt-dlp")
		fmt.Println("ffmpeg: winget install Gyan.FFmpeg")
	}
}

func loadDownloadedSongs() {
	downloadedMutex.Lock()
	defer downloadedMutex.Unlock()

	downloadedSongs = make(map[string]bool)
	file, err := os.Open("downloaded.txt")
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
	file, err := os.Open("songs.txt")
	if err != nil {
		fmt.Println("❌ songs.txt file not found")
		fmt.Println("📝 Creating example file...")
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
			key := fmt.Sprintf("%s - %s", task.Artist, task.Song)
			if loadedDownloadedSongs[key] {
				fmt.Printf("⏭️  Already downloaded: %s\n", key)
				skippedCount++
				continue
			}

			tasks = append(tasks, *task)
			validCount++
		} else {
			fmt.Printf("⚠️  Line %d has invalid format: %s\n", lineCount, line)
		}
	}

	if len(tasks) == 0 {
		fmt.Printf("🎯 No new songs to download. Skipped: %d\n", skippedCount)
		return
	}

	fmt.Printf("🎶 Found %d new songs (%d skipped)\n", validCount, skippedCount)
	processDownloads(tasks)
}

func manualInput() {
	fmt.Println("\n📝 Enter songs (format: artist - song)")
	fmt.Println("Example: pearl jam - even flow")
	fmt.Println("Type 'fin' to finish")

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
			key := fmt.Sprintf("%s - %s", task.Artist, task.Song)
			if loadedDownloadedSongs[key] {
				fmt.Printf("⏭️  Already downloaded: %s\n", key)
				continue
			}

			tasks = append(tasks, *task)
			fmt.Printf("✅ Added: %s - %s\n", task.Artist, task.Song)
		} else {
			fmt.Println("❌ Invalid format. Use: artist - song")
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
	fmt.Printf("\n🚀 Starting TURBO download of %d songs...\n", len(tasks))
	startTime := time.Now()

	var stats DownloadStats

	concurrentWorkers := calculateOptimalWorkers()
	fmt.Printf("⚡ Using %d concurrent workers\n", concurrentWorkers)

	semaphore := make(chan struct{}, concurrentWorkers)
	var wg sync.WaitGroup

	results := make(chan bool, len(tasks))

	for i, task := range tasks {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(t DownloadTask, index int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			var success bool
			if t.URL != "" {
				success = downloadFromURL(t.URL, t.Artist, index+1, len(tasks))
			} else {
				success = downloadSongTurbo(t.Artist, t.Song, index+1, len(tasks))
			}

			results <- success

			if success {
				if t.URL == "" {
					markAsDownloaded(t.Artist, t.Song)
				}
				atomic.AddInt32(&stats.success, 1)
			} else {
				atomic.AddInt32(&stats.failed, 1)
			}
		}(task, i)
	}

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
				fmt.Printf("\r📊 Progress: %.1f%% (%d/%d) - ✅%d ❌%d",
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

	fmt.Printf("\n\n🎊 Downloads completed in %v!\n", duration.Round(time.Second))
	fmt.Printf("✅ Success: %d\n", stats.success)
	fmt.Printf("❌ Failed: %d\n", stats.failed)

	if stats.success > 0 {
		fmt.Printf("📝 Registered in: downloaded.txt\n")
		fmt.Println("\n📁 Current folder structure:")
		showFolderStructure()
	}
}

func calculateOptimalWorkers() int {
	if turboMode {
		return recommendedWorkers
	}
	cpus := runtime.NumCPU()
	workers := cpus * 2
	if workers < 4 {
		workers = 4
	}
	if workers > 10 {
		workers = 10
	}
	return workers
}

func downloadSongTurbo(artist, song string, current, total int) bool {
	artistFolder := sanitizeFolderName(artist)

	if _, err := os.Stat(artistFolder); os.IsNotExist(err) {
		os.MkdirAll(artistFolder, 0755)
	}

	if isAlreadyDownloaded(artist, song, artistFolder) {
		fmt.Printf("   ⏭️  Already exists on disk: %s - %s\n", artist, song)
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
	audioQuality := "5"
	if turboMode {
		audioQuality = "9"
	}
	if qualityMode {
		audioQuality = "1"
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
		"--ignore-errors",
		"--no-warnings",
		"--output", outputTemplate,
	}

	if useCookies {
		args = append(args, "--cookies-from-browser", "chrome")
	}

	if searchPrefix != "" {
		args = append(args, searchPrefix+query)
	} else {
		args = append(args, query)
	}

	cmd := exec.Command("cmd", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)

		if strings.Contains(outputStr, "Finished downloading playlist") {
			return true
		}

		if containsOnlyWarnings(outputStr) {
			return true
		}

		if len(outputStr) > 0 {
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
	lines := strings.Split(output, "\n")
	hasRealError := false

	for _, line := range lines {
		if strings.HasPrefix(line, "ERROR:") {
			if strings.Contains(line, "Sign in to confirm you're not a bot") {
				if useCookies {
					hasRealError = true
				}
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

	downloadedSongs[entry] = true

	file, err := os.OpenFile("downloaded.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening downloaded.txt: %v", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(entry); err != nil {
		log.Printf("Error writing to downloaded.txt: %v", err)
	}
}

func showDownloadedSongs() {
	downloadedMutex.RLock()
	defer downloadedMutex.RUnlock()

	if len(downloadedSongs) == 0 {
		fmt.Println("📝 No songs downloaded yet")
		return
	}

	fmt.Println("\n📋 Already downloaded songs:")
	fmt.Println("==========================")

	count := 0
	for song := range downloadedSongs {
		fmt.Printf("✅ %s\n", song)
		count++
	}

	fmt.Printf("\nTotal: %d songs downloaded\n", count)
}

func sanitizeFolderName(name string) string {
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = reg.ReplaceAllString(name, "_")
	name = strings.TrimRight(name, ". ")
	name = strings.TrimSpace(name)

	if len(name) > 50 {
		name = name[:50]
	}

	if name == "" {
		name = "UnnamedFolder"
	}

	return name
}

func showFolderStructure() {
	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("❌ Error reading directory:", err)
		return
	}

	foundFolders := false
	totalSongs := 0

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			foundFolders = true
			mp3Files, _ := filepath.Glob(filepath.Join(entry.Name(), "*.mp3"))
			fmt.Printf("📁 %s/ (%d songs)\n", entry.Name(), len(mp3Files))
			totalSongs += len(mp3Files)

			for i, file := range mp3Files {
				if i < 2 {
					fmt.Printf("   🎵 %s\n", filepath.Base(file))
				} else if i == 2 && len(mp3Files) > 3 {
					fmt.Printf("   ... and %d more\n", len(mp3Files)-2)
					break
				}
			}
		}
	}

	if foundFolders {
		fmt.Printf("\n🎯 Total songs on disk: %d\n", totalSongs)
	} else {
		fmt.Println("   (no artist folders yet)")
	}
}

func createExampleFile() {
	exampleContent := `# Songs file to download
# Format: artist - song

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

	err := os.WriteFile("songs.txt", []byte(exampleContent), 0644)
	if err != nil {
		fmt.Println("Error creating example file:", err)
		return
	}

	fmt.Println("✅ songs.txt file created with examples")
}
