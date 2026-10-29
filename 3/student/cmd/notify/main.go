package main

import (
	"errors"
	"fmt"

	"student/practice"
)

func main() {
	fmt.Println("=== 1. Успех ===")
	email := &practice.EmailNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "email"},
		Address:      "student@example.com",
	}
	push := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     "device-001",
	}
	channels := []practice.Notifier{email, push}

	if err := practice.Dispatch(channels, "Заказ готов"); err != nil {
		fmt.Println("Неожиданная ошибка:", err)
	} else {
		fmt.Println("Успешно отправлено!")
		fmt.Println("История email:", email.Messages)
		fmt.Println("История push:", push.Messages)
	}

	fmt.Println("\n=== 2. Пустой текст ===")
	if err := practice.Dispatch(channels, ""); err != nil {
		if errors.Is(err, practice.ErrEmptyMessage) {
			fmt.Println("Распознана ошибка пустого сообщения:", err)
		} else {
			fmt.Println("Другая ошибка:", err)
		}
		fmt.Println("История email не изменилась:", email.Messages)
		fmt.Println("История push не изменилась:", push.Messages)
	}

	fmt.Println("\n=== 3. Нет настройки ===")
	badPush := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     "",
	}
	if err := practice.Dispatch([]practice.Notifier{badPush}, "Проверка настроек"); err != nil {
		var sendErr *practice.SendError
		if errors.As(err, &sendErr) {
			fmt.Printf("Ошибка настройки: канал=%q, причина=%q (полный текст: %v)\n",
				sendErr.Channel, sendErr.Reason, err)
		} else {
			fmt.Println("Другая ошибка:", err)
		}
		fmt.Println("История badPush:", badPush.Messages)
	}

	fmt.Println("\n=== 4. Сломанная реализация ===")
	broken := &practice.BrokenNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "broken"},
	}
	workingPush := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     "device-002",
	}
	errs := practice.SafeDispatch([]practice.Notifier{broken, workingPush}, "Важное уведомление")
	for _, err := range errs {
		if errors.Is(err, practice.ErrNotifierPanic) {
			fmt.Println("Перехвачена паника канала:", err)
		} else {
			fmt.Println("Другая ошибка:", err)
		}
	}
	fmt.Println("История исправного push:", workingPush.Messages)
}
