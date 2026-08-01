package certificate

import (
	"os"
	"path/filepath"
)

func inputPDF(path string) {
	os.ReadFile(path)
	cerName := filepath.Base(path)

	
}
