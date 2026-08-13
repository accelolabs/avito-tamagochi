package service

import (
	"context"
	"fmt"
	"html"

	authmodel "github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	"github.com/accelolabs/avito-tamagochi/backend/internal/notifications/mailer"
	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
	"github.com/google/uuid"
)

const subject = "Питомец ждёт вас"

type UserFinder interface {
	FindUser(context.Context, uuid.UUID) (*authmodel.User, error)
}

type PetFinder interface {
	GetPet(context.Context, uuid.UUID) (*petmodel.Stats, error)
}

type Service interface {
	DispatchEnergy(context.Context, uuid.UUID) (*notificationmodel.DispatchResult, error)
}

type service struct {
	users  UserFinder
	pets   PetFinder
	mailer mailer.Mailer
}

type template struct {
	threshold  int
	firstLine  string
	secondLine string
}

var templates = map[int]template{
	50: {threshold: 50, firstLine: "Я по тебе соскучился... 🥺", secondLine: "Навестишь меня?"},
	25: {threshold: 25, firstLine: "Я что-то совсем без сил... 😥", secondLine: "Энергии почти не осталось. Заглянешь зарядить меня?"},
	5:  {threshold: 5, firstLine: "Кажется, я сейчас совсем сяду… 🪫", secondLine: "Когда энергия закончится, я потеряю накопленный прогресс. Поможешь мне сохранить его?"},
	0:  {threshold: 0, firstLine: "Я полностью разрядился... 😭", secondLine: "Все мои накопленные знания стираются. Надеюсь, мы сможем еще увидеться."},
}

func New(users UserFinder, pets PetFinder, mailer mailer.Mailer) Service {
	return &service{users: users, pets: pets, mailer: mailer}
}

func (s *service) DispatchEnergy(ctx context.Context, userID uuid.UUID) (*notificationmodel.DispatchResult, error) {
	user, err := s.users.FindUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find notification recipient: %w", err)
	}
	pet, err := s.pets.GetPet(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get pet for energy notification: %w", err)
	}

	selected, ok := templateForEnergy(pet.Energy)
	if !ok {
		return &notificationmodel.DispatchResult{Status: notificationmodel.StatusSkipped, Energy: pet.Energy}, nil
	}

	message := notificationmodel.Message{
		Recipient: user.Email,
		Subject:   subject,
		TextBody:  selected.firstLine + "\n" + selected.secondLine,
		HTMLBody:  "<p><strong>" + html.EscapeString(selected.firstLine) + "</strong><br>" + html.EscapeString(selected.secondLine) + "</p>",
	}
	if err := s.mailer.Send(ctx, message); err != nil {
		return nil, fmt.Errorf("send energy notification: %w", err)
	}
	threshold := selected.threshold
	return &notificationmodel.DispatchResult{Status: notificationmodel.StatusSent, Energy: pet.Energy, Threshold: &threshold}, nil
}

func templateForEnergy(energy int) (template, bool) {
	switch {
	case energy == 0:
		return templates[0], true
	case energy <= 5:
		return templates[5], true
	case energy <= 25:
		return templates[25], true
	case energy <= 50:
		return templates[50], true
	default:
		return template{}, false
	}
}
