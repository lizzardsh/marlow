package terminal

import (
	"fmt"
	"strconv"
	"strings"
)

// Renkler
const (
	BannerGradientStart = "#FF9F06"
	BannerGradientEnd   = "#FFE45B"
	Success             = "#01c761"
	Info                = "b8b8b8"
	Warning             = "#ffbd2d"
	Error               = "#ff4c40"
	reset               = "\033[0m"
)

func Paint(hex, text string) string {
	color := ansiColor(hex)
	if color == "" {
		return text
	}
	return color + text + reset
}

func ansiColor(value string) string {
	red, green, blue, valid := parseHex(value)
	if !valid {
		return ""
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", red, green, blue)
}

func Gradient(startHex, endHex, text string) string {
	startRed, startGreen, startBlue, startOK := parseHex(startHex)
	endRed, endGreen, endBlue, endOK := parseHex(endHex)
	if !startOK || !endOK {
		return text
	}

	count := 0
	for _, char := range text {
		if char != '\n' && char != '\r' {
			count++
		}
	}
	if count == 0 {
		return text
	}

	var output strings.Builder
	output.Grow(len(text) + count*20)
	index := 0
	for _, char := range text {
		if char == '\n' || char == '\r' {
			output.WriteRune(char)
			continue
		}
		position := 0.0
		if count > 1 {
			position = float64(index) / float64(count-1)
		}
		red := interpolate(startRed, endRed, position)
		green := interpolate(startGreen, endGreen, position)
		blue := interpolate(startBlue, endBlue, position)
		fmt.Fprintf(&output, "\033[38;2;%d;%d;%dm%c", red, green, blue, char)
		index++
	}
	output.WriteString(reset)
	return output.String()
}

func parseHex(value string) (uint64, uint64, uint64, bool) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(value) != 6 {
		return 0, 0, 0, false
	}
	red, errRed := strconv.ParseUint(value[0:2], 16, 8)
	green, errGreen := strconv.ParseUint(value[2:4], 16, 8)
	blue, errBlue := strconv.ParseUint(value[4:6], 16, 8)
	return red, green, blue, errRed == nil && errGreen == nil && errBlue == nil
}

func interpolate(start, end uint64, position float64) int {
	return int(float64(start) + (float64(end)-float64(start))*position + 0.5)
}
