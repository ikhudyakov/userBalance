package api

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

func (u User) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Id,
			validation.Required.Error("id пользователя не может быть не указан либо <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")))
}

func (r Replenishment) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(
			&r.UserID,
			validation.Required.Error("id пользователя не может быть не указан либо <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")),
		validation.Field(
			&r.Amount,
			validation.Required.Error("сумма пополнения должна быть больше 0"),
			validation.Min(1).Error("сумма пополнения должна быть больше 0")))
}

func (m Money) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.FromUserID,
			validation.Required.Error("id пользователя не может быть не указан либо <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")),
		validation.Field(&m.ToUserID,
			validation.Required.Error("id пользователя не может быть не указан либо <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0"),
			validation.NotIn(m.FromUserID).Error("невозможно перевести самому себе")),
		validation.Field(&m.Amount,
			validation.Required.Error("сумма перевода должна быть больше 0"),
			validation.Min(1).Error("сумма перевода должна быть больше 0")))
}

func (t Transaction) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.UserID,
			validation.Required.Error("id пользователя не может быть не указан либо <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")),
		validation.Field(&t.Amount,
			validation.Required.Error("стоимость услуги должна быть больше 0"),
			validation.Min(1).Error("стоимость услуги должна быть больше 0")),
		validation.Field(&t.OrderID,
			validation.Required.Error("номер заказа не может быть <= 0"),
			validation.Min(1).Error("номер заказа не может быть <= 0")),
		validation.Field(&t.ServiceID,
			validation.Required.Error("id услуги не может быть <= 0"),
			validation.Min(1).Error("id услуги не может быть <= 0")))
}

func (r RequestHistory) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(
			&r.UserID,
			validation.Required.Error("id пользователя не может быть <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")))
}

func (r RequestReport) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(
			&r.Month,
			validation.Required.Error("месяца не может быть не указан либо <= 0"),
			validation.Min(1).Error("месяца не может быть <= 0"),
			validation.Max(12).Error("месяца не может быть > 12")),
		validation.Field(
			&r.Year,
			validation.Required.Error("год не может быть не указан либо <= 0"),
			validation.Min(1970).Error("неверно указан год")))
}
