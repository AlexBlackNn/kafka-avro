## Проект Kafka Avro на Go

Этот проект демонстрирует работу с Kafka и Avro на языке Go.

### Описание

Код проекта можно скомпилировать в два независимо запускаемых приложений:

* Производитель (Producer): Записывает данные в формате Avro в топик Kafka.
* Потребитель (Consumer): Считывает данные из топика Kafka и выводит их в консоль. При этом потребитель самостоятельно фиксирует смещения обработанных данных.

### Установка

1. Установите Go с официального сайта: https://go.dev/doc/install

2. Установите требуемые зависимости:
```bash
go mod tidy
```

3. Настройте конфигурационный файл (можно оставить без изменений):
    * `config/local.yaml`: файл конфигурации для подключения к Kafka и установки параметров.

```yaml
env: "local" 
kafka:
  kafkaUrl: "localhost:9094,localhost:9095,localhost:9096" # Адреса для подключения к брокерам кластера
  schemaRegistryURL: "http://localhost:8081" # Адреса для подключения к брокерам   schema Registry
  topic: "users" # Название топика 
```

4. Установите [librdkafka](https://github.com/confluentinc/librdkafka#installation)

### Запуск
1. Запустите инфраструктуру
```bash
cd ../infra && docker compose up -d
```

2. Создайте топик:
```bash
   go run ./cmd/create-topic/main.go localhost:9094 users 3 2
```
Параметры `localhost:9094` - адрес Kafka-брокера, `users` - имя топика, `3` - количество партиций, `2` - на количество реплик.

3. Запустите Producer:
```bash
    go run ./cmd/main.go -c ./config/local.yaml -t producer
```
    а. Программа у вас запросит требуемое действие (Command):   
    введите `send`, чтобы отправить сообщение в брокер
    введите `exit`, чтобы выйти

    б. Программа у вас запросит требуемые данные для отправки - имя (Enter name:), 
    любимое число (Enter favorite number:), любимый цвет (Enter favorite color:). Пожалуйста,
    заполняйте правильно вводимые значения, так как основная задача показать возможность передачи
    данных в формате Avro, используя Kafka, валидация данных не сделана. 
    
    Пример:
    ```
    Command:send
    Enter name: alex
    Enter favorite number: 55
    Enter favorite color:black
    ```

   Примерный вывод программы и действия пользователя 
   ```
   go run ./cmd/main.go -c ./config/local.yaml -t producer
   2024/10/24 14:30:52 application starts with cfg -> type: producer, env: local, kafka url localhost:9094,localhost:9095,localhost:9096, schema registry url http://localhost:8081 
   time=2024-10-24T14:30:52.822+03:00 level=INFO source=/home/alex/Dev/2/kafka-avro/avro-example/app/producer/app.go:41 msg="producer starts"
   Command: send
   Enter name: alex
   Enter favorite number: 55
   Enter favorite color: black
   time=2024-10-24T14:31:00.908+03:00 level=INFO source=/home/alex/Dev/2/kafka-avro/avro-example/internal/broker/producer/producer.go:98 msg="sending message" msg="{Name:alex Favorite_number:55 Favorite_color:black}"
   Command: exit
   2024/10/24 14:31:03 terminate
   exit status 1
   ```

4. Запустите Consumer:
```bash
   go run ./cmd/main.go -c ./config/local.yaml -t consumer
```

   Примерный вывод:
   ```
   2024/10/24 14:32:53 application starts with cfg -> type: consumer, env: local, kafka url localhost:9094,localhost:9095,localhost:9096, schema registry url http://localhost:8081 
   time=2024-10-24T14:32:53.672+03:00 level=INFO source=/home/alex/Dev/2/kafka-avro/avro-example/app/consumer/app.go:39 msg="producer starts"
   time=2024-10-24T14:33:00.003+03:00 level=INFO source=/home/alex/Dev/2/kafka-avro/avro-example/internal/broker/consumer/consumer.go:99 msg="Message received" topic=users[0]@32 message="{Name:Alex Favorite_number:55 Favorite_color:black}"
   time=2024-10-24T14:33:01.004+03:00 level=INFO source=/home/alex/Dev/2/kafka-avro/avro-example/internal/broker/consumer/consumer.go:99 msg="Message received" topic=users[0]@33 message="{Name:alex Favorite_number:12 Favorite_color:blue}"
   time=2024-10-24T14:33:02.005+03:00 level=INFO source=/home/alex/Dev/2/kafka-avro/avro-example/internal/broker/consumer/consumer.go:99 msg="Message received" topic=users[0]@34 message="{Name:alex Favorite_number:55 Favorite_color:black}"
   time=2024-10-24T14:33:04.106+03:00 level=WARN source=/home/alex/Dev/2/kafka-avro/avro-example/internal/broker/consumer/consumer.go:119 msg=Event: msg="OffsetsCommitted (<nil>, [users[0]@35 users[1]@unset users[2]@unset])"
   ```