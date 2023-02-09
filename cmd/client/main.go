package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	c "userbalance/internal/config"
	"userbalance/pkg/api"
)

const help string = `Для выбора RPC метода введите его название:
GetBalance - Получение баланса пользователя
ReplenishmentBalance - Пополнение баланса пользователя
Transfer - Перевод средств от пользователя пользователю
GetHistory - Получение истории пользователя
Reservation - Резервирование средств
Confirmation - Списание зарезервированных средств
CancelReservation - Разрезервирование средств
help - Помощь
quit - Выход
`

func main() {
	var err error
	var conf *c.Config
	var path string
	var defaultPath string = "./configs/config.yaml"
	var in string

	path = defaultPath

	flag.StringVar(&path, "config", "./configs/config.yaml", "example -config ./configs/config.yaml")

	flag.Parse()

	conf, err = c.GetConfig(path)
	if err != nil {
		log.Printf("%s, use default config '%s'", err, defaultPath)
		conf, err = c.GetConfig(defaultPath)
		if err != nil {
			log.Println(err)
			return
		}
	}

	c := new(Client)
	err = c.Run(conf)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Print(help)

	for {
		fmt.Scan(&in)
		switch in {
		case "quit":
			c.Shutdown()
			return

		case "help":
			fmt.Print(help)

		case "GetBalance":
			var id int32
			fmt.Printf("Введите id пользователя: ")
			fmt.Scan(&id)
			response, err := c.client.GetBalance(context.Background(), &api.User{Id: id})
			if err != nil {
				log.Println(err)
			} else {
				log.Println(response)
			}

		case "ReplenishmentBalance":
			var id, amount int32
			var date string
			fmt.Printf("Введите id пользователя: ")
			fmt.Scan(&id)
			fmt.Printf("Введите сумму: ")
			fmt.Scan(&amount)
			fmt.Printf("Введите дату (необязательно): ")
			fmt.Scanln(&date)
			response, err := c.client.ReplenishmentBalance(context.Background(), &api.Replenishment{UserID: id, Amount: amount, Date: date})
			if err != nil {
				log.Println(err)
			} else {
				log.Println(response)
			}

		case "Transfer":
			var fromid, toid, amount int32
			var date string
			fmt.Printf("Введите id отправителья: ")
			fmt.Scan(&fromid)
			fmt.Printf("Введите id получателя: ")
			fmt.Scan(&toid)
			fmt.Printf("Введите сумму: ")
			fmt.Scan(&amount)
			fmt.Printf("Введите дату (необязательно): ")
			fmt.Scanln(&date)
			response, err := c.client.Transfer(context.Background(), &api.Money{FromUserID: fromid, ToUserID: toid, Amount: amount, Date: date})
			if err != nil {
				log.Println(err)
			} else {
				log.Println(response)
			}

		case "GetHistory":
			var id int32
			var sort, dir string
			fmt.Printf("Введите id пользователя: ")
			fmt.Scan(&id)
			fmt.Printf("Укажите поле для сортировки (amount, date) (необязательно):")
			fmt.Scanln(&sort)
			fmt.Printf("Укажите направление сортировки (desc, asc) (необязательно):")
			fmt.Scanln(&dir)
			response, err := c.client.GetHistory(context.Background(), &api.RequestHistory{UserID: id, SortField: sort, Direction: dir})
			if err != nil {
				log.Println(err)
			} else {
				for _, v := range response.Entity {
					log.Printf("date: %s, amount: %d, description: %s\n", v.Date.AsTime().Format("02/01/2006"), v.Amount, v.Description)
				}
			}

		case "Reservation":
			var id, amount, serviceID, orderID int32
			var date string
			fmt.Printf("Введите id пользователя: ")
			fmt.Scan(&id)
			fmt.Printf("Введите сумму: ")
			fmt.Scan(&amount)
			fmt.Printf("Введите дату (необязательно): ")
			fmt.Scanln(&date)
			fmt.Printf("Введите id услуги: ")
			fmt.Scan(&serviceID)
			fmt.Printf("Введите номер заказа: ")
			fmt.Scan(&orderID)
			response, err := c.client.Reservation(context.Background(), &api.Transaction{UserID: id, Amount: amount, Date: date, ServiceID: 1, OrderID: 1})
			if err != nil {
				log.Println(err)
			} else {
				log.Println(response)
			}

		case "Confirmation":
			var id, amount, serviceID, orderID int32
			var date string
			fmt.Printf("Введите id пользователя: ")
			fmt.Scan(&id)
			fmt.Printf("Введите сумму: ")
			fmt.Scan(&amount)
			fmt.Printf("Введите дату (необязательно): ")
			fmt.Scanln(&date)
			fmt.Printf("Введите id услуги: ")
			fmt.Scan(&serviceID)
			fmt.Printf("Введите номер заказа: ")
			fmt.Scan(&orderID)
			response, err := c.client.Confirmation(context.Background(), &api.Transaction{UserID: id, Amount: amount, Date: date, ServiceID: 1, OrderID: 1})
			if err != nil {
				log.Println(err)
			} else {
				log.Println(response)
			}

		case "CancelReservation":
			var id, amount, serviceID, orderID int32
			var date string
			fmt.Printf("Введите id пользователя: ")
			fmt.Scan(&id)
			fmt.Printf("Введите сумму: ")
			fmt.Scan(&amount)
			fmt.Printf("Введите дату (необязательно): ")
			fmt.Scanln(&date)
			fmt.Printf("Введите id услуги: ")
			fmt.Scan(&serviceID)
			fmt.Printf("Введите номер заказа: ")
			fmt.Scan(&orderID)
			response, err := c.client.CancelReservation(context.Background(), &api.Transaction{UserID: id, Amount: amount, Date: date, ServiceID: 1, OrderID: 1})
			if err != nil {
				log.Println(err)
			} else {
				log.Println(response)
			}
		}
	}
}
