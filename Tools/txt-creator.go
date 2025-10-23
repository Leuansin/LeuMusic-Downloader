package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("=== Combinador de archivos TXT ===")

	// Obtener el directorio del ejecutable
	exeDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error al obtener directorio: %v\n", err)
		return
	}

	fmt.Printf("📁 Directorio actual: %s\n", exeDir)

	// Abrir archivo de salida
	outFile, err := os.Create("renombrame a canciones.txt.txt")
	if err != nil {
		fmt.Printf("❌ Error al crear archivo de salida: %v\n", err)
		return
	}
	defer outFile.Close()

	// Leer archivos del directorio
	files, err := os.ReadDir(exeDir)
	if err != nil {
		fmt.Printf("❌ Error al leer directorio: %v\n", err)
		return
	}

	txtCount := 0
	exeCount := 0

	// Contar archivos primero
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".txt" && strings.ToLower(file.Name()) != "renombrame a canciones.txt.txt" {
			txtCount++
		} else if ext == ".exe" {
			exeCount++
		}
	}

	fmt.Printf("🔍 Se encontraron %d archivos .exe\n", exeCount)
	fmt.Printf("📄 Se encontraron %d archivos .txt para procesar\n", txtCount)

	if txtCount == 0 {
		fmt.Println("⚠️  No se encontraron archivos .txt para procesar")
		return
	}

	// Procesar cada archivo
	processed := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".txt" || strings.ToLower(file.Name()) == "renombrame a canciones.txt.txt" {
			continue
		}

		// Abrir archivo de entrada
		inFile, err := os.Open(file.Name())
		if err != nil {
			fmt.Printf("❌ Error al abrir %s: %v\n", file.Name(), err)
			continue
		}

		// Copiar contenido
		_, err = io.Copy(outFile, inFile)
		inFile.Close()
		if err != nil {
			fmt.Printf("❌ Error al copiar %s: %v\n", file.Name(), err)
			continue
		}

		processed++
		fmt.Printf("✅ Procesado: %s\n", file.Name())
	}

	fmt.Printf("\n🎉 ¡Proceso completado! ✔️\n")
	fmt.Printf("📊 Archivos procesados: %d de %d\n", processed, txtCount)
	fmt.Printf("💾 Archivo generado: renombrame a canciones.txt.txt\n")
	fmt.Println("✨ ¡Listo! ¬¬")
}
