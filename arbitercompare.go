package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"
)

type CompareResult struct {
	Name       string  `json:"string"`
	Comparison float32 `json:"float"`

	ContainsLast bool `json:"boolean"`
	ContainsCurr bool `json:"boolean"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func CollectWebsites2(path string, websiteNames *map[string]string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for i := range entries {
		entry := entries[i]

		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "http") {
			name := entry.Name()

			name = strings.ReplaceAll(name, "-", ":")
			name = strings.ReplaceAll(name, "_", "/")
			name = strings.ReplaceAll(name, ".png", "")

			if (*websiteNames)[name] == "" {
				(*websiteNames)[name] = entry.Name()
			}
		}
	}

	return nil
}

func CollectWebsites(lastRunPath string, currentRunPath string) (map[string]string, error) {
	websiteNames := make(map[string]string, 0)

	err := CollectWebsites2(lastRunPath, &websiteNames)
	if err != nil {
		return nil, err
	}
	err = CollectWebsites2(currentRunPath, &websiteNames)
	if err != nil {
		return nil, err
	}

	return websiteNames, nil
}

func OpenImage(path string) (image.Image, error) {
	_, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	img, err := png.Decode(file)

	if err != nil {
		return nil, err
	}

	return img, nil
}

func main() {
	lastRun := flag.Int64("last-run", 0, "The last run to compare against")
	currentRun := flag.Int64("current-run", 1, "The current run to compare against")

	flag.Parse()

	lastRunPath := "./data/" + strconv.FormatInt(*lastRun, 10) + "/"
	currentRunPath := "./data/" + strconv.FormatInt(*currentRun, 10) + "/"

	websiteNames, err := CollectWebsites(lastRunPath, currentRunPath)

	if err != nil {
		fmt.Printf("An error occured: %s", err.Error())

		return
	}

	comparisonResults := make(map[string]CompareResult, 0)

	for name := range websiteNames {
		fileName := websiteNames[name]

		containsLast := true
		containsCurr := true
		var comparison float32 = 0.0

		lastRunFile, err := OpenImage(lastRunPath + fileName)

		if err != nil {
			fmt.Printf("An error occured on last run file: %s", err.Error())
			containsLast = false
		}

		currentRunFile, err := OpenImage(currentRunPath + fileName)

		if err != nil {
			fmt.Printf("An error occured on current run file: %s", err.Error())
			containsCurr = false
		}

		if containsCurr && containsLast {
			unequalPixels := 0

			minX := min(lastRunFile.Bounds().Max.X, currentRunFile.Bounds().Max.X)
			minY := min(lastRunFile.Bounds().Max.Y, currentRunFile.Bounds().Max.Y)
			maxX := max(lastRunFile.Bounds().Max.X, currentRunFile.Bounds().Max.X)
			maxY := max(lastRunFile.Bounds().Max.Y, currentRunFile.Bounds().Max.Y)

			for x := 0; x < minX; x++ {
				for y := 0; y < minY; y++ {
					r1, g1, b1, a1 := lastRunFile.At(x, y).RGBA()
					r2, g2, b2, a2 := currentRunFile.At(x, y).RGBA()

					if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
						unequalPixels++
					}
				}
			}

			diffX := math.Abs(float64(currentRunFile.Bounds().Max.X - lastRunFile.Bounds().Max.X))
			diffY := math.Abs(float64(currentRunFile.Bounds().Max.Y - lastRunFile.Bounds().Max.Y))
			diffPixels := diffX*float64(minY) + float64(maxX)*diffY

			unequalPixels += int(diffPixels)

			comparison = float32(unequalPixels) / (float32(maxX) * float32(maxY))
		} else {
			if containsCurr {
				comparison = 1.0
			}

			if containsLast {
				comparison = 0.0
			}
		}

		comparisonResults[name] = CompareResult{Name: name, Comparison: comparison, ContainsLast: containsLast, ContainsCurr: containsCurr}
	}

	comparisonJson, err := json.Marshal(comparisonResults)
	if err != nil {
		fmt.Printf("An error occured: %s", err.Error())
		return
	}

	err = os.WriteFile(currentRunPath+"comparison.json", comparisonJson, 0777)
	if err != nil {
		fmt.Printf("An error occured: %s", err.Error())
		return
	}

	//GENERATE HTML FILE
}
