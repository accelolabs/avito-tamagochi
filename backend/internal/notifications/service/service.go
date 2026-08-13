package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/rules"
	"github.com/accelolabs/avito-tamagochi/backend/internal/notifications/mailer"
	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
)

const subject = "Питомец ждёт вас"

type Repository interface {
	TryRunLock(context.Context) (release func(), acquired bool, err error)
	ListParticipantIDs(context.Context) ([]notificationmodel.Participant, error)
	ProcessParticipant(context.Context, notificationmodel.Participant, func(notificationmodel.Participant, map[int]bool) (*int, error)) (bool, error)
}

type Service interface {
	DispatchAll(context.Context) (notificationmodel.BatchResult, error)
}

type service struct {
	repo   Repository
	mailer mailer.Mailer
	now    func() time.Time
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
	0:  {threshold: 0, firstLine: "Я полностью разрядился... 😭", secondLine: "Весь мой накопленный опыт стирается. Надеюсь, мы сможем еще увидеться."},
}

func New(repo Repository, mailer mailer.Mailer) Service {
	return &service{repo: repo, mailer: mailer, now: time.Now}
}

func (s *service) DispatchAll(ctx context.Context) (notificationmodel.BatchResult, error) {
	release, acquired, err := s.repo.TryRunLock(ctx)
	if err != nil {
		return notificationmodel.BatchResult{}, fmt.Errorf("acquire notification run lock: %w", err)
	}
	if !acquired {
		return notificationmodel.BatchResult{}, nil
	}
	defer release()

	participants, err := s.repo.ListParticipantIDs(ctx)
	if err != nil {
		return notificationmodel.BatchResult{}, fmt.Errorf("list notification participants: %w", err)
	}
	result := notificationmodel.BatchResult{Participants: len(participants)}
	var failures []error
	for _, participant := range participants {
		sent, dispatchErr := s.repo.ProcessParticipant(ctx, participant, func(current notificationmodel.Participant, delivered map[int]bool) (*int, error) {
			return s.dispatchParticipant(ctx, current, delivered)
		})
		if dispatchErr != nil {
			result.Failed++
			failures = append(failures, fmt.Errorf("participant %s: %w", participant.UserID, dispatchErr))
			continue
		}
		if sent {
			result.Sent++
		} else {
			result.Skipped++
		}
	}
	return result, errors.Join(failures...)
}

func (s *service) dispatchParticipant(ctx context.Context, participant notificationmodel.Participant, delivered map[int]bool) (*int, error) {
	energy := rules.EnergyPercent(participant.EnergyPercent, participant.EnergyUpdatedAt, s.now().UTC())
	selected, ok := templateForEnergy(energy)
	if !ok {
		return nil, nil
	}
	if delivered[selected.threshold] {
		return nil, nil
	}

	message := notificationmodel.Message{
		Recipient: participant.Email,
		Subject:   subject,
		TextBody:  selected.firstLine + "\n" + selected.secondLine,
		HTMLBody:  "<p><strong>" + html.EscapeString(selected.firstLine) + "</strong><br>" + html.EscapeString(selected.secondLine) + "</p>",
	}
	if err := s.mailer.Send(ctx, message); err != nil {
		return nil, fmt.Errorf("send energy notification: %w", err)
	}
	threshold := selected.threshold
	return &threshold, nil
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
