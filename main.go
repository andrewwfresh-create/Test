package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	// ❗️ НОВА ЗАЛЕЖНІСТЬ, яку ми додали
	"github.com/joho/godotenv"
)

// GiphyResponse — це структура верхнього рівня для відповіді GIPHY API.
// Вона використовує ВКАЗІВНИК у зрізі, як вимагалося у завданні.
type GiphyResponse struct {
	Data []*GiphyData `json:"data"`
}

// GiphyData містить поля, які нас цікавлять: назву та посилання.
type GiphyData struct {
	Title string `json:"title"`
	URL   string `json:"url"` // Це посилання на сторінку GIPHY
}

// SearchGif — ця функція залишається БЕЗ ЗМІН
func SearchGif(apiKey, query string, limit int) (*GiphyResponse, error) {

	// 1. Формуємо URL-адресу із параметрами
	baseURL := "https://api.giphy.com/v1/gifs/search"

	// Використовуємо net/url для безпечного кодування параметрів запиту
	params := url.Values{}
	params.Add("api_key", apiKey)
	params.Add("q", query)
	params.Add("limit", strconv.Itoa(limit))

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 2. Виконуємо HTTP GET-запит (бібліотека net/http)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("помилка створення запиту: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	// Обробка помилок мережі (вимога #6)
	if err != nil {
		return nil, fmt.Errorf("помилка виконання запиту (проблеми з мережею): %w", err)
	}
	defer resp.Body.Close()

	// 3. Обробка помилок сервера (вимога #6)
	if resp.StatusCode != http.StatusOK {
		// Спеціальна обробка для невірного ключа API
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("помилка сервера (%d): невірний API-ключ. Перевірте ваш .env файл", resp.StatusCode)
		}
		return nil, fmt.Errorf("помилка сервера (%d): %s", resp.StatusCode, resp.Status)
	}

	// 4. Читаємо тіло відповіді
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("помилка читання відповіді: %w", err)
	}

	// 5. Десеріалізуємо JSON у наші структури (бібліотека encoding/json)
	var apiResponse GiphyResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("помилка розбору JSON: %w", err)
	}

	// Повертаємо вказівник на заповнену структуру
	return &apiResponse, nil
}

// --- ❗️ ФУНКЦІЯ MAIN() З ОНОВЛЕННЯМИ ---
func main() {
	log.Println("--- Запуск GIPHY-пошуку ---")

	// ❗️ НОВИЙ КОД: Завантажуємо змінні з .env файлу
	// Ця функція шукає файл .env у поточній теці
	// і завантажує його змінні в середовище (Environment)
	err := godotenv.Load()
	if err != nil {
		// Це не критична помилка, якщо .env немає,
		// ключ все ще може бути встановлений у системі.
		log.Println("Увага: не вдалося завантажити файл .env. " +
			"Спроба використати системні змінні оточення.")
	}

	// 1. Отримуємо API ключ (тепер він доступний через os.Getenv)
	apiKey := os.Getenv("GIPHY_API_KEY")
	if apiKey == "" {
		// Оновлене повідомлення про помилку
		log.Fatal("ПОМИЛКА: Змінна GIPHY_API_KEY не встановлена.\n" +
			"Переконайтеся, що ви створили .env файл, або встановили її системно.")
	}

	// 2. Запитуємо вхідні дані у користувача (бібліотеки os/bufio)
	// ... (ця частина залишається без змін) ...
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Введіть ключове слово (наприклад, 'теніс'): ")
	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query)

	if query == "" {
		log.Fatal("ПОМИЛКА: Ключове слово не може бути порожнім.")
	}

	fmt.Print("Введіть кількість результатів: ")
	limitStr, _ := reader.ReadString('\n')
	limitStr = strings.TrimSpace(limitStr)

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		log.Fatalf("ПОМИЛКА: Кількість результатів має бути додатним числом. Ви ввели: '%s'", limitStr)
	}

	// 3. Викликаємо нашу функцію для роботи з API
	// ... (ця частина залишається без змін) ...
	log.Println("...Пошук GIF...")
	response, err := SearchGif(apiKey, query, limit)
	if err != nil {
		log.Fatalf("Не вдалося отримати GIF: %v", err)
	}

	// 4. Виводимо результат у консоль
	// ... (ця частина залишається без змін) ...
	if len(response.Data) == 0 {
		fmt.Printf("\nНа жаль, за запитом '%s' нічого не знайдено.\n", query)
		return
	}

	fmt.Printf("\n--- Результати пошуку для '%s' (%d шт.) ---\n", query, len(response.Data))

	for i, gif := range response.Data {
		if gif != nil {
			fmt.Printf("\n%d. Назва: %s\n", i+1, gif.Title)
			fmt.Printf("   Посилання: %s\n", gif.URL)
		}
	}
	fmt.Println("\n--- Кінець результатів ---")
}
