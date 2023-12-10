package files

import (
	"fmt"
	"os"
	"time"
)

// FileSizeInMB returns the size of the file in MB
func FileSizeInMB(name string) (float64, error) {
	fileInfo, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("file does not exist: %v", err)
		}
		return 0, err
	}

	size := fileInfo.Size() / (1024 * 1024)
	return float64(size), nil
}

// LastModifiedTime returns the last modification time of file
func LastModifiedTime(name string) (time.Time, error) {
	fileInfo, err := os.Stat(name)
	if err != nil {
		return time.Time{}, err
	}
	return fileInfo.ModTime(), nil
}

// RenameFile renames the file
func RenameFile(oldName, newName string) error {
	return os.Rename(oldName, newName)
}

// PrepareName prepares new name for the file
func PrepareName(name string, t time.Time) string {
	return fmt.Sprintf("%s%s.txt", name, t.Format("20060102"))
}

// SplitFile checks if file size is greater than 1MB, and if so, renames it
func SplitFile(name string) error {
	size, err := FileSizeInMB(name + ".txt")
	if err != nil {
		return err
	}
	if size > 1 {
		modTime, err := LastModifiedTime(name + ".txt")
		if err != nil {
			return err
		}

		newName := PrepareName(name, modTime)
		if err = RenameFile(name+".txt", newName); err != nil {
			return err
		}
	}

	return nil
}
