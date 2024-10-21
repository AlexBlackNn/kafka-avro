# Введение в Apache Kafka. 

## Зачем нужны брокеры сообщений

Брокеры сообщений, играют важную роль в архитектуре современных распределенных систем. 
Они обеспечивают асинхронную передачу данных между различными компонентами системы, что позволяет 
улучшить масштабируемость, надежность и производительность.

Предположим, у нас есть интернет-магазин с несколькими компонентами:

    Веб-сервер: Обрабатывает запросы пользователей.
    Система управления заказами: Обрабатывает заказы и управляет их состоянием.
    Система управления запасами: Отслеживает наличие товаров на складе.
    Система уведомлений: Отправляет уведомления пользователям (например, по электронной почте или SMS).

Сценарий: Обработка заказа
    Пользователь делает заказ: Когда пользователь завершает покупку, веб-сервер получает запрос на создание заказа.
    Отправка сообщения о новом заказе: Вместо того чтобы сразу обрабатывать заказ, веб-сервер отправляет сообщение о новом заказе в брокер сообщений.

json

    {
        "orderId": "12345",
        "userId": "67890",
        "items": [
            {"productId": "535", "quantity": 1, "price": 300},
            {"productId": "125", "quantity": 2, "price": 100}
        ],
        "totalPrice": 400.00
    }

    Обработка заказа: Система управления заказами подписана на сообщения о новых заказах. Она получает сообщение, обрабатывает его (например, проверяет наличие товаров, создает запись в базе данных) и обновляет статус заказа.
    Обновление запасов: После успешной обработки заказа система управления заказами отправляет сообщение в брокер о том, что товары были заказаны. Система управления запасами подписана на эти сообщения и обновляет количество доступных товаров на складе.
    Отправка уведомлений: После того как заказ был успешно обработан, система управления заказами отправляет сообщение о завершении заказа в брокер. Система уведомлений подписана на эти сообщения и отправляет пользователю уведомление о том, что заказ был успешно оформлен.
    Мониторинг и аналитика: Все сообщения могут быть записаны в брокер сообщений для последующего анализа. Это позволяет собирать статистику о продажах, отслеживать популярные товары и улучшать бизнес-процессы.

Давайте попробуем сделать эмитацию такого брокера сообщений с использованием файла. 

Для начала напишем клиентский код, который отвечает за запись сообщениий в наш импровизируемый брокер. Такой клиентский код называется Продюсером (Producer) он публикуют события в брокер. 


```go
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

	wg.Add(1)
	// запустим продюсера в горутине
	go produce(orders, &wg)
	wg.Wait()

}

// produce - имитриуем продьюсера
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
		log.Printf("Заказ c ID %s успешно записан в файл.", order.OrderID)
	}
}

```

Запустите код, будет  создан файл  с именем `orders.json`, в который будут записаны данные. 

```json
{"order_id":"0001","user_id":"00001","items":[{"product_id":"535","quantity":1,"price":300},{"product_id":"125","quantity":2,"price":100}],"total_price":500}
{"order_id":"0002","user_id":"00002","items":[{"product_id":"035","quantity":7,"price":100},{"product_id":"525","quantity":1,"price":500}],"total_price":1200}
{"order_id":"0003","user_id":"00003","items":[{"product_id":"035","quantity":10,"price":100},{"product_id":"525","quantity":2,"price":500}],"total_price":2000}

```


В терминале должно появиться 

```
2024/10/21 11:54:32 Заказ c ID 0001 успешно записан в файл.
2024/10/21 11:54:32 Заказ c ID 0002 успешно записан в файл.
2024/10/21 11:54:32 Заказ c ID 0003 успешно записан в файл.
```

Теперь перейдем к  клиентскому приложению, которое отвечает за чтение сообщениий из нашего импровизированного брокера. Такой клиентский код называется Консьюмером (Consumer) он читает события из брокера.

```go
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

	wg.Add(1)
	// запустим консбюмер в горутине
	go consume(&wg)
	wg.Wait()

}

func consume(wg *sync.WaitGroup) {
	defer wg.Done()

	// Открываем файл для чтения
	file, err := os.OpenFile("orders.json", os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalln("Ошибка при открытии файла:", err)
	}
	defer file.Close()

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
	}
}

```
В терминале появятся заказы:
```
2024/10/21 11:52:13 Прочитанный заказ: {Offset:0 OrderID:0001 UserID:00001 Items:[{ProductID:535 Quantity:1 Price:300} {ProductID:125 Quantity:2 Price:100}] TotalPrice:500}
2024/10/21 11:52:13 Прочитанный заказ: {Offset:0 OrderID:0002 UserID:00002 Items:[{ProductID:035 Quantity:7 Price:100} {ProductID:525 Quantity:1 Price:500}] TotalPrice:1200}
2024/10/21 11:52:13 Прочитанный заказ: {Offset:0 OrderID:0003 UserID:00003 Items:[{ProductID:035 Quantity:10 Price:100} {ProductID:525 Quantity:2 Price:500}] TotalPrice:2000}
```



Отличие от RabbitMQ и других брокеров сообщений (push, pull, персистентность)


Основная терминология (Пишем Kafkа на коленке) 

Базовые компоненты

Порядок установки и настройки кластера локально с использованием Docker


```go
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
	orders := []*Order{
		&Order{
			Offset:  1,
			OrderID: "0001",
			UserID:  "00001",
			Items: []Item{
				{ProductID: "535", Quantity: 1, Price: 300},
				{ProductID: "125", Quantity: 2, Price: 100},
			},
			TotalPrice: 500.00,
		},
		&Order{
			Offset:  2,
			OrderID: "0002",
			UserID:  "00002",
			Items: []Item{
				{ProductID: "035", Quantity: 7, Price: 100},
				{ProductID: "525", Quantity: 1, Price: 500},
			},
			TotalPrice: 1200.00,
		},
		&Order{
			Offset:  3,
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

	for {
		// Открываем файл для чтения
		file, err := os.OpenFile("orders.json", os.O_CREATE|os.O_RDONLY, 0644)
		if err != nil {
			log.Fatalln("Ошибка при открытии файла:", err)
		}
		defer file.Close()

		// Читаем файл построчно
		var order Order
		decoder := json.NewDecoder(file)

		if err := decoder.Decode(&order); err != nil {
			if err.Error() == "EOF" {
				break // Достигнут конец файла
			}
			log.Fatalln("Ошибка при декодировании JSON:", err)
		}
		log.Printf("Прочитанный заказ: %+v\n", order)
	}
}

```