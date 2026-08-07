package main

import (
	"archive/zip"
	"io"
	"os"
)

// CompressFile takes a single input file and zips it into an output archive
func CompressFile(inputFile, outputFile string) error {
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	zipWriter := zip.NewWriter(out)
	defer zipWriter.Close()

	fileToZip, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Create an entry in the zip archive matching the file's name
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate // Standard zip compression

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Copy the file bytes into the zip entry
	_, err = io.Copy(writer, fileToZip)
	return err
}
