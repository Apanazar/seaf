package ui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"seaf/archiver"
)

type GUI struct {
	app                    fyne.App
	window                 fyne.Window
	mainTabs               *container.AppTabs
	passwordEntry          *widget.Entry
	saltEntry              *widget.Entry
	filesList              *widget.List
	outputEntry            *widget.Entry
	generateSaltBtn        *widget.Button
	createArchiveBtn       *widget.Button
	selectedFiles          []string
	extractPasswordEntry   *widget.Entry
	extractSaltEntry       *widget.Entry
	archivePathLabel       *widget.Label
	extractBtn             *widget.Button
	selectedArchive        string
	progressBar            *widget.ProgressBar
	progressLabel          *widget.Label
	resultsText            *widget.Label
	resultsContainer       *fyne.Container
	closeResultsBtn        *widget.Button
	compressionLevelSelect *widget.Select
}

type Statistics struct {
	OriginalSize     int64
	CompressedSize   int64
	EncryptedSize    int64
	Entropy          float64
	CompressionRatio float64
	FileStats        []FileStat
}

type FileStat struct {
	Filename         string
	OriginalSize     int64
	CompressedSize   int64
	EncryptedSize    int64
	Entropy          float64
	CompressionRatio float64
}

func NewGUI() *GUI {
	gui := &GUI{
		app:           app.NewWithID("seaf.archiver"),
		selectedFiles: make([]string, 0),
	}

	gui.window = gui.app.NewWindow("S.E.A.F. - Secure Encrypted Archive Format")
	gui.window.Resize(fyne.NewSize(900, 700))

	gui.createUI()
	return gui
}

func (g *GUI) createUI() {
	createTab := g.createArchiveTab()
	extractTab := g.createExtractTab()

	g.progressBar = widget.NewProgressBar()
	g.progressBar.Hide()
	g.progressLabel = widget.NewLabel("")
	g.progressLabel.Hide()

	g.resultsText = widget.NewLabel("")
	g.resultsText.Wrapping = fyne.TextWrapWord

	resultsScroll := container.NewScroll(g.resultsText)
	resultsScroll.SetMinSize(fyne.NewSize(0, 300))

	g.closeResultsBtn = widget.NewButton("Close", g.clearResults)
	g.closeResultsBtn.Importance = widget.LowImportance
	g.closeResultsBtn.Hide()

	resultsHeader := container.NewHBox(
		widget.NewLabel("Compression Results"),
		container.NewHBox(),
		g.closeResultsBtn,
	)

	g.resultsContainer = container.NewVBox(
		resultsHeader,
		widget.NewSeparator(),
		resultsScroll,
	)
	g.resultsContainer.Hide()

	progressContainer := container.NewVBox(
		g.progressBar,
		g.progressLabel,
	)

	g.mainTabs = container.NewAppTabs(
		container.NewTabItem("Create Archive", createTab),
		container.NewTabItem("Extract Archive", extractTab),
	)

	mainContent := container.NewBorder(
		nil,
		container.NewVBox(
			progressContainer,
			g.resultsContainer,
		),
		nil,
		nil,
		g.mainTabs,
	)

	g.window.SetContent(mainContent)
}

