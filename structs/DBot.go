package structs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Глобальный объект, который представляет сообщение о погоде.
var messageObject *CurrentWeatherMessage = &CurrentWeatherMessage{}

// Глобальный список объектов, который представляет сообщение о прогнозе на пять дней.
var forecastHours []*ForecastMessage

type DBot struct{}

// Метод, который получает координаты города по названию.
func (d *DBot) getCoords(city string) (float64, float64) {
	// Восстанавливаемся от паники, вдруг погода плохая.
	defer d.recoverFromPanic()

	var cityCoords []CityCoords

	// Отправляем запрос к API.
	resp, err := http.Get("https://api.openweathermap.org/geo/1.0/direct?q=" + city + "&limit=5&appid=0e8cf5b1fb682bd1754b07bfbbd7f038")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Читаем тело запроса.
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	// Преобразовываем тело запроса в объект CityCoords
	unmarhallErr := json.Unmarshal(respBody, &cityCoords)
	if unmarhallErr != nil {
		log.Fatal(unmarhallErr)
	}

	// Сразу добавляем полученную информацию из первого запроса.
	messageObject.Name = cityCoords[0].Name
	messageObject.Country = cityCoords[0].Country
	messageObject.State = cityCoords[0].State

	return cityCoords[0].Lat, cityCoords[0].Lon
}

// Метод, который получает информацию о текущей погоде по координатам.
func (d *DBot) GetWeather(city string) {
	// Восстанавливаемся от паники, вдруг погода плохая.
	defer d.recoverFromPanic()

	var curWeather CurrentWeather

	// Получаем координаты. Широта и долгота.
	lat, lon := d.getCoords(city)

	// Значения координат типа float64, но нам нужна строка.
	latString := fmt.Sprintf("%f", lat)
	lonString := fmt.Sprintf("%f", lon)

	// Отправляем второй запрос.
	resp, err := http.Get("https://api.openweathermap.org/data/2.5/weather?lat=" + latString + "&lon=" + lonString + "&appid=0e8cf5b1fb682bd1754b07bfbbd7f038&units=metric")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Читаем тело запроса.
	respBytes, readBytesErr := io.ReadAll(resp.Body)
	if readBytesErr != nil {
		log.Fatal(readBytesErr)
	}

	// Преобразовываем в объект CurrentWeather.
	unmarshallErr := json.Unmarshal(respBytes, &curWeather)
	if unmarshallErr != nil {
		log.Fatal(unmarshallErr)
	}

	// Сразу заполняем наш объект сообщения.
	messageObject.WeatherDescription = curWeather.Weather[0].Description
	messageObject.Temperature = curWeather.Main.Temp
	messageObject.FeelsLike = curWeather.Main.FeelsLike
	messageObject.Pressure = curWeather.Main.Pressure
	messageObject.Humidity = curWeather.Main.Humidity
	messageObject.WindSpeed = curWeather.Wind.Speed
	messageObject.Sunrise = time.Unix(curWeather.Sys.Sunrise, 0).Format("2006-01-02 15:04:05")
	messageObject.Sunset = time.Unix(curWeather.Sys.Sunset, 0).Format("2006-01-02 15:04:05")
}

// Метод, который получает прогноз погоды на пять дней.
func (d *DBot) GetForecast(city string) {
	defer d.recoverFromPanic()

	var forecastObj WeatherForecast

	// Получаем координаты. Широта и долгота.
	lat, lon := d.getCoords(city)

	// Значения координат типа float64, но нам нужна строка.
	latString := fmt.Sprintf("%f", lat)
	lonString := fmt.Sprintf("%f", lon)

	resp, err := http.Get("https://api.openweathermap.org/data/2.5/forecast?lat=" + latString + "&lon=" + lonString + "&appid=0e8cf5b1fb682bd1754b07bfbbd7f038&units=metric")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	respBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	unmarshallErr := json.Unmarshal(respBytes, &forecastObj)
	if unmarshallErr != nil {
		log.Fatal(unmarshallErr)
	}

	for _, obj := range forecastObj.Forecasts {
		forecastHours = append(forecastHours, &ForecastMessage{
			Date:        time.Unix(obj.DateTime, 0).Format("2006-01-02 15:04:05"),
			Description: obj.WeatherObj[0].Description,
			Temperature: obj.MainObj.Temp,
			FeelsLike:   obj.MainObj.FeelsLike,
			Pressure:    obj.MainObj.Pressure,
			Humidity:    obj.MainObj.Humidity,
			WindSpeed:   obj.WindObj.Speed,
		})
	}
}

