package structs

// Структура, из которой мы можем получить координаты.
// Заполняется при первом запросе к API.
type CityCoords struct {
	Name       string
	LocalNames []LocalNames
	Lat        float64
	Lon        float64
	Country    string
	State      string
}

type LocalNames struct {
	En string
}

// Структура, из которой мы можем получить информацию о погоде.
// Заполняется при втором запросе к API.
type CurrentWeather struct {
	Weather []Weather
	Main    Main
	Wind    Wind
	Sys     Sys
}

type Weather struct {
	Description string
}

type Main struct {
	Temp      float64
	FeelsLike float64 `json:"feels_like"`
	Pressure  int
	Humidity  int
}

type Wind struct {
	Speed float64
}

type Sys struct {
	Sunrise int64
	Sunset  int64
}

// Структура содержит в себе прогноз погоды на три дня.
type WeatherForecast struct {
	Forecasts []Forecast `json:"list"`
}

type Forecast struct {
	DateTime   int64              `json:"dt"`
	MainObj    ForecastParameters `json:"main"`
	WeatherObj []WeatherObj       `json:"weather"`
	WindObj    WindObj            `json:"wind"`
}

type ForecastParameters struct {
	Temp      float64
	FeelsLike float64 `json:"feels_like"`
	Pressure  int
	Humidity  int
}

type WeatherObj struct {
	Description string
}

type WindObj struct {
	Speed float64
}

// Структура представляет объект отправляемого сообщения с текущей погодой.
type CurrentWeatherMessage struct {
	Name               string
	Country            string
	State              string
	WeatherDescription string
	Temperature        float64
	FeelsLike          float64
	Pressure           int
	Humidity           int
	WindSpeed          float64
	Sunrise            string
	Sunset             string
}

// Структура представляет объект отправляемого сообщения с прогнозом на три дня.
type ForecastMessage struct {
	Date        string
	Description string
	Temperature float64
	FeelsLike   float64
	Pressure    int
	Humidity    int
	WindSpeed   float64
}
