package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	authmodel "github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
	"github.com/google/uuid"
)

type fakeUsers struct {
	user *authmodel.User
	err  error
}

func (f fakeUsers) FindUser(context.Context, uuid.UUID) (*authmodel.User, error) {
	return f.user, f.err
}

type fakePets struct {
	stats *petmodel.Stats
	err   error
}

func (f fakePets) GetPet(context.Context, uuid.UUID) (*petmodel.Stats, error) {
	return f.stats, f.err
}

type recordingMailer struct {
	message notificationmodel.Message
	calls   int
	err     error
}

func (m *recordingMailer) Send(_ context.Context, message notificationmodel.Message) error {
	m.calls++
	m.message = message
	return m.err
}

func TestEnergyThresholdSelection(t *testing.T) {
	tests := []struct {
		energy    int
		status    notificationmodel.Status
		threshold *int
	}{
		{100, notificationmodel.StatusSkipped, nil},
		{51, notificationmodel.StatusSkipped, nil},
		{50, notificationmodel.StatusSent, intPointer(50)},
		{26, notificationmodel.StatusSent, intPointer(50)},
		{25, notificationmodel.StatusSent, intPointer(25)},
		{6, notificationmodel.StatusSent, intPointer(25)},
		{5, notificationmodel.StatusSent, intPointer(5)},
		{1, notificationmodel.StatusSent, intPointer(5)},
		{0, notificationmodel.StatusSent, intPointer(0)},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%d", test.status, test.energy), func(t *testing.T) {
			mailer := &recordingMailer{}
			service := New(fakeUsers{user: &authmodel.User{Email: "registered@example.com"}}, fakePets{stats: &petmodel.Stats{Energy: test.energy}}, mailer)
			result, err := service.DispatchEnergy(context.Background(), uuid.New())
			if err != nil {
				t.Fatalf("dispatch: %v", err)
			}
			if result.Status != test.status || result.Energy != test.energy || !equalPointers(result.Threshold, test.threshold) {
				t.Fatalf("result = %+v, want status %q energy %d threshold %v", result, test.status, test.energy, test.threshold)
			}
			wantCalls := 1
			if test.status == notificationmodel.StatusSkipped {
				wantCalls = 0
			}
			if mailer.calls != wantCalls {
				t.Fatalf("mailer calls = %d, want %d", mailer.calls, wantCalls)
			}
		})
	}
}

func TestTemplatesContainExactCopyAndFormatting(t *testing.T) {
	tests := []struct {
		energy int
		first  string
		second string
	}{
		{50, "Я по тебе соскучился... 🥺", "Навестишь меня?"},
		{25, "Я что-то совсем без сил... 😥", "Энергии почти не осталось. Заглянешь зарядить меня?"},
		{5, "Кажется, я сейчас совсем сяду… 🪫", "Когда энергия закончится, я потеряю накопленный прогресс. Поможешь мне сохранить его?"},
		{0, "Я полностью разрядился... 😭", "Весь мой накопленный опыт стирается. Надеюсь, мы сможем еще увидеться."},
	}
	for _, test := range tests {
		t.Run(test.first, func(t *testing.T) {
			mailer := &recordingMailer{}
			service := New(fakeUsers{user: &authmodel.User{Email: "registered@example.com"}}, fakePets{stats: &petmodel.Stats{Energy: test.energy}}, mailer)
			if _, err := service.DispatchEnergy(context.Background(), uuid.New()); err != nil {
				t.Fatalf("dispatch: %v", err)
			}
			if mailer.message.Subject != "Питомец ждёт вас" {
				t.Fatalf("subject = %q", mailer.message.Subject)
			}
			if mailer.message.TextBody != test.first+"\n"+test.second {
				t.Fatalf("text body = %q", mailer.message.TextBody)
			}
			if strings.Contains(mailer.message.TextBody, "<") {
				t.Fatalf("plain text contains HTML: %q", mailer.message.TextBody)
			}
			wantHTML := "<p><strong>" + test.first + "</strong><br>" + test.second + "</p>"
			if mailer.message.HTMLBody != wantHTML {
				t.Fatalf("HTML body = %q, want %q", mailer.message.HTMLBody, wantHTML)
			}
		})
	}
}

func TestDispatchUsesRegisteredEmail(t *testing.T) {
	mailer := &recordingMailer{}
	service := New(fakeUsers{user: &authmodel.User{Email: "registered@example.com"}}, fakePets{stats: &petmodel.Stats{Energy: 25}}, mailer)
	if _, err := service.DispatchEnergy(context.Background(), uuid.New()); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if mailer.message.Recipient != "registered@example.com" {
		t.Fatalf("recipient = %q", mailer.message.Recipient)
	}
}

func TestDispatchPropagatesDependencyErrors(t *testing.T) {
	dependencyError := errors.New("dependency failed")
	tests := []struct {
		name   string
		users  UserFinder
		pets   PetFinder
		mailer *recordingMailer
	}{
		{"user", fakeUsers{err: dependencyError}, fakePets{stats: &petmodel.Stats{Energy: 25}}, &recordingMailer{}},
		{"pet", fakeUsers{user: &authmodel.User{Email: "registered@example.com"}}, fakePets{err: dependencyError}, &recordingMailer{}},
		{"mailer", fakeUsers{user: &authmodel.User{Email: "registered@example.com"}}, fakePets{stats: &petmodel.Stats{Energy: 25}}, &recordingMailer{err: dependencyError}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.users, test.pets, test.mailer).DispatchEnergy(context.Background(), uuid.New())
			if !errors.Is(err, dependencyError) {
				t.Fatalf("error = %v, want wrapped dependency error", err)
			}
		})
	}
}

func intPointer(value int) *int { return &value }

func equalPointers(left, right *int) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}
