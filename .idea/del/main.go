// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
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

	// Проверяем, что количество параметров при запуске нашей программы ровно 3
	if len(os.Args) != 3 {
		log.Fatalf("Пример использования: %s <bootstrap-servers> <topic>\n", os.Args[0])
	}

	// Парсим параметы и получаем адрес брокера и имя топика
	bootstrapServers := os.Args[1]
	topic := os.Args[2]

	// Создаем продьюсера
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})
	if err != nil {
		log.Fatalf("Невозможно создать продьюсера: %s\n", err)
	}

	log.Printf("Продьюсер создан %v\n", p)

	// Канал доставки событий (информации об отправленном сообщении)
	deliveryChan := make(chan kafka.Event)

	// Создаем заказ
	value := &Order{
		OrderID: "0001",
		UserID:  "00001",
		Items: []Item{
			{ProductID: "535", Quantity: 1, Price: 300},
			{ProductID: "125", Quantity: 2, Price: 100},
		},
		TotalPrice: 500.00,
	}

	// Сериализуем заказ
	payload, err := json.Marshal(value)
	if err != nil {
		log.Fatalf("Невозможно сериализовать заказ: %s\n", err)
	}

	// Отправляем сообщение в брокер
	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
		Headers:        []kafka.Header{{Key: "myTestHeader", Value: []byte("header values are binary")}},
	}, deliveryChan)
	if err != nil {
		log.Fatalf("Ошибка при отправке сообщения: %v\n", err)
	}

	// Ждем информацию об отправленном сообщении. Для простоты сделана синхронная запись.
	// В реальных проектах ее использование не рекомендуется, так как она снижает пропускную
	// способность (https://docs.confluent.io/kafka-clients/go/current/overview.html#synchronous-writes)
	e := <-deliveryChan

	// Приводим Events к типу *kafka.Message, подробнее про Events, можно почитать тут (https://docs.confluent.io/platform/current/clients/confluent-kafka-go/index.html#hdr-Events)
	m := e.(*kafka.Message)

	// Если возникла ошибка доставки сообщения
	if m.TopicPartition.Error != nil {
		fmt.Printf("Ошибка доставки сообщения: %v\n", m.TopicPartition.Error)
	} else {
		fmt.Printf("Сообщение отправлено в топик %s [%d] офсет %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}
	// Закрываем отправителя
	p.Close()
	// Не забываем закрыть канал доставки событий
	close(deliveryChan)
}