func (g *GUI) createArchiveTab() *container.Scroll {
	g.passwordEntry = widget.NewPasswordEntry()
	g.passwordEntry.SetPlaceHolder("Enter encryption password")

	g.saltEntry = widget.NewEntry()
	g.saltEntry.SetPlaceHolder("Hex salt or generate new")

	g.generateSaltBtn = widget.NewButton("Generate Salt", g.generateSalt)

	saltContainer := container.NewBorder(
		nil, nil, nil, g.generateSaltBtn, g.saltEntry,
	)

	compressionLevels := []string{
		"0 - No compression",
		"1 - Fastest",
		"2", "3", "4", "5",
		"6 - Default",
		"7", "8",
		"9 - Best compression",
	}

	g.compressionLevelSelect = widget.NewSelect(compressionLevels, func(selected string) {})
	g.compressionLevelSelect.SetSelected("6 - Default")
	g.compressionLevelSelect.PlaceHolder = "Select compression level"

	g.filesList = widget.NewList(
		func() int {
			return len(g.selectedFiles)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(filepath.Base(g.selectedFiles[i]))
		},
	)

	selectFilesBtn := widget.NewButton("Select Files", g.selectFiles)
	clearFilesBtn := widget.NewButton("Clear All", g.clearFiles)

	fileButtons := container.NewHBox(selectFilesBtn, clearFilesBtn)

	g.outputEntry = widget.NewEntry()
	g.outputEntry.SetText("archive.seaf")
	g.outputEntry.SetPlaceHolder("Output filename")

	g.createArchiveBtn = widget.NewButton("Create Archive", g.createArchive)
	g.createArchiveBtn.Importance = widget.HighImportance

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Password", Widget: g.passwordEntry},
			{Text: "Salt", Widget: saltContainer},
			{Text: "Compression Level", Widget: g.compressionLevelSelect},
			{Text: "Files", Widget: container.NewBorder(
				nil, fileButtons, nil, nil, g.filesList,
			)},
			{Text: "Output File", Widget: g.outputEntry},
		},
	}

	return container.NewScroll(container.NewVBox(
		widget.NewLabel("Create encrypted and compressed archive"),
		widget.NewSeparator(),
		form,
		g.createArchiveBtn,
	))
}

func (g *GUI) createExtractTab() *container.Scroll {
	g.extractPasswordEntry = widget.NewPasswordEntry()
	g.extractPasswordEntry.SetPlaceHolder("Enter decryption password")

	g.extractSaltEntry = widget.NewEntry()
	g.extractSaltEntry.SetPlaceHolder("Hex salt used for encryption")

	g.archivePathLabel = widget.NewLabel("No archive selected")
	g.archivePathLabel.Wrapping = fyne.TextWrapWord

	selectArchiveBtn := widget.NewButton("Select Archive", g.selectArchive)

	g.extractBtn = widget.NewButton("Extract Archive", g.extractArchive)
	g.extractBtn.Importance = widget.HighImportance

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Password", Widget: g.extractPasswordEntry},
			{Text: "Salt", Widget: g.extractSaltEntry},
			{Text: "Archive File", Widget: container.NewVBox(
				g.archivePathLabel,
				selectArchiveBtn,
			)},
		},
	}

	return container.NewScroll(container.NewVBox(
		widget.NewLabel("Extract files from encrypted archive"),
		widget.NewSeparator(),
		form,
		g.extractBtn,
	))
}

func (g *GUI) generateSalt() {
	salt, err := generateRandomSalt(16)
	if err != nil {
		dialog.ShowError(err, g.window)
		return
	}

	g.saltEntry.SetText(salt)
	g.extractSaltEntry.SetText(salt)
}

func (g *GUI) selectFiles() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}
		if reader == nil {
			return
		}

		fileURI := reader.URI()
		filePath := fileURI.Path()

		for _, existingFile := range g.selectedFiles {
			if existingFile == filePath {
				reader.Close()
				return
			}
		}

		g.selectedFiles = append(g.selectedFiles, filePath)
		g.filesList.Refresh()
		reader.Close()
	}, g.window)
}

func (g *GUI) clearFiles() {
	g.selectedFiles = make([]string, 0)
	g.filesList.Refresh()
	g.clearResults()
}

func (g *GUI) selectArchive() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}
		if reader == nil {
			return
		}

		fileURI := reader.URI()
		g.selectedArchive = fileURI.Path()
		g.archivePathLabel.SetText(filepath.Base(g.selectedArchive))
		reader.Close()
	}, g.window)
}

