# Микросервис для работы с балансом пользователей

## Установка
- Склонировать репозиторий _```git clone https://github.com/ikhudyakov/userBalance.git -b grpc```_
- Перейти в папку _userBalance_
- Выполнить команду ```docker-compose -f deploy/docker-compose.yml up -d```
***

## Параметры для сервера
- Для указания пути до файла конфигурации, запускаем программу с параметром `-config "путь_до_файла"` (по умолчанию используется `./configs/config.yaml`)
- Для выполнения миграции используется флаг `-migrationup`
- Для отката миграции используется флаг `-migrationdown`

Пример: 
```
./server -config ./configs/newconfig.yaml -migrationdown -migrationup
```

## Параметры для клиента
- Для указания пути до файла конфигурации, запускаем программу с параметром `-config "путь_до_файла"` (по умолчанию используется `./configs/config.yaml`)

Пример: 
```
./client -config ./configs/newconfig.yaml
```
***

## RPC Методы 
### 1. Пополнение баланса пользователя
ReplenishmentBalance(ctx context.Context, replenishment *api.Replenishment) (*api.Response, error)

*где `replenishment.userId` - ID пользователя, `replenishment.amount` - сумма пополнения, `replenishment.date` - дата пополнения в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
При успешном пополнении в ответ получаем структуру Response:
```
{
    Message: "баланс пополнен",
}
```

***
### 2. Получение баланса пользователя
GetBalance(ctx context.Context, user *api.User) (*api.User, error)

*где `user.id` - ID пользователя*</br>
При успешном выполнении запроса в ответ получаем структуру User:
```
{
    Id: 1,
    Balance: 100,
}
```
*где `id` - ID пользователя, `balance` - текущий баланс пользователя*</br>
***
### 3. Перевод средств от пользователя пользователю 
Transfer(ctx context.Context, money *api.Money) (*api.Response, error) 

*где `money.fromuserid` - ID пользователя-отправителя, `money.touserid` - ID пользователя-получателя, `money.amount` - сумма, `money.date` - дата в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
При успешном выполнении запроса в ответ получаем структуру Response:
```
{
    Message: "перевод стредств выполнен",
}
```
***
### 4. Резервирование средств 
Reservation(ctx context.Context, transaction *api.Transaction) (*api.Response, error)

*где `transaction.userid` - ID пользователя, `transaction.amount` - сумма, `transaction.serviceid` - ID услуги, `transaction.orderid` - ID заказа, `transaction.date` - дата в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
При успешном выполнении запроса в ответ получаем структуру Response:
```
{
    Message: "резервирование средств прошло успешно",
}
```
***
### 5. Списание зарезервированных средств
Confirmation(ctx context.Context, transaction *api.Transaction) (*api.Response, error)

*где `transaction.userid` - ID пользователя, `transaction.amount` - сумма, `transaction.serviceid` - ID услуги, `transaction.orderid` - ID заказа, `transaction.date` - дата в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
Для списания зарезервированных средств значения полей должны соответствовать значениям полей, введенных при резервировании</br> 
При успешном выполнении запроса в ответ получаем структуру Response:
```
{
    Message: "средства из резерва были списаны успешно",
}
```
***
### 6. Разрезервирование средств
CancelReservation(ctx context.Context, transaction *api.Transaction) (*api.Response, error)

*где `transaction.userid` - ID пользователя, `transaction.amount` - сумма, `transaction.serviceid` - ID услуги, `transaction.orderid` - ID заказа, `transaction.date` - дата в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
Для разрезервирования средств значения полей должны соответствовать значениям полей, введенных при резервировании</br> 
При успешном выполнении запроса в ответ получаем структуру Response:
```
{
    Message: "разрезервирование средств прошло успешно",
}
```
***
### 7. Получение истории пользователя
GetHistory(ctx context.Context, requestHistory *api.RequestHistory) (*api.Histories, error)

*где `requestHistory.userid` - ID пользователя, `requestHistory.sortfield` - по какому полю сортировать: сумма, дата ("amount", "date"), `requestHistory.direction` - направление сортировки: по возрастанию, по убыванию ("asc", "desc")</br>(при отсутствии или неверном формате полей сортировка происходит по возрастанию суммы)*</br>
При успешном выполнении запроса в ответ получаем структуру Histories со всей историей передвижения средст пользователя:
```
[
    {
        "Date": "01/08/2022",
        "Amount": 500,
        "Description": "Пополнение баланса"
    },
    {
        "Date": "10/10/2022",
        "Amount": 100,
        "Description": "Заказ №10025, услуга \"Услуга 1\""
    },
    {
        "Date": "10/10/2022",
        "Amount": 100,
        "Description": "Отмена заказа №10025, услуга \"Услуга 1\""
    }
]
```