package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricardofabila/fox/src/constants"
)

func FileHasZIPExtension(filename string) bool {
	for _, extension := range constants.ZIPExtensions {
		if strings.HasSuffix(filename, extension) {
			return true
		}
	}

	return false
}

func ZIPWithoutExtension(fileName string) string {
	extension := ""
	for _, e := range constants.ZIPExtensions {
		if strings.HasSuffix(fileName, e) {
			extension = e
			break
		}
	}

	return strings.TrimSuffix(fileName, extension)
}

// ExtractZIP takes a destination path and a reader; a tar reader loops over the zip file
// creating the file structure at 'dst' along the way, and writing any files
func ExtractZIP(src, dst string) error {
	archive, err := zip.OpenReader(src)
	if err != nil {
		return err
	}

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path")
		}
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, errD := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if errD != nil {
			return errD
		}

		fileInArchive, errF := f.Open()
		if errF != nil {
			return errF
		}

		if _, err = io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		err = dstFile.Close()
		if err != nil {
			return err
		}

		err = fileInArchive.Close()
		if err != nil {
			return err
		}
	}

	err = archive.Close()
	if err != nil {
		return err
	}

	return nil
}
