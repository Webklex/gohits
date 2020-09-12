package filesystem

import (
	"os"
	"path/filepath"
)

func MakeDir(filename string) (dbdir string, err error) {
	dbdir = filepath.Dir(filename)
	_, err = os.Stat(dbdir)
	if err != nil {
		err = os.MkdirAll(dbdir, 0755)
		if err != nil {
			return "", err
		}
	}
	return dbdir, nil
}

func RenameFile(fromName string, toName string) error {
	if err := os.Rename(toName, toName+".bak"); err != nil {
	}
	if _, err := MakeDir(toName); err != nil {
		return err
	}
	return os.Rename(fromName, toName)
}

func CreateDirectory(dirName string) bool {
	src, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirName, 0755)
		if errDir != nil {
			panic(err)
		}
		return true
	}

	if src.Mode().IsRegular() {
		return false
	}

	return false
}
