package photosave

import (
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// Сохраняем фотки
func SaveReader(d SaveObj) (fname string, err error) {
	// Читаем картинку
	img, format, err := image.Decode(d.Reader)
	if err != nil {
		log.Println("[error]", err)
		return
	}

	// Если надо добавить водяной знак
	if d.WatermarkPath != "" {
		img = setWatermark(d, img)
	}

	// Если надо сделать случайное имя файла
	if d.FileNameRnd {
		d.FileName = getRandomName(d, format)
	}

	// Имя фотки
	fname = d.FileName + "." + format

	// Создаем файл фотки
	out, err := os.Create(d.FilePath + fname)
	if err != nil {
		log.Println("[error]", err)
		return
	}
	// Закрываем файл фотки
	defer out.Close()

	// Возвращаем указатель на начало фотки
	_, err = d.Reader.Seek(0, 0)
	if err != nil {
		log.Println("[error]", err)
		return
	}

	// Если есть водяной знак, то сохраняем фотку с конвертированием
	if d.WatermarkPath != "" {
		if format == "png" {
			err = png.Encode(out, img)
		} else if format == "gif" {
			err = gif.Encode(out, img, nil)
		} else {
			err = jpeg.Encode(out, img, nil)
		}
	} else {
		// Копируем фотку в файл
		_, err = io.Copy(out, d.Reader)
	}
	if err != nil {
		log.Println("[error]", err)
		return
	}

	// Оптимизируем фотку
	imgopti(format, d.FilePath+fname)

	return
}

// Читаем данные и сохраняем фотки
func Save(d SaveObj) (fname string, err error) {
	// Читаем фотку из http
	file, _, err := d.R.FormFile(d.ValueName)
	if err != nil {
		if err.Error() != "http: no such file" {
			log.Println("[error]", err)
		}
		return
	}
	defer file.Close()

	d.Reader = file

	return SaveReader(d)
}

// Оптимизируем фотку
func imgopti(format, fname string) {
	// Формируем команду
	var cmd *exec.Cmd
	if format == "jpeg" {
		cmd = exec.Command("/usr/bin/jpegoptim", "--strip-all", fname)
	} else if format == "png" {
		cmd = exec.Command("/usr/bin/optipng", fname)
	}

	// Выполняем команду
	ans, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("[error]", err, string(ans))
	}
}

// Получаем случайное имя для файла
func getRandomName(d SaveObj, format string) (fname string) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 10000; i++ {
		// Формируем имя и проверяем его
		name := strconv.FormatInt(time.Now().UnixNano(), 16) + strconv.Itoa(rand.Int())
		if _, err := os.Stat(d.FilePath + name + "." + format); os.IsNotExist(err) {
			fname = name
			break
		}
		// Если слишком много попыток подобрать имя
		if i == 9999 {
			log.Fatalln("[fatal]", "can't create name")
		}
	}

	return
}

// Добавляем watermark
func setWatermark(d SaveObj, img image.Image) image.Image {
	// Открываем файл с водяным знаком
	wmb, err := os.Open(d.WatermarkPath)
	if err != nil {
		log.Println("[error]", err)
		return img
	}

	// Читаем водяной знак
	watermark, err := png.Decode(wmb)
	if err != nil {
		log.Println("[error]", err)
		return img
	}
	defer wmb.Close()

	// Получаем объект для определения размера фотки
	b := img.Bounds()

	// Определяем отступ
	var x, y int
	if d.WatermarkX != 0 {
		if d.WatermarkXFromMax {
			x = b.Max.X + d.WatermarkX
		} else {
			x = d.WatermarkX
		}
	}
	if d.WatermarkY != 0 {
		if d.WatermarkYFromMax {
			y = b.Max.Y + d.WatermarkY
		} else {
			y = d.WatermarkY
		}
	}

	// Ставим отступ
	offset := image.Pt(x, y)

	// Добавляем водяной знак
	m := image.NewRGBA(b)
	draw.Draw(m, b, img, image.ZP, draw.Src)
	draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)

	return m
}
