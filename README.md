# Тестовое задание - TCP сервер с цитатами

[![Actions Status](https://github.com/mainpart/tcp-pow/workflows/main/badge.svg)](https://github.com/mainpart/tcp-pow/actions)

## 1. Запуск в докере
`docker-compose up -d`

## 2. Запуск локально

`go run server.go`

`go run client.go`

## 3. Описание задания (из telegram)
Design and implement “Word of Wisdom” tcp server. 
TCP server should be protected from DDOS attacks with the [Proof of Work](https://en.wikipedia.org/wiki/Proof_of_work), 
the challenge-response protocol should be used.  
The choice of the PoW algorithm should be explained.  
After Prof Of Work verification, server should send one of the quotes from “word of wisdom” book or any other collection of the quotes.  
Docker file should be provided both for the server and for the client that solves the PoW challenge.


### 4 Описание результата

Результат представляет собой 2 приложения - один сервер, другой клиент. Протокол общения выбран TCP, первой строкой идет тип сообщения, далее payload. В качестве алгоритма выбран [HashCash](https://en.wikipedia.org/wiki/Hashcash) как наиболее поддающийся конфигурации.
Цитаты "подтягиваются" с online api 


+ [Merkle tree](https://en.wikipedia.org/wiki/Merkle_tree)
+ [Hashcash](https://en.wikipedia.org/wiki/Hashcash)
+ [Guided tour puzzle](https://en.wikipedia.org/wiki/Guided_tour_puzzle_protocol)