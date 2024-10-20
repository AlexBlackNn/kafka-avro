
Установите 
gogen-avro
Генерирует типобезопасный код Go на основе ваших схем Avro, включая сериализаторы и десериализаторы, которые поддерживают правила эволюции схем Avro. Также поддерживает десериализацию общих данных Avro (в бета-версии).

gogen-avro --package=kafkapracticum --containers=false --sources-comment=false --short-unions=false /home/user/Dev/kafka-avro/producer/cmd /home/user/Dev/kafka-avro/producer/cmd/user.avsc



```bash
go run ./cmd/create-topic/main.go localhost:9094 users 3 2
```


```bash
go run ./cmd/main.go -c ./config/local.yaml -t producer
```

```bash
go run ./cmd/main.go -c ./config/local.yaml -t consumer
```
