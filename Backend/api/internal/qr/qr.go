package qr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/skip2/go-qrcode"
)

var ErrQRNotFound = errors.New("QR kod görseli bulunamadı")

// Store QR kod PNG dosyalarının disk üzerinde yönetimini sağlar.
type Store struct {
	dir string
}

// NewStore çalışma ortamına göre QR depolama dizinini belirler ve oluşturur.
func NewStore(dir string) (*Store, error) {
	if strings.TrimSpace(dir) == "" {
		for _, cand := range []string{"/storage/qrcodes", "../storage/qrcodes", "./Backend/storage/qrcodes"} {
			if info, err := os.Stat(cand); err == nil && info.IsDir() {
				dir = cand
				break
			}
		}
		if dir == "" {
			dir = "../storage/qrcodes"
		}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("qr depolama dizini oluşturulamadı (%s): %w", dir, err)
	}

	return &Store{dir: dir}, nil
}

// Generate verilen publicID ve kısa link URL'i için 512x512 PNG QR kodu üretir ve dosya adını döner.
func (s *Store) Generate(publicID, shortURL string) (string, error) {
	cleanID := filepath.Base(strings.TrimSpace(publicID))
	if cleanID == "" || cleanID == "." || cleanID == ".." {
		return "", errors.New("geçersiz public_id")
	}

	filename := cleanID + ".png"
	fullPath := filepath.Join(s.dir, filename)

	if err := qrcode.WriteFile(shortURL, qrcode.Medium, 512, fullPath); err != nil {
		return "", fmt.Errorf("qr kodu üretilemedi: %w", err)
	}

	return filename, nil
}

// ReadPath verilen dosya adının güvenli tam yolunu doğrular ve döner.
func (s *Store) ReadPath(filename string) (string, error) {
	clean := filepath.Base(strings.TrimSpace(filename))
	if clean == "" || clean == "." || clean == ".." || !strings.HasSuffix(clean, ".png") {
		return "", ErrQRNotFound
	}

	fullPath := filepath.Join(s.dir, clean)
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		return "", ErrQRNotFound
	}

	return fullPath, nil
}

// Delete disk üzerindeki QR kod PNG dosyasını güvenli şekilde siler.
func (s *Store) Delete(filename string) {
	clean := filepath.Base(strings.TrimSpace(filename))
	if clean == "" || clean == "." || clean == ".." || !strings.HasSuffix(clean, ".png") {
		return
	}
	_ = os.Remove(filepath.Join(s.dir, clean))
}
