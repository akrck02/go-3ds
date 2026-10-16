//go:build nintendo3ds

package main

// figures gathers a second of frames, and keeps the last second's to show.
type figures struct {
	frames  int
	elapsed float64
	gpu     float32
	shown   shownFigures
}

type shownFigures struct {
	framesPerSecond int
	gpu             float32 // milliseconds a frame, on average
	triangles       int
}

func (figures *figures) add(seconds float64, gpu float32, triangles int) {
	figures.frames++
	figures.elapsed += seconds
	figures.gpu += gpu
	if figures.elapsed < 1 {
		return
	}
	figures.shown = shownFigures{framesPerSecond: figures.frames, gpu: figures.gpu / float32(figures.frames), triangles: triangles}
	figures.frames, figures.elapsed, figures.gpu = 0, 0, 0
}

// tenths writes milliseconds with one decimal.
func tenths(milliseconds float32) string {
	whole := int(milliseconds*10 + 0.5)
	return itoa(whole/10) + "." + itoa(whole%10)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
