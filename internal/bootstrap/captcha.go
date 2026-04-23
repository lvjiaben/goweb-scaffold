package bootstrap

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"strings"
	"sync"
	"time"
)

var (
	ErrCaptchaInvalid  = errors.New("验证码错误")
	ErrCaptchaExpired  = errors.New("验证码已过期")
	ErrCaptchaNotFound = errors.New("验证码不存在")
)

type CaptchaService struct {
	ttl   time.Duration
	mu    sync.Mutex
	items map[string]captchaRecord
}

type captchaRecord struct {
	Code      string
	ExpiresAt time.Time
}

func NewCaptchaService(ttl time.Duration) *CaptchaService {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &CaptchaService{
		ttl:   ttl,
		items: make(map[string]captchaRecord),
	}
}

func (s *CaptchaService) Generate() (string, string, error) {
	code, err := randomDigits(4)
	if err != nil {
		return "", "", fmt.Errorf("captcha randomDigits: %w", err)
	}
	id, err := randomHex(16)
	if err != nil {
		return "", "", fmt.Errorf("captcha randomHex: %w", err)
	}

	now := time.Now()
	s.mu.Lock()
	s.cleanupExpiredLocked(now)
	s.items[id] = captchaRecord{
		Code:      code,
		ExpiresAt: now.Add(s.ttl),
	}
	s.mu.Unlock()

	rawPNG, err := renderCaptchaPNG(code)
	if err != nil {
		return "", "", fmt.Errorf("captcha renderCaptchaPNG: %w", err)
	}
	return id, "data:image/png;base64," + base64.StdEncoding.EncodeToString(rawPNG), nil
}

func (s *CaptchaService) Verify(id string, code string) error {
	id = strings.TrimSpace(id)
	code = normalizeCaptchaCode(code)
	if id == "" {
		return ErrCaptchaNotFound
	}
	if code == "" {
		return ErrCaptchaInvalid
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.items[id]
	if !ok {
		return ErrCaptchaNotFound
	}
	if now.After(record.ExpiresAt) {
		delete(s.items, id)
		return ErrCaptchaExpired
	}
	if normalizeCaptchaCode(record.Code) != code {
		return ErrCaptchaInvalid
	}

	delete(s.items, id)
	return nil
}

func (s *CaptchaService) cleanupExpiredLocked(now time.Time) {
	for id, record := range s.items {
		if now.After(record.ExpiresAt) {
			delete(s.items, id)
		}
	}
}

func normalizeCaptchaCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func randomDigits(length int) (string, error) {
	if length <= 0 {
		length = 4
	}
	builder := strings.Builder{}
	builder.Grow(length)
	for range length {
		number, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('0' + number.Int64()))
	}
	return builder.String(), nil
}

func randomHex(byteLength int) (string, error) {
	if byteLength <= 0 {
		byteLength = 16
	}
	buffer := make([]byte, byteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func renderCaptchaPNG(code string) ([]byte, error) {
	const (
		width      = 160
		height     = 56
		top        = 8
		left       = 10
		cellWidth  = 34
		digitWidth = 22
		digitH     = 34
		thickness  = 5
	)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{245, 248, 252, 255}}, image.Point{}, draw.Src)

	for i := 0; i < 8; i++ {
		x1, y1, err := randomPoint(width, height)
		if err != nil {
			return nil, fmt.Errorf("noise line start point: %w", err)
		}
		x2, y2, err := randomPoint(width, height)
		if err != nil {
			return nil, fmt.Errorf("noise line end point: %w", err)
		}
		drawLine(img, x1, y1, x2, y2, randomPastel())
	}

	for i := 0; i < 48; i++ {
		x, y, err := randomPoint(width, height)
		if err != nil {
			return nil, fmt.Errorf("noise pixel point: %w", err)
		}
		img.Set(x, y, randomDark())
	}

	for index, digit := range code {
		offsetX := left + index*cellWidth
		jitterY, err := randomOffset(-2, 2)
		if err != nil {
			return nil, fmt.Errorf("digit y jitter: %w", err)
		}
		jitterX, err := randomOffset(-1, 2)
		if err != nil {
			return nil, fmt.Errorf("digit x jitter: %w", err)
		}
		drawDigit(img, offsetX+jitterX, top+jitterY, digitWidth, digitH, thickness, digit, randomDark())
	}

	buffer := bytes.NewBuffer(nil)
	if err := png.Encode(buffer, img); err != nil {
		return nil, fmt.Errorf("png encode: %w", err)
	}
	return buffer.Bytes(), nil
}