func (g *GUI) createArchive() {
	if g.passwordEntry.Text == "" {
		dialog.ShowInformation("Validation Error", "Please enter password", g.window)
		return
	}
	if g.saltEntry.Text == "" {
		dialog.ShowInformation("Validation Error", "Please enter salt", g.window)
		return
	}
	if g.compressionLevelSelect.Selected == "" {
		dialog.ShowInformation("Validation Error", "Please select compression level", g.window)
		return
	}
	if len(g.selectedFiles) == 0 {
		dialog.ShowInformation("Validation Error", "Please select files", g.window)
		return
	}
	if g.outputEntry.Text == "" {
		dialog.ShowInformation("Validation Error", "Please enter output filename", g.window)
		return
	}

	g.showProgress("Creating archive...")
	g.clearResults()

	go func() {
		defer g.hideProgress()

		compressLevel := g.getSelectedCompressionLevel()

		if err := g.createOutputDir(); err != nil {
			g.showError(fmt.Sprintf("Error creating output directory: %v", err))
			return
		}
		fullOutputPath := filepath.Join("output", g.outputEntry.Text)

		files, err := archiver.CollectFiles(g.selectedFiles)
		if err != nil {
			g.showError(fmt.Sprintf("Error collecting files: %v", err))
			return
		}

		stats, err := g.calculateStatistics(files, g.passwordEntry.Text, g.saltEntry.Text, compressLevel)
		if err != nil {
			g.showError(fmt.Sprintf("Error calculating statistics: %v", err))
			return
		}

		err = archiver.CreateArchive(g.passwordEntry.Text, g.saltEntry.Text, fullOutputPath, files, compressLevel)
		if err != nil {
			g.showError(fmt.Sprintf("Error creating archive: %v", err))
			return
		}

		g.showResults(stats, fullOutputPath)
		g.showSuccess("Archive created successfully!")
	}()
}

func (g *GUI) extractArchive() {
	if g.extractPasswordEntry.Text == "" {
		dialog.ShowInformation("Validation Error", "Please enter password", g.window)
		return
	}
	if g.extractSaltEntry.Text == "" {
		dialog.ShowInformation("Validation Error", "Please enter salt", g.window)
		return
	}
	if g.selectedArchive == "" {
		dialog.ShowInformation("Validation Error", "Please select archive file", g.window)
		return
	}

	g.showProgress("Extracting archive...")
	g.clearResults()

	go func() {
		defer g.hideProgress()

		err := archiver.ExtractArchive(g.extractPasswordEntry.Text,
			g.extractSaltEntry.Text, g.selectedArchive)
		if err != nil {
			g.showError(fmt.Sprintf("Error extracting archive: %v", err))
			return
		}

		g.showSuccess("Archive extracted successfully!")
	}()
}

func (g *GUI) getSelectedCompressionLevel() int {
	selected := g.compressionLevelSelect.Selected
	if selected == "" {
		return 6
	}

	level := int(selected[0] - '0')
	if level < 0 || level > 9 {
		return 6
	}
	return level
}

func (g *GUI) calculateStatistics(files []archiver.FileInfo, password, saltHex string, compressLevel int) (*Statistics, error) {
	stats := &Statistics{
		FileStats: make([]FileStat, 0, len(files)),
	}

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return nil, err
	}

	key, err := archiver.GenerateKey(password, salt)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		data, err := os.ReadFile(file.Path)
		if err != nil {
			return nil, err
		}

		compressedData, err := archiver.Compress(data, compressLevel)
		if err != nil {
			return nil, err
		}

		encryptedData, err := archiver.Encrypt(compressedData, key)
		if err != nil {
			return nil, err
		}

		entropy := calculateEntropy(data)
		compressionRatio := float64(len(compressedData)) / float64(len(data)) * 100

		fileStat := FileStat{
			Filename:         filepath.Base(file.Path),
			OriginalSize:     int64(len(data)),
			CompressedSize:   int64(len(compressedData)),
			EncryptedSize:    int64(len(encryptedData)),
			Entropy:          entropy,
			CompressionRatio: compressionRatio,
		}

		stats.FileStats = append(stats.FileStats, fileStat)
		stats.OriginalSize += int64(len(data))
		stats.CompressedSize += int64(len(compressedData))
		stats.EncryptedSize += int64(len(encryptedData))
	}

	if stats.OriginalSize > 0 {
		stats.CompressionRatio = float64(stats.CompressedSize) / float64(stats.OriginalSize) * 100
	}

	return stats, nil
}

