package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"sync"
	"time"
)

const (
	NumWorkers = 4 // Nombre de travailleurs pour la répartition des tâches
)

func main() {
	filePath := "input.jpg"
	outputPath := "output.jpg"

	// Charger l'image d'entrée
	inputImage, err := loadImage(filePath)
	if err != nil {
		log.Fatal("Impossible de charger l'image d'entrée :", err)
	}

	// Créer une copie de l'image d'entrée pour chaque méthode de répartition des tâches
	inputImageWG := cloneImage(inputImage)
	inputImageCh := cloneImage(inputImage)

	// Utiliser WaitGroup pour répartir les tâches
	startTimeWG := time.Now()
	processImageWithWaitGroup(inputImageWG)
	elapsedTimeWG := time.Since(startTimeWG)

	// Utiliser Channel pour répartir les tâches
	startTimeCh := time.Now()
	processImageWithChannel(inputImageCh)
	elapsedTimeCh := time.Since(startTimeCh)

	// Enregistrer les images traitées
	saveImage(outputPath, inputImageWG)
	saveImage(outputPath, inputImageCh)

	// Afficher les résultats de performance
	fmt.Println("Temps écoulé avec WaitGroup:", elapsedTimeWG)
	fmt.Println("Temps écoulé avec Channel:", elapsedTimeCh)
}

func loadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func saveImage(filePath string, img image.Image) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return err
	}

	return nil
}

func cloneImage(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	clone := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			clone.Set(x, y, src.At(x, y))
		}
	}

	return clone
}

func processImageWithWaitGroup(img *image.RGBA) {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Créer un WaitGroup pour suivre les goroutines
	var wg sync.WaitGroup
	wg.Add(NumWorkers)

	// Calculer le nombre de pixels à traiter par goroutine
	pixelsPerWorker := width * height / NumWorkers

	for i := 0; i < NumWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			startPixel := workerID * pixelsPerWorker
			endPixel := startPixel + pixelsPerWorker

			// Appliquer le filtre de grisaille aux pixels attribués
			for pixel := startPixel; pixel < endPixel; pixel++ {
				x := pixel % width
				y := pixel / width

				r, g, b, _ := img.At(x, y).RGBA()
				gray := uint8((r + g + b) / 3)
				grayColor := color.RGBA{gray, gray, gray, 255}

				img.Set(x, y, grayColor)
			}
		}(i)
	}

	// Attendre que toutes les goroutines se terminent
	wg.Wait()
}

func processImageWithChannel(img *image.RGBA) {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Créer un channel pour les tâches
	taskCh := make(chan int)

	// Créer un WaitGroup pour suivre les goroutines
	var wg sync.WaitGroup
	wg.Add(NumWorkers)

	// Lancer les goroutines de traitement
	for i := 0; i < NumWorkers; i++ {
		go func() {
			defer wg.Done()

			// Traitement des tâches du channel jusqu'à ce qu'il soit fermé
			for pixel := range taskCh {
				x := pixel % width
				y := pixel / width

				r, g, b, _ := img.At(x, y).RGBA()
				gray := uint8((r + g + b) / 3)
				grayColor := color.RGBA{gray, gray, gray, 255}

				img.Set(x, y, grayColor)
			}
		}()
	}

	// Envoyer les tâches au channel
	for pixel := 0; pixel < width*height; pixel++ {
		taskCh <- pixel
	}

	// Fermer le channel et attendre que toutes les goroutines se terminent
	close(taskCh)
	wg.Wait()
}
