package internal

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
)

// CopyDir recursively copies a directory from src to dst
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}

		defer srcFile.Close()
		dstFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}

		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	
	return out.Sync()
}

// CreateDirIfNotExist creates a directory if it doesn't exist
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// CopyCertificates copies certificate files to the destination directory
func CopyCertificates(crtPath, keyPath, dstDir string) error {
	if err := CreateDirIfNotExist(dstDir); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}
	
	// Copy certificate
	if err := CopyFile(crtPath, filepath.Join(dstDir, "cert.crt")); err != nil {
		return fmt.Errorf("failed to copy certificate: %w", err)
	}
	
	// Copy key
	if err := CopyFile(keyPath, filepath.Join(dstDir, "cert.key")); err != nil {
		return fmt.Errorf("failed to copy certificate key: %w", err)
	}
	
	return nil
}

// CopyTemplatedDirectories copies directories from template to project directory
func CopyTemplatedDirectories(templateDir, projectDir string, folders []string) error {
	for _, dir := range folders {
		src := filepath.Join(templateDir, dir)
		dst := filepath.Join(projectDir, dir)
		
		// Check if source directory exists
		if _, err := os.Stat(src); err == nil {
			if err := CopyDir(src, dst); err != nil {
				return fmt.Errorf("failed to copy %s: %w", dir, err)
			}
		}
	}
	return nil
}
