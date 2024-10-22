# Введение в Apache Kafka. 

Брокеры сообщений, играют важную роль в архитектуре современных распределенных систем. 
Они обеспечивают асинхронную передачу данных между различными компонентами системы, что позволяет 
улучшить масштабируемость, надежность и производительность, а также позволяют реализовать [kappa-архитектуру](https://bigdataschool.ru/blog/kappa-architecture.html)
и делают связность между сервисами более слабой.

Предположим, у нас есть интернет-магазин с несколькими компонентами:

  1. Веб-сервер: Обрабатывает запросы пользователей.
  2. Система управления заказами: Обрабатывает заказы и управляет их состоянием.
  3. Система управления запасами: Отслеживает наличие товаров на складе.
  4. Cистема уведомлений: Отправляет уведомления пользователям (например, по электронной почте или SMS).

Рассмотрим сценарий обработки заказа
    Пользователь делает заказ. Когда пользователь завершает покупку, веб-сервер получает запрос на создание заказа. 
    Вместо того чтобы сразу обрабатывать заказ, веб-сервер отправляет сообщение о новом заказе в брокер сообщений.

```json
  {
      "orderId": "12345",
      "userId": "67890",
      "items": [
          {"productId": "535", "quantity": 1, "price": 300},
          {"productId": "125", "quantity": 2, "price": 100}
      ],
      "totalPrice": 400.00
  }
```
Система управления заказами подписана на сообщения о новых заказах. Она получает сообщение, обрабатывает его (например, проверяет наличие товаров, создает запись в базе данных) и обновляет статус заказа. После успешной обработки заказа система управления заказами отправляет сообщение в брокер о том, что товары были заказаны. Система управления запасами подписана на эти сообщения и обновляет количество доступных товаров на складе. После того как заказ был успешно обработан, система управления заказами отправляет сообщение о завершении заказа в брокер. Система уведомлений подписана на эти сообщения и отправляет пользователю уведомление о том, что заказ был успешно оформлен.

## Базовые понятия 
Давайте попробуем сделать имитацию такого брокера сообщений самостоятельно с использованием обычного файла и поговорим о терминалогии. 

Для начала напишем клиентский код, который отвечает за запись сообщениий в наш импровизируемый брокер. Такой клиентский код называется *Продюсером (Producer)* он публикуют события в брокер. 


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
  // Фиктивные заказы, которые будем отправлять в брокер
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
	// запустим продюсер в горутине
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

Запустите код, будет  создан файл  с именем `orders.json`, в который будут записаны данные: 

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

Теперь перейдем к  клиентскому приложению, которое отвечает за чтение сообщениий из нашего импровизированного брокера. Такой клиентский код называется *Консьюмером (Consumer)* он читает события из брокера.

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
	// запустим консьюмер в горутине
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


## Задание. 
Доработайте код: объедените код читателя и писателя. Сделайте, чтобы читатель и писатель могли работать параллельно. А если читатель вдруг по какой-то причине будет выключен, необходимо, чтобы он начал читать с того самого места (*Оффсет*), на котором он остановился. Для хранения оффсета используйте файл

<br>
<details> 
<summary>Подсказка: работа с офсетом консьюмера  (нажмите, чтобы увидеть код)</summary>

```go
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

	// 
 	// Код работы с данными из импровизированного брокера
	//
	
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
```

</details> 

<br>
<details> 
<summary>Ознакомиться с решением (нажмите, чтобы увидеть код)</summary>

```go
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

```
</details> 

<br>
В нашем импровизированном примере с созданием брокера сообщений в одном файле мы столкнулись с ограничениями, связанными с масштабируемостью и устойчивостью системы. Понятно, что брокеры сообщений устроены куда сложнее. Чтобы улучшить архитектуру, нам необходимо рассмотреть возможность распределенного хранения сообщений, что позволит запускать брокер на кластере машин. Это обеспечит высокую доступность и устойчивость к сбоям, так как в случае выхода из строя одной из машин, другие смогут продолжать обработку сообщений. Кроме того, распределенная система позволит значительно увеличить пропускную способность, что критично для современных приложений с высокими требованиями к производительности. В этом контексте стоит обратить внимание на Apache Kafka, который является мощным решением для организации потоковой передачи данных и управления сообщениями в распределенных системах, обеспечивая надежность и масштабируемость.
Тем не менее, пример выше позволяет вам ознакомиться с основными концепциями и принципами работы брокеров сообщений. 

<br>

*Давайте повторим базовые понятия:*

1. *Писатель (Producer)* — это клиентское приложение, которые публикует сообщения в брокер. 
2. *Читателем (Consumer)* называется приложение, которое подписывается на интересующие его сообщения и обрабатывает их.
3. *Оффсет(Offset)* - порядковый номер, который указывает на положение сообщения.

## Apache Kafka 

[Kafka](https://habr.com/ru/companies/piter/articles/352978/) была разработана в компании LinkedIn в 2011 году и с тех пор претерпела значительные улучшения. Сегодня Kafka представляет собой полноценную платформу, обеспечивающую избыточность, необходимую для хранения огромных объемов данных. Она предлагает шину сообщений с высокой пропускной способностью, позволяя в реальном времени обрабатывать все данные, проходящие через нее.

Однако, если свести Kafka к основным характеристикам, то это будет распределенный, горизонтально масштабируемый и отказоустойчивый лог сообщений.

Давайте разберем каждый из этих терминов и выясним, что они означают.

*Распределенной* называется система, которая работает на множестве машин, образующих кластер, при этом для конечного пользователя она выглядит как единый узел. В Kafka хранение, получение и рассылка сообщений организованы на различных узлах, называемых «брокерами». Основные преимущества такого подхода – высокая доступность и отказоустойчивость.


Прежде чем опишем, что такое *горизонтально масштабирование*, расскажем про вертикальное. Например, у нас есть традиционный сервер базы данных, который постепенно перестает справляться с растущей нагрузкой. Чтобы решить эту проблему, можно увеличить ресурсы (CPU, RAM, SSD) на сервере. Это и есть вертикальное масштабирование – добавление ресурсов к одной машине. Однако у этого подхода есть два серьезных недостатка:

1. Существуют физические ограничения, связанные с возможностями оборудования, и бесконечно увеличивать ресурсы невозможно.
2. Процесс масштабирования часто сопровождается простоями, что недопустимо для крупных компаний.

*Горизонтальная масштабируемость* решает ту же проблему, но с помощью подключения дополнительных машин. При добавлении новой машины не происходит простоев, и количество машин, которые можно включить в кластер, не имеет ограничений. Однако стоит отметить, что не все системы поддерживают горизонтальную масштабируемость: многие из них не предназначены для работы с кластерами (наш импровизированный "брокер" сообщений, как раз такой), а те, что предназначены, часто оказываются сложными в эксплуатации.

После достижения определенного порога горизонтальное масштабирование становится значительно более экономичным по сравнению с вертикальным.

Нераспределенные системы часто имеют единую точку отказа. Если единственный сервер вашей базы данных выйдет из строя, это приведет к серьезным проблемам. В отличие от этого, распределенные системы проектируются с учетом возможности адаптации к сбоям (*отказоустойчивость*). Например, кластер Kafka из пяти узлов продолжает функционировать, даже если два узла выходят из строя. Однако стоит отметить, что для обеспечения отказоустойчивости иногда приходится жертвовать производительностью: чем лучше система справляется с отказами, тем ниже ее общая производительность.

*Лог сообщений* представляет собой долговременную упорядоченную структуру данных, в которую можно только добавлять записи. Мы делали также с нашим файлом в продьюсере: 

```go
file, err := os.OpenFile("orders.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
```
Изменять или удалять данные в этом журнале нельзя. Информация считывается слева направо, что обеспечивает правильный порядок элементов.
Такие файлы-журналы объеденяются в *партиции*. А набор партиций составляет *топик*.  Топики в Kafka могут иметь одного или нескольких писателей/читателей, а также могут не иметь их вовсе. 

<br>
<details> 
<summary>Интересный факт (Нажми, чтобы прочитать)</summary>
    Операции чтения и записи выполняются за постоянное время O(1) (если известен ID записи), что значительно экономит время по сравнению с операциями O(log N) в других структурах данных, так как каждая операция перемещения головок на диске является затратной. При этом операции чтения и записи не блокируют друг друга. Эти два аспекта значительно увеличивают производительность, так как она не зависит от объема данных. Kafka работает одинаково эффективно, независимо от того, храните ли вы несколько килоБайт данных или сотни терабайт!

</details> 
<br>

Такое распределение данных критически важно для масштабирования, так как оно позволяет клиентским приложениям читать и записывать данные параллельно на разных брокерах. Когда новое событие публикуется в топик, оно добавляется в одну из его партиций.

Механизм выбора партиции зависит от писателя и настроек. Если при записи сообщения указан ключ, Kafka использует его для определения, в какую партицию будет отправлено сообщение. В этом случае все события с одним и тем же ключом, например, ID пользователя, всегда будут добавлены в одну и ту же партицию. Если ключ не указан, сообщения распределяются по всем партициям топика равномерно.

Kafka гарантирует, что любой потребитель для конкретного топика и партиции будет всегда считывать события в том порядке, в котором они были записаны. Когда новое событие публикуется в топик, оно добавляется в одну из его партиций. Kafka выбирает подходящую партицию с помощью ключа партиции, расчитывая хэщ. *Ключ партиции* (partition key) — это опциональный компонент сообщения, который используется для определения, в какую партицию будет отправлено сообщение. Если ключ партиции указан, Kafka гарантирует, что все сообщения с одним и тем же ключом будут направлены в одну и ту же партицию. 

Принцип расчета хэша от ключа зависит от языка разработки. В частности, Java-библиотека для продюсеров Kafka для вычисления хэш-значения ключа партиционирования использует 32-битный алгоритм хэширования [murmur2](https://ru.wikipedia.org/wiki/MurmurHash2). Это простая и быстрая хеш-функция общего назначения, разработанная Остином Эпплби. Она не является криптографически-безопасной и возвращает 32-разрядное беззнаковое число. Ее главными достоинствами является простота, хорошее распределение, мощный лавинный эффект, высокая скорость и сравнительно высокая устойчивость к коллизиям. 

Однако, далеко не все разработчики используют Java для создания продюсеров Kafka. Мы, в частонсти, используем язык программирования Go в наших примерах. В раде реализаций библиотек для работы с Kafka на ЯП Python, [Go](https://github.com/confluentinc/confluent-kafka-go) , .NET, C# может использоваться библиотека [librdkafka](https://github.com/confluentinc/librdkafka) (написанная на С++). Она предоставляет высокопроизводительную, легкую и многофункциональную реализацию протокола Kafka, позволяя клиентским приложениям взаимодействовать с кластерами Kafka.

librdkafka по умолчанию использует другой алгоритм хэширования — CRC32 — 32-битная циклическая проверка контрольной суммы (Cyclic Redundancy Check). Этот алгоритм представляет собой способ цифровой идентификации некоторой последовательности данных, который заключается в вычислении контрольного значения её циклического избыточного кода. Подробнее [тут](https://www.confluent.io/blog/standardized-hashing-across-java-and-non-java-producers/). Мы попозже еще вернемся к использованию Go библиотеки для организации связи между продюсером и писателем.  

Давайте перейдем к практике и закрепим полученне знания из теории. Для начала нам надо разобраться с инфраструктурой. 


## Порядок установки и настройки кластера локально с использованием Docker

Одним из самых простых способов запустить Kafka является использование Docker Compose.

1. Установите [Docker](https://docs.docker.com/engine/install/)

2. Создайте файл docker-compose.yml
Создайте новый файл с именем docker-compose.yml в удобном для вас каталоге. В этом файле вы будете описывать конфигурацию для Kafka.

3. Напишите конфигурацию
Вставьте следующий код в ваш docker-compose.yml [https://habr.com/ru/articles/810061/ - взят за основу] файл:

```yaml
version: "3.5"
services:

  x-kafka-common:
    &kafka-common
    image: bitnami/kafka:3.7
    environment:
      &kafka-common-env
      KAFKA_ENABLE_KRAFT: yes
      ALLOW_PLAINTEXT_LISTENER: yes
      KAFKA_KRAFT_CLUSTER_ID: practicum
      KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP: "CONTROLLER:PLAINTEXT,EXTERNAL:PLAINTEXT,PLAINTEXT:PLAINTEXT"
      KAFKA_CFG_PROCESS_ROLES: broker,controller
      KAFKA_CFG_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 0@kafka-0:9093,1@kafka-1:9093,2@kafka-2:9093
      KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE: false
    networks:
      - proxynet

  kafka-0:
    <<: *kafka-common
    restart: always
    ports:
      - "127.0.0.1:9094:9094"
    environment:
      <<: *kafka-common-env
      KAFKA_CFG_NODE_ID: 0
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093,EXTERNAL://:9094
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://kafka-0:9092,EXTERNAL://127.0.0.1:9094
    volumes:
      - kafka_0_data:/bitnami/kafka

  kafka-1:
    <<: *kafka-common
    restart: always
    ports:
      - "127.0.0.1:9095:9095"
    environment:
      <<: *kafka-common-env
      KAFKA_CFG_NODE_ID: 1
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093,EXTERNAL://:9095
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://kafka-1:9092,EXTERNAL://127.0.0.1:9095
    volumes:
      - kafka_1_data:/bitnami/kafka

  kafka-2:
    <<: *kafka-common
    restart: always
    ports:
      - "127.0.0.1:9096:9096"
    environment:
      <<: *kafka-common-env
      KAFKA_CFG_NODE_ID: 2
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093,EXTERNAL://:9096
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://kafka-2:9092,EXTERNAL://127.0.0.1:9096
    volumes:
      - kafka_2_data:/bitnami/kafka

  schema-registry:
    image: bitnami/schema-registry:7.6
    ports:
      - '8081:8081'
    depends_on:
      - kafka-0
      - kafka-1
      - kafka-2
    environment:
      SCHEMA_REGISTRY_LISTENERS: http://0.0.0.0:8081
      SCHEMA_REGISTRY_KAFKA_BROKERS: PLAINTEXT://kafka-0:9092,PLAINTEXT://kafka-1:9092,PLAINTEXT://kafka-2:9092
    networks:
      - proxynet
   
  ui:
    image: provectuslabs/kafka-ui:v0.7.0
    restart: always
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      KAFKA_CLUSTERS_0_BOOTSTRAP_SERVERS: kafka-0:9092
      KAFKA_CLUSTERS_0_NAME: kraft
    networks:
      - proxynet

networks:
  proxynet:
    name: custom_network

volumes:
  kafka_0_data:
  kafka_1_data:
  kafka_2_data:
```

4. В терминале перейдите в каталог, где находится ваш docker-compose.yml файл, и выполните команду:

```bash
docker compose up -d
```

5. Для проверки успешности запуска введите в терминале:

```bash
docker ps -a 
```

Ожидаемый вывод в терминале
```bash
CONTAINER ID   IMAGE                           COMMAND                  CREATED          STATUS          PORTS                                       NAMES
d39848462767   bitnami/schema-registry:7.6     "/opt/bitnami/script…"   21 seconds ago   Up 18 seconds   0.0.0.0:8081->8081/tcp, :::8081->8081/tcp   infra-schema-registry-1
e8efd3eb23d8   provectuslabs/kafka-ui:v0.7.0   "/bin/sh -c 'java --…"   21 seconds ago   Up 19 seconds   127.0.0.1:8080->8080/tcp                    infra-ui-1
435e5dc83747   bitnami/kafka:3.7               "/opt/bitnami/script…"   21 seconds ago   Up 19 seconds   9092/tcp, 127.0.0.1:9096->9096/tcp          infra-kafka-2-1
c249a9c01c72   bitnami/kafka:3.7               "/opt/bitnami/script…"   21 seconds ago   Up 20 seconds   9092/tcp, 127.0.0.1:9094->9094/tcp          infra-kafka-0-1
a65bf04f14b8   bitnami/kafka:3.7               "/opt/bitnami/script…"   21 seconds ago   Up 19 seconds   9092/tcp, 127.0.0.1:9095->9095/tcp          infra-kafka-1-1
```


Рассмотрим, фрагмент кода из docker-compose.yaml 

```yaml
x-kafka-common:
    &kafka-common
    image: bitnami/kafka:3.7
    environment:
      &kafka-common-env
      KAFKA_ENABLE_KRAFT: yes
      ALLOW_PLAINTEXT_LISTENER: yes
      KAFKA_KRAFT_CLUSTER_ID: practicum
      KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP: "CONTROLLER:PLAINTEXT,EXTERNAL:PLAINTEXT,PLAINTEXT:PLAINTEXT"
      KAFKA_CFG_PROCESS_ROLES: broker,controller
      KAFKA_CFG_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 0@kafka-0:9093,1@kafka-1:9093,2@kafka-2:9093
      KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE: false
    networks:
      - proxynet
``` 

Данный фрагмент кода представляет собой конфигурацию для контейнера Apache Kafka, использующего образ bitnami/kafka:3.7. Он определяет общие параметры, которые будут применяться ко всем экземплярам Kafka в кластере. В частности, здесь настраиваются такие важные параметры, как включение режима KRaft (Kafka Raft)  https://habr.com/ru/companies/slurm/articles/685694/ и https://raft.github.io/, который позволяет Kafka работать без Zookeeper, а также параметры безопасности и сетевого взаимодействия.

Конфигурация включает в себя установку идентификатора кластера, настройку протоколов безопасности для различных слушателей, а также определение ролей для каждого экземпляра Kafka (брокер и контроллер). Кроме того, параметр KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE отключает автоматическое создание тем, что позволяет более точно управлять структурой данных в кластере. Все эти настройки обеспечивают согласованность и надежность работы кластера Kafka, а также его интеграцию в сеть proxynet.


Для подключения интерфейса для взаимодействия с Kafka добавим следующий [код](https://habr.com/ru/articles/753398/) :

```
  ui:
    image: provectuslabs/kafka-ui:v0.7.0
    restart: always
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      KAFKA_CLUSTERS_0_BOOTSTRAP_SERVERS: kafka-0:9092
      KAFKA_CLUSTERS_0_NAME: kraft
    networks:
      - proxynet
```





## Брокеры сообщений

Среди множества доступных решений, два брокера сообщений выделяются своей популярностью и широким применением: Apache Kafka и RabbitMQ. Apache Kafka, разработанный для обработки больших объемов данных в реальном времени, идеально подходит для сценариев, требующих высокой пропускной способности и низкой задержки. В то же время RabbitMQ, с его поддержкой различных моделей обмена сообщениями и надежной доставкой, является отличным выбором для приложений, где важна гибкость и простота интеграции.

Apache Kafka

Apache Kafka — это распределенная платформа потоковой передачи данных, которая позволяет публиковать, подписываться, хранить и обрабатывать потоки записей в реальном времени. Kafka разработан для обработки больших объемов данных с высокой пропускной способностью и низкой задержкой. Он использует концепцию "топиков" для организации сообщений и поддерживает горизонтальное масштабирование, что делает его идеальным для обработки потоков данных в реальном времени, таких как журналы событий, метрики и данные IoT.

2. RabbitMQ

RabbitMQ — это брокер сообщений с открытым исходным кодом, который реализует протокол AMQP (Advanced Message Queuing Protocol). Он позволяет приложениям обмениваться сообщениями асинхронно и поддерживает различные модели обмена сообщениями, такие как очереди, публикация/подписка и маршрутизация. RabbitMQ обеспечивает надежную доставку сообщений, управление очередями и возможность обработки сообщений с помощью различных механизмов, таких как подтверждения и повторные попытки.

Отличие от RabbitMQ и других брокеров сообщений (push, pull, персистентность)


Основная терминология 

Базовые компоненты
