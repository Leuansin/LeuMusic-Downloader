package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("=== TXT Files Combiner ===")

	// Get current directory
	exeDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error getting directory: %v\n", err)
		return
	}

	fmt.Printf("📁 Current directory: %s\n", exeDir)

	// Create output file
	outFile, err := os.Create("rename me to songs.txt.txt")
	if err != nil {
		fmt.Printf("❌ Error creating output file: %v\n", err)
		return
	}
	defer outFile.Close()

	// Read directory files
	files, err := os.ReadDir(exeDir)
	if err != nil {
		fmt.Printf("❌ Error reading directory: %v\n", err)
		return
	}

	txtCount := 0
	exeCount := 0

	// Count files first
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".txt" && strings.ToLower(file.Name()) != "rename me to songs.txt.txt" {
			txtCount++
		} else if ext == ".exe" {
			exeCount++
		}
	}

	fmt.Printf("🔍 Found %d .exe files\n", exeCount)
	fmt.Printf("📄 Found %d .txt files to process\n", txtCount)

	if txtCount == 0 {
		fmt.Println("⚠️  No .txt files found to process")
		return
	}

	// Process each file
	processed := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".txt" || strings.ToLower(file.Name()) == "rename me to songs.txt.txt" {
			continue
		}

		// Open input file
		inFile, err := os.Open(file.Name())
		if err != nil {
			fmt.Printf("❌ Error opening %s: %v\n", file.Name(), err)
			continue
		}

		// Copy content
		_, err = io.Copy(outFile, inFile)
		inFile.Close()
		if err != nil {
			fmt.Printf("❌ Error copying %s: %v\n", file.Name(), err)
			continue
		}

		processed++
		fmt.Printf("✅ Processed: %s\n", file.Name())
	}

	fmt.Printf("\n🎉 Process completed! ✔️\n")
	fmt.Printf("📊 Files processed: %d of %d\n", processed, txtCount)
	fmt.Printf("💾 Output file: rename me to songs.txt.txt\n")
	fmt.Println("✨ Done! ¬¬")
}
