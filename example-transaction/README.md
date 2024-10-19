
Установите 
gogen-avro
Генерирует типобезопасный код Go на основе ваших схем Avro, включая сериализаторы и десериализаторы, которые поддерживают правила эволюции схем Avro. Также поддерживает десериализацию общих данных Avro (в бета-версии).

gogen-avro --package=kafkapracticum --containers=false --sources-comment=false --short-unions=false /home/user/Dev/kafka-avro/producer/cmd /home/user/Dev/kafka-avro/producer/cmd/user.avsc


go run ./example-transaction/cmd/main.go -c ./example-transaction/config/local.yaml -t producer
go run ./example-transaction/cmd/main.go -c ./example-transaction/config/local.yaml -t consumer

