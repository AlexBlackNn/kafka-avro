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

// Количество заказов для генерции
const productNum = 3

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

type orderGenerator struct {
	orderId   int
	numOrders int
}

// generateOrders функция генератор заказов
func (og *orderGenerator) generate() []*Order {
	// Инициализация генератора случайных чисел, чтобы были повторяемые результаты
	rand.Seed(0)

	orders := make([]*Order, og.numOrders)
	k := 0
	current_order := og.orderId
	for og.orderId < current_order+og.numOrders {

		// Формируем данные

		orderID := fmt.Sprintf("%04d", og.orderId)
		userID := fmt.Sprintf("%05d", rand.Intn(100000))

		// Генерируем случайное кол-во товаров от 1 до 5
		numItems := rand.Intn(5) + 1
		items := make([]Item, numItems)
		totalPrice := 0.0

		for j := 0; j < numItems; j++ {
			productID := fmt.Sprintf("%03d", rand.Intn(1000))
			quantity := rand.Intn(10) + 1
			price := float64(rand.Intn(1000)) + rand.Float64()

			items[j] = Item{
				ProductID: productID,
				Quantity:  quantity,
				Price:     price,
			}
			totalPrice += price * float64(quantity) // Суммируем общую стоимость
		}
		orders[k] = &Order{
			OrderID:    orderID,
			UserID:     userID,
			Items:      items,
			TotalPrice: totalPrice,
		}
		k++
		og.orderId++
	}

	return orders
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	var lock sync.Mutex

	// продюсер и консьюмер работают параллельно
	go produce(&wg, &lock)
	go consume(&wg, &lock)

	wg.Wait()

}

// produce - имулирует работу продюсера
func produce(wg *sync.WaitGroup, lock *sync.Mutex) {
	defer wg.Done()
	ordGenerator := orderGenerator{numOrders: productNum}
	for {
		// Получаем слайс случайных заказов

		orders := ordGenerator.generate()

		var jsons []byte

		for _, order := range orders {
			// Сериализуем объект заказа в JSON
			orderJSON, err := json.Marshal(order)
			if err != nil {
				log.Fatalln("Ошибка при сериализации в JSON:", err)
			}

			// Добавляем новую строку для удобства чтения
			orderJSON = append(orderJSON, '\n')
			jsons = append(jsons, orderJSON...)
		}

		// защищаем общие данные и участки кода от одновременного доступа.
		lock.Lock()
		// Открываем файл в режиме добавления
		file, err := os.OpenFile("orders.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln("Ошибка при открытии файла:", err)
		}

		if _, err := file.Write(jsons); err != nil {
			log.Fatalln("Ошибка при записи в файл:", err)
		}
		log.Println("Заказы успешно записаны в файл.")

		// Закрываем файл и снимаем блокировку.
		file.Close()
		lock.Unlock()
		// Делаем паузу перед следующим добавлением данных
		time.Sleep(2 * time.Second)
	}
}

// consume - имулирует работу консьюмера
func consume(wg *sync.WaitGroup, lock *sync.Mutex) {
	defer wg.Done()

	for {
		// защищаем общие данные и участки кода от одновременного доступа.
		lock.Lock()
		// Открываем файл для чтения
		file, err := os.OpenFile("orders.json", os.O_CREATE|os.O_RDONLY, 0644)
		if err != nil {
			log.Fatalln("Ошибка при открытии файла:", err)
		}

		// Открываем файл для чтения офсета
		offsetFile, err := os.OpenFile("offset.txt", os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			log.Fatalln("Ошибка при открытии файла офсета:", err)
		}

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
		// Закрываем файлы и снимаем блокировку.
		file.Close()
		offsetFile.Close()

		lock.Unlock()
		time.Sleep(2 * time.Second)
	}
}
