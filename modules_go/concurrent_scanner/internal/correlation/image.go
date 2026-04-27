package correlation

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"sync"
	"time"

	"github.com/nfnt/resize" // Note: While I try to avoid dependencies, resize is standard for this task. 
	// If I can't use it, I'll implement a very basic downsampler.
	// Actually, I'll implement a basic bilinear scaler to stay dependency-free as requested.
)

var (
	hashCache = make(map[string]string)
	cacheMu   sync.RWMutex
	client    = &http.Client{
		Timeout: 10 * time.Second,
	}
)

// GetImageHash downloads and generates an Average Hash (aHash) for an image
func GetImageHash(url string) (string, error) {
	if url == "" {
		return "", nil
	}

	// 1. Check Cache
	cacheMu.RLock()
	if hash, exists := hashCache[url]; exists {
		cacheMu.RUnlock()
		return hash, nil
	}
	cacheMu.RUnlock()

	// 2. Download Image
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", err
	}

	// 3. Generate aHash
	hash := calculateAHash(img)

	// 4. Update Cache
	cacheMu.Lock()
	hashCache[url] = hash
	cacheMu.Unlock()

	return hash, nil
}

// calculateAHash implements a simple Average Hash algorithm
func calculateAHash(img image.Image) string {
	// Resize to 8x8
	resized := resize8x8(img)
	
	// Convert to Grayscale and find average
	pixels := make([]uint32, 64)
	var sum uint32
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			// Standard grayscale conversion
			gray := (r*299 + g*587 + b*114) / 1000
			pixels[y*8+x] = gray
			sum += gray
		}
	}
	avg := sum / 64

	// Generate hash bits
	var hash uint64
	for i, p := range pixels {
		if p >= avg {
			hash |= (1 << uint(i))
		}
	}

	return fmt.Sprintf("%016x", hash)
}

// resize8x8 is a very basic nearest-neighbor downsampler to 8x8
func resize8x8(img image.Image) image.Image {
	res := image.NewRGBA(image.Rect(0, 0, 8, 8))
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			srcX := x * w / 8
			srcY := y * h / 8
			res.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}
	return res
}

// HammingDistance calculates the number of bits that differ between two hex hashes
func HammingDistance(h1, h2 string) int {
	if len(h1) != len(h2) || len(h1) == 0 {
		return 64 // Max difference for 16-char hex (64 bits)
	}

	var dist int
	for i := 0; i < len(h1); i++ {
		v1 := hexDigitToInt(h1[i])
		v2 := hexDigitToInt(h2[i])
		diff := v1 ^ v2
		// Count set bits
		for diff > 0 {
			dist += diff & 1
			diff >>= 1
		}
	}
	return dist
}

func hexDigitToInt(b byte) int {
	if b >= '0' && b <= '9' {
		return int(b - '0')
	}
	if b >= 'a' && b <= 'f' {
		return int(b - 'a' + 10)
	}
	if b >= 'A' && b <= 'F' {
		return int(b - 'A' + 10)
	}
	return 0
}
