package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ganeshrvel/go-mtpfs/mtp"
	"github.com/ganeshrvel/go-mtpx"
)

const SEPARATOR string = "----------------------------------------"

func printFiles(dev *mtp.Device, storageId uint32, targetPath string) {
	fmt.Printf("\nFiles in %s (storage ID: %d):\n", targetPath, storageId)
	fmt.Println(SEPARATOR)

	_, totalFiles, totalDirs, err := mtpx.Walk(
		dev,
		storageId,
		targetPath,
		false, // non-recursive (only root level)
		true,  // skip disallowed files
		false, // don't skip hidden files
		func(objectId uint32, fi *mtpx.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Display directory indicator
			typeIndicator := "📄"
			if fi.IsDir {
				typeIndicator = "📁"
			}

			// Format size
			sizeStr := fmt.Sprintf("%d bytes", fi.Size)
			if fi.Size > 1024*1024 {
				sizeStr = fmt.Sprintf("%.2f MB", float64(fi.Size)/(1024*1024))
			} else if fi.Size > 1024 {
				sizeStr = fmt.Sprintf("%.2f KB", float64(fi.Size)/1024)
			}

			fmt.Printf("%s %s", typeIndicator, fi.Name)
			if !fi.IsDir {
				fmt.Printf(" (%s)", sizeStr)
			}
			fmt.Println()

			return nil
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list directory: %v", err)
	}

	fmt.Println(SEPARATOR)
	fmt.Printf("Total: %d files, %d directories\n", totalFiles, totalDirs)
}

func uploadFile(dev *mtp.Device, sourcePath string, storageId uint32, destPath string) {
	// Verify the source file exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Source file not found: %v\n", err)
	}

	fmt.Printf("Uploading: %s (%d bytes)\n", filepath.Base(sourcePath), sourceInfo.Size())
	fmt.Printf("Destination: %s\n", destPath)
	fmt.Println(SEPARATOR)

	sources := []string{sourcePath}

	// Upload the file with progress tracking
	_, totalFiles, totalSize, err := mtpx.UploadFiles(
		dev,
		storageId,
		sources,
		destPath,
		true, // preprocessing enabled to get total size for progress calculation
		// Preprocessing callback - called for each file before upload starts
		func(fi *os.FileInfo, fullPath string, err error) error {
			if err != nil {
				return err
			}
			return nil
		},
		// Progress callback
		func(pi *mtpx.ProgressInfo, err error) error {
			if err != nil {
				return err
			}

			// Display progress
			switch pi.Status {
			case mtpx.InProgress:
				progressPct := pi.ActiveFileSize.Progress

				fmt.Printf("\rProgress: %.1f%%", progressPct)
			case mtpx.Completed:
				fmt.Printf("\rProgress: 100.0%% | Completed!\n")
			}

			return nil
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Upload failed: %v", err)
	}

	fmt.Println(SEPARATOR)
	fmt.Printf("Upload complete! Files: %d, Total size: %d bytes\n", totalFiles, totalSize)
}

func printStorages(storages []mtpx.StorageData) {
	fmt.Printf("\nFound %d storage(s):\n", len(storages))
	for i, storage := range storages {
		fmt.Printf("  [%d] %s (ID: %d)\n", i, storage.Info.StorageDescription, storage.Sid)
		fmt.Printf("      Free: %d MB / Total: %d MB\n",
			storage.Info.FreeSpaceInBytes/(1024*1024),
			storage.Info.MaxCapability/(1024*1024))
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Too few arguments provided\n")
		os.Exit(1)
	}

	fmt.Println("Connecting to MTP device...")

	// Initialize
	dev, err := mtpx.Initialize(mtpx.Init{DebugMode: false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize device: %v", err)
	}
	defer mtpx.Dispose(dev)

	// Fetch device info
	info, err := mtpx.FetchDeviceInfo(dev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get device info: %v", err)
	}
	fmt.Printf("Model: %s by %s\n", info.Model, info.Manufacturer)

	// Fetch available storages
	storages, err := mtpx.FetchStorages(dev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch storages: %v", err)
	}

	printStorages(storages)

	// Use the first storage always
	if len(storages) == 0 {
		fmt.Fprintf(os.Stderr, "No storage found on device\n")
		os.Exit(1)
	}
	storageId := storages[0].Sid

	opMode := os.Args[1]

	switch opMode {
	case "-l":
		destListPath := "/Download"
		if len(os.Args) >= 3 {
			destListPath = os.Args[2]
		}
		printFiles(dev, storageId, destListPath)
	case "-u":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: -u <source_file> [destination_path]")
		}
		filePath := os.Args[2]
		// Default destination is /Download, but can be overridden
		destPath := "/Download"
		if len(os.Args) >= 4 {
			destPath = os.Args[3]
		}
		uploadFile(dev, filePath, storageId, destPath)
	default:
		fmt.Fprintf(os.Stderr, "Wrong mode!")
	}
}