func randomPoint(width int, height int) (int, int, error) {
	x, err := randomRange(0, width-1)
	if err != nil {
		return 0, 0, err
	}
	y, err := randomRange(0, height-1)
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

func randomOffset(min int, max int) (int, error) {
	return randomRange(min, max)
}

func randomRange(min int, max int) (int, error) {
	if max <= min {
		return min, nil
	}
	size := int64(max - min + 1)
	number, err := rand.Int(rand.Reader, big.NewInt(size))
	if err != nil {
		return 0, err
	}
	return min + int(number.Int64()), nil
}

func randomPastel() color.RGBA {
	return color.RGBA{R: 180, G: 205, B: 230, A: 255}
}

func randomDark() color.RGBA {
	return color.RGBA{R: 35, G: 52, B: 74, A: 255}
}

func drawDigit(img *image.RGBA, x int, y int, width int, height int, thickness int, digit rune, stroke color.RGBA) {
	segments, ok := sevenSegmentMap[digit]
	if !ok {
		return
	}

	horizontalWidth := width
	verticalHeight := (height - thickness) / 2

	if segments[0] {
		fillRect(img, x, y, horizontalWidth, thickness, stroke)
	}
	if segments[1] {
		fillRect(img, x+width-thickness, y, thickness, verticalHeight, stroke)
	}
	if segments[2] {
		fillRect(img, x+width-thickness, y+verticalHeight, thickness, verticalHeight, stroke)
	}
	if segments[3] {
		fillRect(img, x, y+height-thickness, horizontalWidth, thickness, stroke)
	}
	if segments[4] {
		fillRect(img, x, y+verticalHeight, thickness, verticalHeight, stroke)
	}
	if segments[5] {
		fillRect(img, x, y, thickness, verticalHeight, stroke)
	}
	if segments[6] {
		fillRect(img, x, y+verticalHeight-(thickness/2), horizontalWidth, thickness, stroke)
	}
}

func drawLine(img *image.RGBA, x0 int, y0 int, x1 int, y1 int, stroke color.RGBA) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	errValue := dx + dy

	for {
		if image.Pt(x0, y0).In(img.Bounds()) {
			img.Set(x0, y0, stroke)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		doubleError := 2 * errValue
		if doubleError >= dy {
			errValue += dy
			x0 += sx
		}
		if doubleError <= dx {
			errValue += dx
			y0 += sy
		}
	}
}

func fillRect(img *image.RGBA, x int, y int, width int, height int, fill color.RGBA) {
	rect := image.Rect(x, y, x+width, y+height).Intersect(img.Bounds())
	if rect.Empty() {
		return
	}
	draw.Draw(img, rect, &image.Uniform{C: fill}, image.Point{}, draw.Src)
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

var sevenSegmentMap = map[rune][7]bool{
	'0': {true, true, true, true, true, true, false},
	'1': {false, true, true, false, false, false, false},
	'2': {true, true, false, true, true, false, true},
	'3': {true, true, true, true, false, false, true},
	'4': {false, true, true, false, false, true, true},
	'5': {true, false, true, true, false, true, true},
	'6': {true, false, true, true, true, true, true},
	'7': {true, true, true, false, false, false, false},
	'8': {true, true, true, true, true, true, true},
	'9': {true, true, true, true, false, true, true},
}
