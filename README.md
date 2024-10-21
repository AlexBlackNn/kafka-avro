# Введение в Apache Kafka. 

## Зачем нужны брокеры сообщений

Брокеры сообщений, играют важную роль в архитектуре современных распределенных систем. 
Они обеспечивают асинхронную передачу данных между различными компонентами системы, что позволяет 
улучшить масштабируемость, надежность и производительность, а также позволяют реализовать [kappa-архитектуру](https://bigdataschool.ru/blog/kappa-architecture.html)
и делают связность между сервисами более слабой.

Предположим, у нас есть интернет-магазин с несколькими компонентами:

    Веб-сервер: Обрабатывает запросы пользователей.
    Система управления заказами: Обрабатывает заказы и управляет их состоянием.
    Система управления запасами: Отслеживает наличие товаров на складе.
    Система уведомлений: Отправляет уведомления пользователям (например, по электронной почте или SMS).

Рассмотрим сценарий обработки заказа
    Пользователь делает заказ. Когда пользователь завершает покупку, веб-сервер получает запрос на создание заказа. 
    Вместо того чтобы сразу обрабатывать заказ, веб-сервер отправляет сообщение о новом заказе в брокер сообщений.

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

    Система управления заказами подписана на сообщения о новых заказах. Она получает сообщение, обрабатывает его (например, проверяет наличие товаров, создает запись в базе данных) и обновляет статус заказа.После успешной обработки заказа система управления заказами отправляет сообщение в брокер о том, что товары были заказаны. Система управления запасами подписана на эти сообщения и обновляет количество доступных товаров на складе. После того как заказ был успешно обработан, система управления заказами отправляет сообщение о завершении заказа в брокер. Система уведомлений подписана на эти сообщения и отправляет пользователю уведомление о том, что заказ был успешно оформлен.

## Базовые понятия 
Давайте попробуем сделать эмитацию такого брокера сообщений самостоятельно с использованием обычного файла и поговорим о терминалогии. 

Для начала напишем клиентский код, который отвечает за запись сообщениий в наш импровизируемый брокер. Такой клиентский код называется *Продюсером* (Producer) он публикуют события в брокер. 


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

Теперь перейдем к  клиентскому приложению, которое отвечает за чтение сообщениий из нашего импровизированного брокера. Такой клиентский код называется *Консьюмером* (Consumer) он читает события из брокера.

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

Давайте теперь объеденим вместе код читателя и писателя. Хотелось бы также, чтобы читатель и писатель могли работать параллельно. А если читатель вдруг по какой-то причине будет выключен, было бы здорово, чтобы он начал читать с того места (*Оффсета*), на котором он остановился.

```
код.
```

Конечно, брокеры сообщений разработаны с гораздо большей сложностью, чем наш импровизированный брокер. Тем не менее, использование этого кода позволяет вам ознакомиться с основными концепциями и принципами работы брокеров сообщений. 

## Брокеры сообщений

Среди множества доступных решений, два брокера сообщений выделяются своей популярностью и широким применением: Apache Kafka и RabbitMQ. Apache Kafka, разработанный для обработки больших объемов данных в реальном времени, идеально подходит для сценариев, требующих высокой пропускной способности и низкой задержки. В то же время RabbitMQ, с его поддержкой различных моделей обмена сообщениями и надежной доставкой, является отличным выбором для приложений, где важна гибкость и простота интеграции.

Apache Kafka

Apache Kafka — это распределенная платформа потоковой передачи данных, которая позволяет публиковать, подписываться, хранить и обрабатывать потоки записей в реальном времени. Kafka разработан для обработки больших объемов данных с высокой пропускной способностью и низкой задержкой. Он использует концепцию "топиков" для организации сообщений и поддерживает горизонтальное масштабирование, что делает его идеальным для обработки потоков данных в реальном времени, таких как журналы событий, метрики и данные IoT.

2. RabbitMQ

RabbitMQ — это брокер сообщений с открытым исходным кодом, который реализует протокол AMQP (Advanced Message Queuing Protocol). Он позволяет приложениям обмениваться сообщениями асинхронно и поддерживает различные модели обмена сообщениями, такие как очереди, публикация/подписка и маршрутизация. RabbitMQ обеспечивает надежную доставку сообщений, управление очередями и возможность обработки сообщений с помощью различных механизмов, таких как подтверждения и повторные попытки.

Отличие от RabbitMQ и других брокеров сообщений (push, pull, персистентность)


Основная терминология 

Базовые компоненты


## Порядок установки и настройки кластера локально с использованием Docker

Одним из самых простых способов запустить Kafka является использование Docker Compose.

1. Установите [Docker](https://docs.docker.com/engine/install/)

2. Создайте файл docker-compose.yml
Создайте новый файл с именем docker-compose.yml в удобном для вас каталоге. В этом файле вы будете описывать конфигурацию для Kafka.

3. Напишите конфигурацию
Вставьте следующий код в ваш docker-compose.yml файл:

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
docker-compose up -d
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

Данный фрагмент кода представляет собой конфигурацию для контейнера Apache Kafka, использующего образ bitnami/kafka:3.7. Он определяет общие параметры, которые будут применяться ко всем экземплярам Kafka в кластере. В частности, здесь настраиваются такие важные параметры, как включение режима KRaft (Kafka Raft)  https://habr.com/ru/companies/slurm/articles/685694/ и https://raft.github.io/, который позволяет Kafka работать без Zookeeper, а также параметры безопасности и сетевого взаимодействия.

Конфигурация включает в себя установку идентификатора кластера, настройку протоколов безопасности для различных слушателей, а также определение ролей для каждого экземпляра Kafka (брокер и контроллер). Кроме того, параметр KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE отключает автоматическое создание тем, что позволяет более точно управлять структурой данных в кластере. Все эти настройки обеспечивают согласованность и надежность работы кластера Kafka, а также его интеграцию в сеть proxynet.
