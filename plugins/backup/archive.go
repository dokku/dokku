package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TarGzCreate writes a gzip-compressed tarball of srcDir's contents to
// destPath. Entries are stored with paths relative to srcDir (no leading
// directory), so extraction yields the backup tree directly.
func TarGzCreate(srcDir string, destPath string) error {
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("unable to create tarball: %w", err)
	}
	defer out.Close()

	gzw := gzip.NewWriter(out)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	walkErr := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == srcDir {
			return nil
		}

		// Symlinks are not archived: backups contain only regular files and
		// directories, and restoring a symlink from archive data is a path-
		// traversal risk, so they are skipped entirely.
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})
	if walkErr != nil {
		return fmt.Errorf("unable to write tarball: %w", walkErr)
	}

	if err := tw.Close(); err != nil {
		return err
	}
	if err := gzw.Close(); err != nil {
		return err
	}
	return out.Close()
}

// TarGzExtract extracts a gzip-compressed tarball at srcPath into destDir. It
// rejects entries that would escape destDir (path traversal / absolute paths)
// so a malicious archive cannot write outside the extraction directory.
func TarGzExtract(srcPath string, destDir string) error {
	in, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("unable to open backup file: %w", err)
	}
	defer in.Close()

	gzr, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("unable to read backup file: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	cleanDest := filepath.Clean(destDir)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("unable to read backup entry: %w", err)
		}

		target := filepath.Join(cleanDest, header.Name)
		if !isWithin(cleanDest, target) {
			return fmt.Errorf("refusing to extract entry outside backup directory: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)&os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := writeFileFromTar(tr, target, os.FileMode(header.Mode)&os.ModePerm); err != nil {
				return err
			}
		}
		// Other entry types (symlinks, hardlinks, devices) are intentionally
		// skipped: a backup contains only regular files and directories, and
		// recreating links from archive data would be a path-traversal risk.
	}

	return nil
}

// writeFileFromTar copies the current tar entry into target with the given mode.
func writeFileFromTar(tr *tar.Reader, target string, mode os.FileMode) error {
	file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, tr); err != nil {
		return err
	}
	return file.Close()
}

// isWithin reports whether target is the base directory or a path inside it.
func isWithin(base string, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
