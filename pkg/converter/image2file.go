package converter

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

func Image2File(img image.Image, path string) (*os.File, error) {
	p, _ := filepath.Abs(path)

	f, err := os.Create(p)
	if err != nil {
		return nil, fmt.Errorf("Error creating wordcloud file: %w", err)
	}

	if err := png.Encode(f, img); err != nil {
		return nil, fmt.Errorf("Error encoding wordcloud: %w", err)
	}

	_, _ = f.Seek(0, io.SeekStart)

	return f, nil
}