// Метод, который анализирует отправленные боту команды и соответствующее сообщение.
func (d *DBot) SendWeatherMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Исключаем реакцию бота на свои же сообщения.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Если вводим !help, то получаем справку.
	// Здесь чтение файла происходит в горутине. Если представить ситуацию,
	// что вдруг текстовый файл будет большим, то лучше прочтем его в отдельной горунтине.
	if strings.Contains(m.Content, "!help") {
		var msgChan chan string = make(chan string)

		go d.getHelp(msgChan)
		helpMessage := <-msgChan
		d.sendMessage(s, m.ChannelID, helpMessage)
	}

	// Если комада содержит !w, значит мы хотим погоду.
	if strings.Contains(m.Content, "!w") {
		d.GetWeather(m.Content[2:])
		weatherMessage := d.generateWeatherMessage()
		d.sendMessage(s, m.ChannelID, weatherMessage)
	}

	// Если команда содержит !f, значит мы хотим прогноз на пять дней.
	if strings.Contains(m.Content, "!f") {
		d.GetForecast(m.Content[2:])
		d.sendMessage(s, m.ChannelID, d.generateForecastMessage())
	}
}

// Метод, который отправляет сообщения.
func (d *DBot) sendMessage(s *discordgo.Session, chanId, message string) {
	_, err := s.ChannelMessageSend(chanId, message)
	if err != nil {
		log.Println(err)
	}
}

// Метод, который генерирует сообщение для отправки.
func (d *DBot) generateWeatherMessage() string {
	tNow := time.Now()
	formattedDate := fmt.Sprintf("%d %s %d", tNow.Year(), tNow.Month(), tNow.Day())

	messageToSend := "Today's date: " + formattedDate + "\n" +
		"City: " + messageObject.Name + "\n" +
		"Country: " + messageObject.Country + "\n" +
		"State: " + messageObject.State + "\n" +
		"Weather description: " + messageObject.WeatherDescription + "\n" +
		"Temperature: " + fmt.Sprintf("%.0f", messageObject.Temperature) + "\n" +
		"Feels like: " + fmt.Sprintf("%.0f", messageObject.FeelsLike) + "\n" +
		"Pressure: " + fmt.Sprintf("%d", messageObject.Pressure) + "\n" +
		"Humidity: " + fmt.Sprintf("%d", messageObject.Humidity) + "\n" +
		"WindSpeed: " + fmt.Sprintf("%.0f", messageObject.WindSpeed) + " m/s" + "\n" +
		"Sunrise: " + messageObject.Sunrise + "\n" +
		"Sunset: " + messageObject.Sunset + "\n"

	return messageToSend
}

// Метод, который генерирует сообщение для отправки прогноза на пять дней.
func (d *DBot) generateForecastMessage() string {
	var messageToSend string

	for i := 1; i < len(forecastHours); i += 8 {
		messageToSend += "Date: " + forecastHours[i].Date + "\n" +
			"Description: " + forecastHours[i].Description + "\n" +
			"Temperature: " + fmt.Sprintf("%.0f", forecastHours[i].Temperature) + "\n" +
			"Feels Like: " + fmt.Sprintf("%.0f", forecastHours[i].FeelsLike) + "\n" +
			"Pressure: " + fmt.Sprintf("%d", forecastHours[i].Pressure) + "\n" +
			"Humidity: " + fmt.Sprintf("%d", forecastHours[i].Humidity) + "\n" +
			"Wind speed: " + fmt.Sprintf("%.0f", forecastHours[i].WindSpeed) + " m/s" + "\n" +
			"\n" + "\n"
	}

	return messageToSend
}

// Метод, который читает текстовый файл справки.
// Должен быть вызван как горутина.
func (d *DBot) getHelp(msgChan chan string) {
	file, err := os.OpenFile("help.txt", os.O_RDONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, readErr := io.ReadAll(file)
	if readErr != nil {
		log.Fatal(readErr)
	}

	msgChan <- string(bytes)
}

// Метод, для восстановления от паники.
func (d *DBot) recoverFromPanic() {
	if r := recover(); r != nil {
		log.Println("RECOVERED! ", r)
	}
}
