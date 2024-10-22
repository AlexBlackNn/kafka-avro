package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"golang.org/x/exp/rand"
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

func generateOrders(numOrders int) []*Order {
	rand.Seed(0) // Инициализация генератора случайных чисел
	orders := make([]*Order, numOrders)

	for i := 0; i < numOrders; i++ {
		orderID := fmt.Sprintf("%04d", i+1)              // Форматирование OrderID с ведущими нулями
		userID := fmt.Sprintf("%05d", rand.Intn(100000)) // Случайный UserID от 00000 до 99999

		// Генерация случайных товаров
		numItems := rand.Intn(5) + 1 // Случайное количество товаров от 1 до 5
		items := make([]Item, numItems)
		totalPrice := 0.0

		for j := 0; j < numItems; j++ {
			productID := fmt.Sprintf("%03d", rand.Intn(1000))  // Случайный ProductID от 000 до 999
			quantity := rand.Intn(10) + 1                      // Случайное количество от 1 до 10
			price := float64(rand.Intn(1000)) + rand.Float64() // Случайная цена от 0 до 1000

			items[j] = Item{
				ProductID: productID,
				Quantity:  quantity,
				Price:     price,
			}
			totalPrice += price * float64(quantity) // Суммируем общую стоимость
		}

		orders[i] = &Order{
			OrderID:    orderID,
			UserID:     userID,
			Items:      items,
			TotalPrice: totalPrice,
		}
	}

	return orders
}

func main() {
	var wg sync.WaitGroup
	orders := generateOrders(10)

	wg.Add(2)

	var lock sync.Mutex
	// продюсер и консьюмер работают параллельно

	go produce(orders, &wg, &lock)
	go consume(&wg, &lock)

	wg.Wait()

}

func produce(orders []*Order, wg *sync.WaitGroup, lock *sync.Mutex) {
	defer wg.Done()
	for {
		for _, order := range orders {
			// защищаем общие данные и участки кода (такие участки называются критической секцией) от одновременного доступа.
			lock.Lock()
			// Открываем файл в режиме добавления
			file, err := os.OpenFile("orders.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalln("Ошибка при открытии файла:", err)
			}

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

			file.Close()
			lock.Unlock()
		}
		time.Sleep(2 * time.Second)
	}
}

func consume(wg *sync.WaitGroup, lock *sync.Mutex) {
	defer wg.Done()

	for {
		lock.Lock()
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

		// Читаем файл и обрабатываем его
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
		lock.Unlock()
		time.Sleep(2 * time.Second)
	}
}

// func consume(wg *sync.WaitGroup, lock *sync.Mutex) {
// 	defer wg.Done()
// 	lock.Lock()
// 	defer lock.Unlock()

// 	// --- Работа с файлами - подготовка к чтению данных ---

// 	// Открываем файл для чтения
// 	file, err := os.OpenFile("orders.json", os.O_CREATE|os.O_RDONLY, 0644)
// 	if err != nil {
// 		log.Fatalln("Ошибка при открытии файла:", err)
// 	}
// 	defer file.Close()

// 	// Открываем файл для чтения офсета
// 	offsetFile, err := os.OpenFile("offset.txt", os.O_CREATE|os.O_RDWR, 0644)
// 	if err != nil {
// 		log.Fatalln("Ошибка при открытии файла офсета:", err)
// 	}
// 	defer offsetFile.Close()

// 	// Читаем офсет из файла
// 	var offset int64
// 	if _, err := fmt.Fscanf(offsetFile, "%d", &offset); err != nil && err.Error() != "EOF" {
// 		log.Fatalln("Ошибка при чтении офсета:", err)
// 	}

// 	// Устанавливаем офсет для чтения
// 	_, err = file.Seek(offset, io.SeekStart)
// 	if err != nil {
// 		log.Fatalln("Ошибка при установке офсета:", err)
// 	}

// 	// --- Работа с данными - чтение данных из файла ---

// 	var order Order
// 	decoder := json.NewDecoder(file)

// 	// Декодируем один JSON-объект
// 	if err := decoder.Decode(&order); err != nil {
// 		if err == io.EOF {
// 			log.Println("Достигнут конец файла")
// 			return // Достигнут конец файла
// 		}
// 		log.Fatalln("Ошибка при декодировании JSON:", err)
// 	}
// 	log.Printf("Прочитанный заказ: %+v\n", order)

// 	// --- Работа с файлами - Запись офсета ---
// 	// Обновляем офсет
// 	offset, err = file.Seek(0, io.SeekCurrent)
// 	fmt.Println(offset)
// 	if err != nil {
// 		log.Fatalln("Ошибка при получении текущего офсета:", err)
// 	}

// 	// Записываем новый офсет в файл
// 	err = offsetFile.Truncate(0) // Очищаем файл
// 	if err != nil {
// 		log.Fatalln("Ошибка при очистке файла текущего офсета:", err)
// 	}
// 	offsetFile.Seek(0, io.SeekStart) // Возвращаемся в начало файла
// 	if _, err := fmt.Fprintf(offsetFile, "%d", offset); err != nil {
// 		log.Fatalln("Ошибка при записи офсета:", err)
// 	}
// }
