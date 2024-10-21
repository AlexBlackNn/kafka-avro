package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// Item - структура описывающая продукт в заказе
type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Order - структура описывающая заказ с продуктами
type Order struct {
	OrderID    string  `json:"order_id"`
	UserID     string  `json:"user_id"`
	Items      []Item  `json:"items"`
	TotalPrice float64 `json:"total_price"`
}

func main() {
	var wg sync.WaitGroup
	orders := []*Order{
		{
			OrderID: "0001",
			UserID:  "00001",
			Items: []Item{
				{ProductID: "535", Quantity: 1, Price: 300},
				{ProductID: "125", Quantity: 2, Price: 100},
			},
			TotalPrice: 500.00,
		},
		{
			OrderID: "0002",
			UserID:  "00002",
			Items: []Item{
				{ProductID: "035", Quantity: 7, Price: 100},
				{ProductID: "525", Quantity: 1, Price: 500},
			},
			TotalPrice: 1200.00,
		},
		{
			OrderID: "0003",
			UserID:  "00003",
			Items: []Item{
				{ProductID: "035", Quantity: 10, Price: 100},
				{ProductID: "525", Quantity: 2, Price: 500},
			},
			TotalPrice: 2000.00,
		},
	}

	wg.Add(2)
	// продюсер и консьюмер работают параллельно
	go produce(orders, &wg)
	go consume(&wg)

	wg.Wait()

}

func produce(orders []*Order, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, order := range orders {
		// Открываем файл в режиме добавления
		file, err := os.OpenFile("orders.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln("Ошибка при открытии файла:", err)
		}
		defer file.Close()

		// Сериализуем объект заказа в JSON
		orderJSON, err := json.Marshal(order)
		if err != nil {
			log.Fatalln("Ошибка при сериализации в JSON:", err)
		}

		// Добавляем новую строку для удобства чтения
		orderJSON = append(orderJSON, '\n')

		// Записываем JSON в файл
		if _, err := file.Write(orderJSON); err != nil {
			log.Fatalln("Ошибка при записи в файл:", err)
		}
		log.Println("Заказ успешно записан в файл.")
	}
}

func consume(wg *sync.WaitGroup) {
	defer wg.Done()

	// Открываем файл для чтения
	file, err := os.OpenFile("orders.json", os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalln("Ошибка при открытии файла:", err)
	}
	defer file.Close()

	// Открываем файл для чтения офсета
	offsetFile, err := os.OpenFile("offset.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalln("Ошибка при открытии файла офсета:", err)
	}
	defer offsetFile.Close()

	// Читаем офсет из файла
	var offset int64
	if _, err := fmt.Fscanf(offsetFile, "%d", &offset); err != nil && err.Error() != "EOF" {
		log.Fatalln("Ошибка при чтении офсета:", err)
	}

	// Устанавливаем офсет для чтения
	_, err = file.Seek(offset, 0)
	if err != nil {
		log.Fatalln("Ошибка при установке офсета:", err)
	}

	// Читаем файл построчно
	var order Order
	decoder := json.NewDecoder(file)
	for {
		if err := decoder.Decode(&order); err != nil {
			if err.Error() == "EOF" {
				break // Достигнут конец файла
			}
			log.Fatalln("Ошибка при декодировании JSON:", err)
		}
		log.Printf("Прочитанный заказ: %+v\n", order)

		// Обновляем офсет
		offset, err = file.Seek(0, io.SeekCurrent)
		if err != nil {
			log.Fatalln("Ошибка при получении текущего офсета:", err)
		}

		// Записываем новый офсет в файл
		offsetFile.Truncate(0) // Очищаем файл
		offsetFile.Seek(0, 0)  // Возвращаемся в начало файла
		if _, err := fmt.Fprintf(offsetFile, "%d", offset); err != nil {
			log.Fatalln("Ошибка при записи офсета:", err)
		}
	}
}
