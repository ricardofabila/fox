package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricardofabila/fox/src/constants"
)

func FileHasTarExtension(filename string) bool {
	for _, extension := range constants.TarExtensions {
		if strings.HasSuffix(filename, extension) {
			return true
		}
	}

	return false
}

func TarWithoutExtension(fileName string) string {
	extension := ""
	for _, e := range constants.TarExtensions {
		if strings.HasSuffix(fileName, e) {
			extension = e
			break
		}
	}

	return strings.TrimSuffix(fileName, extension)
}

// ExtractTAR takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func ExtractTAR(src, dst string) error {
	r, e := os.Open(src)
	if e != nil {
		err := r.Close()
		if err != nil {
			return err
		}

		return e
	}

	gzr, gErr := gzip.NewReader(r)
	if gErr != nil {
		return gErr
	}

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			err = gzr.Close()
			if err != nil {
				return err
			}
			err = r.Close()
			if err != nil {
				return err
			}

			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the current location where the dir/file should be created
		current := filepath.Join(dst, header.Name)
		dir, _ := filepath.Split(current)
		CreateDirectoryIfNotExists(dir)

		// the following switch could also be done using fi.Mode(), not sure if is there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {
		// if it's a dir, and it doesn't, exist create it
		case tar.TypeDir:
			if _, err = os.Stat(current); err != nil {
				if err = os.MkdirAll(current, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, fErr := os.OpenFile(current, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if fErr != nil {
				// for some reason directories are sometimes mistaken as regular files
				if strings.Contains(fErr.Error(), "is a directory") {
					if _, err = os.Stat(current); err != nil {
						if err = os.MkdirAll(current, 0755); err != nil {
							return err
						}
					}

					continue
				}

				return fErr
			}

			// copy over contents
			if _, err = io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			err = f.Close()
			if err != nil {
				return err
			}
		}
	}
}
