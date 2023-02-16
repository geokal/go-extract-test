package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var appPackagesBasePath = "/home/ubuntu/go-extract-test/packages"

func main() {
	packageId := "f59620d2-5f89-4776-97aa-01f21a191a49"
	hostIp := "192.168.2.1"
	
	packagePath := appPackagesBasePath + "/" + packageId + hostIp
	

	packageFilePath := packagePath + "/" + packageId + ".csar"
	
	packagePath, err := extractCsarPackage(packageFilePath)
	if err != nil {
		log.Error(err)
		return
	}

	fmt.Println(packagePath)
}

// Create directory
func CreateDir(path string) bool {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 488)
		if errDir != nil {
			return false
		}
	}
	return true
}

// extract CSAR package
func extractCsarPackage(packagePath string) (string, error) {
	zipReader, _ := zip.OpenReader(packagePath)
	if len(zipReader.File) > 1024 {
		return "", errors.New("Too many files contains in zip file")
	}
	defer zipReader.Close()
	var totalWrote int64
	packageDir := path.Dir(packagePath)
	err := os.MkdirAll(packageDir, 0750)
	if err != nil {
		log.Error("Failed to make directory")
		return "" ,errors.New("Failed to make directory")
	}
	for _, file := range zipReader.Reader.File {

		zippedFile, err := file.Open()
		if err != nil || zippedFile == nil {
			log.Error("Failed to open zip file")
			continue
		}
		if file.UncompressedSize64 > 104857600 || totalWrote > 104857600 {
			log.Error("File size limit is exceeded")
		}

		defer zippedFile.Close()

		isContinue, wrote := extractFiles(file, zippedFile, totalWrote, packageDir)
		if isContinue {
			continue
		}
		totalWrote = wrote
	}
	return packageDir, nil
}

// Extract files
func extractFiles(file *zip.File, zippedFile io.ReadCloser, totalWrote int64, dirName string) (bool, int64) {
	targetDir := dirName
	extractedFilePath := filepath.Join(
		targetDir,
		file.Name,
	)

	if file.FileInfo().IsDir() {
		err := os.MkdirAll(extractedFilePath, 0750)
		if err != nil {
			log.Error("Failed to create directory")
		}
	} else {
		parent := filepath.Dir(extractedFilePath)
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			_ = os.MkdirAll(parent, 0750)
		}
		outputFile, err := os.OpenFile(
			extractedFilePath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0750,
		)
		if err != nil || outputFile == nil {
			log.Error("The output file is nil")
			return true, totalWrote
		}

		defer outputFile.Close()

		wt, err := io.Copy(outputFile, zippedFile)
		if err != nil {
			log.Error("Failed to copy zipped file")
		}
		totalWrote += wt
	}
	return false, totalWrote
}

