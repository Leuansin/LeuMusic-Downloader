# <div align="center"> 🎵 LeuMusic Downloader 🎵 </div>
<div align="center">

**Fast, efficient music downloader written in Go**

*Descargador de música rápido y eficiente escrito en Go*

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey)
![Go](https://img.shields.io/badge/Go-1.19+-blue)

</div>

## 🌟 About | Acerca de

**LeuMusic Downloader** is a powerful and fast music downloader that organizes your music by artist folders. It uses yt-dlp to download songs from YouTube and offers both Turbo Mode for speed and Quality Mode for the best audio.

**LeuMusic Downloader** es un descargador de música potente y rápido que organiza tu música en carpetas por artista. Utiliza yt-dlp para descargar canciones de YouTube y ofrece tanto Modo Turbo para velocidad como Modo Calidad para el mejor audio.

---

## 🚀 Quick Start | Inicio Rápido

### English Version

**Prerequisites:**
- Install Go (version 1.19 or higher)
- Install yt-dlp and FFmpeg

**Download and run:**
```bash
# Clone the repository
git clone https://github.com/leuan/leumusic-downloader.git
cd leumusic-downloader

# Build the executable
go build -o leumusic.exe

# Run the program
./leumusic.exe
```

### Versión en Español

**Prerrequisitos:**
- Instala Go (versión 1.19 o superior)
- Instala yt-dlp y FFmpeg

**Descarga y ejecución:**
```bash
# Clona el repositorio
git clone https://github.com/leuan/leumusic-downloader.git
cd leumusic-downloader

# Construye el ejecutable
go build -o leumusic.exe

# Ejecuta el programa
./leumusic.exe
```

---

## 🛡️ Features | Características

### Download Features | Características de Descarga

| Feature | Description | Descripción |
|---|---:|---|
| Fast Downloads | Multiple concurrent downloads using goroutines | Descargas concurrentes múltiples usando goroutines |
| Artist Folder Organization | Automatically organizes songs into artist folders | Organiza automáticamente las canciones en carpetas de artistas |
| Duplicate Checking | Prevents downloading the same song twice | Evita descargar la misma canción dos veces |
| Turbo Mode | Lower quality for faster downloads | Calidad más baja para descargas más rápidas |
| Quality Mode | Higher quality for best audio experience | Calidad más alta para la mejor experiencia de audio |
| Cookie Support | Use browser cookies to avoid blocks | Usa cookies del navegador para evitar bloqueos |
| Playlist Support | Download entire YouTube playlists | Descarga listas de reproducción completas de YouTube |

### ⚡ Technical Features | Características Técnicas

| Feature | Description | Descripción |
|---|---:|---|
| Written in Go | Fast and efficient concurrent programming | Escrito en Go para programación concurrente rápida y eficiente |
| Cross-Platform | Works on Windows, Linux, and macOS | Funciona en Windows, Linux y macOS |
| No Dependencies | Standalone executable after compilation | Ejecutable independiente después de la compilación |
| Easy to Use | Simple console interface | Interfaz de consola simple |

---

## 📖 Usage Examples | Ejemplos de Uso

### Basic Usage | Uso Básico

**English Version:**
1. Run the program
2. Choose option 1 to download from songs.txt
3. Program will create an example songs.txt if it doesn't exist
4. Edit songs.txt with your songs in the format: artist - song
5. Run again and enjoy!

**Versión en Español:**
1. Ejecuta el programa
2. Elige la opción 1 para descargar desde songs.txt
3. El programa creará un archivo de ejemplo songs.txt si no existe
4. Edita songs.txt con tus canciones en el formato: artista - canción
5. ¡Ejecuta de nuevo y disfruta!

### Advanced Usage | Uso Avanzado

- **Turbo Mode:** Enable Turbo Mode (option 10) for fastest downloads using more workers with lower quality.
- **Quality Mode:** Enable Quality Mode (option 11) for highest audio quality.

---

## 🔧 Installation | Instalación

### Method 1: From Source | Método 1: Desde el Código Fuente
- Install Go 1.19+ | Instala Go 1.19+
- Clone the repository | Clona el repositorio
- Build the executable | Construye el ejecutable
- Install yt-dlp and FFmpeg | Instala yt-dlp y FFmpeg

### Method 2: Pre-built Binary | Método 2: Binario Precompilado
- Download the latest release from the releases page | Descarga la última versión desde la página de releases

---

## 🎯 Use Cases | Casos de Uso

- Music Lovers - Build your local music library | Amantes de la música - Construye tu biblioteca de música local
- Offline Listening - Download songs for offline listening | Escucha sin conexión - Descarga canciones para escuchar sin internet
- Organization - Keep your music organized by artist | Organización - Mantén tu música organizada por artista

---

## ⚠️ Important Notes | Notas Importantes

**Legal Disclaimer | Aviso Legal:**
This tool is intended for personal use and educational purposes only. Users are responsible for complying with all applicable laws and regulations regarding copyright and music distribution.

Esta herramienta está destinada solo para uso personal y fines educativos. Los usuarios son responsables de cumplir con todas las leyes y regulaciones aplicables sobre derechos de autor y distribución de música.

**Technical Notes | Notas Técnicas:**
- ✅ Requires yt-dlp and FFmpeg to be installed | Requiere yt-dlp y FFmpeg instalados
- ✅ Cookies support to avoid YouTube blocks | Soporte de cookies para evitar bloqueos de YouTube
- ✅ Automatic duplicate checking | Verificación automática de duplicados

---

## 🤝 Contributing | Contribuyendo

- Fork the project | Haz fork del proyecto
- Create your feature branch | Crea tu rama de características
- Commit your changes | Haz commit de tus cambios
- Push to the branch | Push a la rama
- Open a Pull Request | Abre un Pull Request

---

## 📄 License | Licencia

MIT License | Licencia MIT. All credits to Leuan | Todos los créditos a Leuan.

---

## 🐛 Reporting Issues | Reportar Problemas

Found a bug? Have a feature request? | ¿Encontraste un bug? ¿Tienes una solicitud de característica?
- Open an issue on GitHub: https://github.com/leuan/leumusic-downloader/issues

---

## 📞 Support | Soporte

- Discord: leuan

Made with ❤️ for the music & go community | Hecho con ❤️ para la comunidad de música y go

