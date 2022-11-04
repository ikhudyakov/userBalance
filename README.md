# Микросервис для работы с балансом пользователей

## Установка
- Склонировать репозиторий _```git clone https://github.com/ikhudyakov/userBalance.git```_
- Перейти в папку _userBalance_
- Выполнить команду ```docker-compose -f deploy/docker-compose.yml up -d```

## Использование 
### 1. Пополнение баланса пользователя
Для пополнения баланса пользователя в теле POST запроса по адресу ```localhost:8081/topup``` отправляем JSON следующего вида:
```json
{
    "userid":15,
    "amount":500,
    "date":"2022-08-01"
}
```
*где `userid` - ID пользователя, `amount` - сумма пополнения, `date` - дата пополнения в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
При успешном пополнении в ответ получаем JSON:
```json
{
    "message": "баланс пополнен"
}
```
***
### 2. Получение баланса пользователя
Для получения баланса пользователя в теле POST запроса по адресу ```localhost:8081/``` отправляем JSON следующего вида:
```json
{
    "userid":15
}
```
*где `userid` - ID пользователя*</br>
При успешном выполнении запроса в ответ получаем JSON:
```json
{
    "userid": 15,
    "balance": 500
}
```
*где `userid` - ID пользователя, `balance` - текущий баланс пользователя*</br>
***
### 3. Перевод средств от пользователя пользователю 
Для перевода средств от пользователя пользователю в теле POST запроса по адресу ```localhost:8081/transfer``` отправляем JSON следующего вида:
```json
{
    "fromuserid":1,
    "touserid":2,
    "amount":100,
    "date":"2022-10-10"
}
```
*где `fromuserid` - ID пользователя-отправителя, `touserid` - ID пользователя-получателя, `amount` - сумма, `date` - дата в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
При успешном выполнении запроса в ответ получаем JSON:
```json
{
    "message": "перевод стредств выполнен"
}
```
***
### 4. Резервирование средств 
Для резервирования средств на балансе в теле POST запроса по адресу ```localhost:8081/reserv``` отправляем JSON следующего вида:
```json
{
    "userid":15,
    "amount":100,
    "serviceid":1,
    "orderid":10025,
    "date":"2022-10-10"
}
```
*где `userid` - ID пользователя, `amount` - сумма, `serviceid` - ID услуги, `orderid` - ID заказа, `date` - дата в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
При успешном выполнении запроса в ответ получаем JSON:
```json
{
    "message": "резервирование средств прошло успешно"
}
```
***
### 5. Списание зарезервированных средств
Для списания зарезервированных средств в теле POST запроса по адресу ```localhost:8081/confirm``` отправляем JSON следующего вида:
```json
{
    "userid":15,
    "amount":100,
    "serviceid":1,
    "orderid":10025,
    "date":"2022-10-10"
}
```
*где `userid` - ID пользователя, `amount` - сумма, `serviceid` - ID услуги, `orderid` - ID заказа, `date` - дата в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
Для списания зарезервированных средств значения полей должны соответствовать значениям полей, введенных при резервировании</br> 
При успешном выполнении запроса в ответ получаем JSON:
```json
{
    "message": "средства из резерва были списаны успешно"
}
```
***
### 6. Разрезервирование средств
Если услугу применить не удалось, нужно провести разрезервирование средств, для этого в теле POST запроса по адресу ```localhost:8081/cancel``` отправляем JSON следующего вида:
```json
{
    "userid":15,
    "amount":100,
    "serviceid":1,
    "orderid":10025,
    "date":"2022-10-10"
}
```
*где `userid` - ID пользователя, `amount` - сумма, `serviceid` - ID услуги, `orderid` - ID заказа, `date` - дата в формате `yyy-mm-dd` (при отсутствии поля `date` или неверном формате устанавливается текущая дата)*</br>
Для разрезервирования средств значения полей должны соответствовать значениям полей, введенных при резервировании</br> 
При успешном выполнении запроса в ответ получаем JSON:
```json
{
    "message": "разрезервирование средств прошло успешно"
}
```
***
### 7. Получение отчета по услугам
Для получения отчета в теле POST запроса по адресу ```localhost:8081/report``` отправляем JSON следующего вида:
```json
{
    "fromdate":"2022-10-20",
    "todate":"2022-10-31"
}
```
*где `fromdate` - начала периода, `todate` - конец периода. Дата указывается в формате `yyy-mm-dd (при отсутствии полей или неверном формате отчет формируется за весь текущий месяц)*</br>
При успешном выполнении запроса в ответ получаем JSON:
```json
{
    "message": "localhost:8081/file/1234567890.csv"
}
```
*в котором будет ссылка на скачивание сформированного отчета в формате .csv*</br>
***
### 8. Получение истории пользователя
Для получения истории пользователя в теле POST запроса по адресу ```localhost:8081/history``` отправляем JSON следующего вида:
```json
{
    "userid":15,
    "sortfield":"amount",
    "direction":"desc"
}
```
*где `userid` - ID пользователя, `sortfield` - по какому полю сортировать: сумма, дата ("amount", "date"), `direction` - направление сортировки: по возрастанию, по убыванию ("asc", "desc")</br>(при отсутствии или неверном формате полей сортировка происходит по возрастанию суммы)*</br>
При успешном выполнении запроса в ответ получаем JSON со всей историей передвижения средст пользователя:
```json
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
***

## Swagger-документация
 По адресу ``http://localhost:8081/swagger/index.html`` доступна swagger-документация