func (g *GUI) showResults(stats *Statistics, outputPath string) {
	var result strings.Builder
	result.WriteString("=== COMPRESSION STATISTICS ===\n\n")

	compressionLevel := g.getSelectedCompressionLevel()
	result.WriteString(fmt.Sprintf("Compression Level: %d\n\n", compressionLevel))

	for _, fileStat := range stats.FileStats {
		result.WriteString(fmt.Sprintf("📁 %s\n", fileStat.Filename))
		result.WriteString(fmt.Sprintf("   Original: %s\n", formatFileSize(fileStat.OriginalSize)))
		result.WriteString(fmt.Sprintf("   Compressed: %s (%.2f%%)\n", formatFileSize(fileStat.CompressedSize), fileStat.CompressionRatio))
		result.WriteString(fmt.Sprintf("   Encrypted: %s\n", formatFileSize(fileStat.EncryptedSize)))
		result.WriteString(fmt.Sprintf("   Entropy: %.4f\n\n", fileStat.Entropy))
	}

	result.WriteString("=== FINAL RESULTS ===\n")
	result.WriteString(fmt.Sprintf("Original total: %s (%.2f MB)\n",
		formatFileSize(stats.OriginalSize), float64(stats.OriginalSize)/(1024*1024)))
	result.WriteString(fmt.Sprintf("Compressed total: %s (%.2f MB)\n",
		formatFileSize(stats.CompressedSize), float64(stats.CompressedSize)/(1024*1024)))
	result.WriteString(fmt.Sprintf("Encrypted total: %s (%.2f MB)\n",
		formatFileSize(stats.EncryptedSize), float64(stats.EncryptedSize)/(1024*1024)))
	result.WriteString(fmt.Sprintf("Final compression: %.2f%%\n", stats.CompressionRatio))
	result.WriteString(fmt.Sprintf("Archive overhead: %.2f%%\n",
		float64(stats.EncryptedSize-stats.OriginalSize)/float64(stats.OriginalSize)*100))
	result.WriteString(fmt.Sprintf("Output file: %s\n", outputPath))

	fyne.Do(func() {
		g.resultsText.SetText(result.String())
		g.resultsContainer.Show()
		g.closeResultsBtn.Show()
	})
}

func (g *GUI) clearResults() {
	fyne.Do(func() {
		g.resultsText.SetText("")
		g.resultsContainer.Hide()
		g.closeResultsBtn.Hide()
	})
}

func (g *GUI) createOutputDir() error {
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		err := os.Mkdir("output", 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking output directory: %v", err)
	}
	return nil
}

func (g *GUI) showProgress(message string) {
	fyne.Do(func() {
		g.progressBar.Show()
		g.progressLabel.Show()
		g.progressLabel.SetText(message)
		g.progressBar.SetValue(0)

		g.createArchiveBtn.Disable()
		g.extractBtn.Disable()
	})
}

func (g *GUI) hideProgress() {
	fyne.Do(func() {
		g.progressBar.Hide()
		g.progressLabel.Hide()

		g.createArchiveBtn.Enable()
		g.extractBtn.Enable()
	})
}

func (g *GUI) showError(message string) {
	fyne.Do(func() {
		dialog.ShowError(fmt.Errorf(message), g.window)
	})
}

func (g *GUI) showSuccess(message string) {
	fyne.Do(func() {
		dialog.ShowInformation("Success", message, g.window)
	})
}

func (g *GUI) ShowAndRun() {
	g.window.ShowAndRun()
}

func generateRandomSalt(length int) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

func calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}

	var freq [256]int
	for _, b := range data {
		freq[b]++
	}

	var entropy float64
	for _, count := range freq {
		if count > 0 {
			p := float64(count) / float64(len(data))
			entropy -= p * math.Log2(p)
		}
	}

	return entropy / 8.0
}

func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
