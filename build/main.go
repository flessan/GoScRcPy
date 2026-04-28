package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Mapping lengkap sesuai aset GitHub Scrcpy v3.3.4
var downloadURLs = map[string]string{
	"windows-amd64": "https://github.com/Genymobile/scrcpy/releases/download/v3.3.4/scrcpy-win64-v3.3.4.zip",
	"windows-386":   "https://github.com/Genymobile/scrcpy/releases/download/v3.3.4/scrcpy-win32-v3.3.4.zip",
	"windows-arm64": "https://github.com/Genymobile/scrcpy/releases/download/v3.3.4/scrcpy-win64-v3.3.4.zip", // Scrcpy win64 biasanya support emulation di arm64
	"linux-amd64":   "https://github.com/Genymobile/scrcpy/releases/download/v3.3.4/scrcpy-linux-x86_64-v3.3.4.tar.gz",
	"darwin-amd64":  "https://github.com/Genymobile/scrcpy/releases/download/v3.3.4/scrcpy-macos-x86_64-v3.3.4.tar.gz",
	"darwin-arm64":  "https://github.com/Genymobile/scrcpy/releases/download/v3.3.4/scrcpy-macos-aarch64-v3.3.4.tar.gz",
}

const scrcpyFolder = "scrcpy_core"

func main() {
	osType := runtime.GOOS
	archType := runtime.GOARCH
	platformKey := fmt.Sprintf("%s-%s", osType, archType)

	fmt.Printf("GoScRcPy! v1\n")
	fmt.Printf("Platform Detected: %s\n", platformKey)

	baseDir, _ := os.Getwd()
	targetDir := filepath.Join(baseDir, scrcpyFolder)

	// 1. Logic Pencarian & Download
	scrcpyPath := findScrcpyFolder(targetDir)
	if scrcpyPath == "" {
		url, ok := downloadURLs[platformKey]
		if !ok {
			fmt.Printf("sorry! thi plat %s not supported, automaticaly.\n", platformKey)
			return
		}

		fmt.Printf("downloading engine for %s...\n", platformKey)
		err := downloadAndSetup(url, osType)
		if err != nil {
			fmt.Printf("we have a malfunction setup, sir!: %v\n", err)
			return
		}
		scrcpyPath = findScrcpyFolder(targetDir)
	}

	// 2. Eksekusi ADB & Scrcpy
	fmt.Println("searching da device...")
	out := runCmd(scrcpyPath, "adb", "devices")
	fmt.Print(out)

	if strings.Contains(out, "\tdevice") {
		fmt.Println("device connected!! running Scrcpy...")
		go runCmd(scrcpyPath, "scrcpy", "--always-on-top", "--window-title", "GoScrcpy-Universal")
	} else {
		fmt.Println("umm, device not found!! try to check da US-Bee")
	}

	time.Sleep(5 * time.Second)
}

// --- FUNGSI HELPER (SAMA SEPERTI SEBELUMNYA TAPI LEBIH SOLID) ---

func findScrcpyFolder(root string) string {
	files, err := os.ReadDir(root)
	if err != nil {
		return ""
	}
	for _, f := range files {
		if f.IsDir() && (strings.Contains(f.Name(), "scrcpy-") || strings.Contains(f.Name(), "bin")) {
			return filepath.Join(root, f.Name())
		}
	}
	return ""
}

func runCmd(dir, name string, args ...string) string {
	fullName := name
	if runtime.GOOS == "windows" {
		fullName += ".exe"
	}
	// Perbaikan untuk Linux/macOS path
	path := filepath.Join(dir, fullName)
	
	// Jika di Linux/Mac, berikan izin eksekusi (chmod +x) secara otomatis
	if runtime.GOOS != "windows" {
		os.Chmod(path, 0755)
	}

	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func downloadAndSetup(url string, osType string) error {
	ext := ".tar.gz"
	if osType == "windows" {
		ext = ".zip"
	}
	tempFile := "download_temp" + ext

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, _ := os.Create(tempFile)
	io.Copy(f, resp.Body)
	f.Close()

	fmt.Println("Mengekstrak aset...")
	if osType == "windows" {
		err = unzip(tempFile, scrcpyFolder)
	} else {
		err = untar(tempFile, scrcpyFolder)
	}
	os.Remove(tempFile)
	return err
}

func unzip(src, dest string) error {
	r, _ := zip.OpenReader(src)
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
		out, _ := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		rc, _ := f.Open()
		io.Copy(out, rc)
		out.Close()
		rc.Close()
	}
	return nil
}

func untar(src, dest string) error {
	f, _ := os.Open(src)
	defer f.Close()
	gzr, _ := gzip.NewReader(f)
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF { break }
		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			f, _ := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			io.Copy(f, tr)
			f.Close()
		}
	}
	return nil
}