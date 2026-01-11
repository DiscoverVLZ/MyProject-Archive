package main

import (
	"image"
	"image/color"
	"image/draw"
	"os"

	"golang.org/x/image/bmp"
)

func createIcon() {
	// Создаем яркое изображение 256x256
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	
	// Градиентный фон
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			// Сине-фиолетовый градиент
			r := uint8(100 + x/2)
			g := uint8(50 + y/3)
			b := uint8(200 - x/4)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	
	// Рисуем значок архива (папка с глазком)
	drawFolderIcon(img)
	
	// Сохраняем как BMP (потом конвертируем в ICO через онлайн или resedit)
	f, _ := os.Create("assets/icon.bmp")
	defer f.Close()
	bmp.Encode(f, img)
}

func drawFolderIcon(img *image.RGBA) {
	// Рисуем папку
	folderColor := color.RGBA{255, 215, 0, 255} // Золотой
	eyeColor := color.RGBA{255, 255, 255, 255}  // Белый
	pupilColor := color.RGBA{0, 100, 200, 255}  // Синий
	
	// Контур папки
	for y := 80; y < 180; y++ {
		for x := 60; x < 196; x++ {
			if (y >= 80 && y <= 100) || // Верхняя часть
				(x == 60 || x == 195) || // Боковые стороны
				(y == 179) {             // Нижняя часть
				if x >= 70 && x <= 185 {
					img.Set(x, y, folderColor)
				}
			}
		}
	}
	
	// Заполнение папки
	for y := 100; y < 179; y++ {
		for x := 61; x < 195; x++ {
			img.Set(x, y, folderColor)
		}
	}
	
	// Глаз на папке
	drawEye(img, 128, 130, 30, eyeColor, pupilColor)
}

func drawEye(img *image.RGBA, cx, cy, radius int, eye, pupil color.Color) {
	// Внешний круг (глаз)
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				img.Set(cx+x, cy+y, eye)
			}
		}
	}
	
	// Зрачок
	pupilRadius := radius / 3
	for y := -pupilRadius; y <= pupilRadius; y++ {
		for x := -pupilRadius; x <= pupilRadius; x++ {
			if x*x+y*y <= pupilRadius*pupilRadius {
				img.Set(cx+x+5, cy+y, pupil) // Смещаем для эффекта наблюдения
			}
		}
	}
}