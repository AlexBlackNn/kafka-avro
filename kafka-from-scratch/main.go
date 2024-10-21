package main

import (
	"encoding/json"
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
	Offset     int     `json:"offset"`
	OrderID    string  `json:"order_id"`
	UserID     string  `json:"user_id"`
	Items      []Item  `json:"items"`
	TotalPrice float64 `json:"total_price"`
}

func main() {
	var wg sync.WaitGroup

	// Создаем объект заказа
	order := &Order{
		Offset:  1,
		OrderID: "12345",
		UserID:  "67890",
		Items: []Item{
			{ProductID: "535", Quantity: 1, Price: 300},
			{ProductID: "125", Quantity: 2, Price: 100},
		},
		TotalPrice: 400.00,
	}

	wg.Add(1)
	go produce(order, &wg)
	wg.Wait()
}

func produce(order *Order, wg *sync.WaitGroup) {

	defer wg.Done()

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
