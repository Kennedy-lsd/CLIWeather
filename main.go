package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type GeoRespch struct {
	Lat   string `json:"lat"`
	Lon   string `json:"lon"`
	Error error
}

type WeatherResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
	} `json:"current_weather"`
}

type WeatherResponseResult struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
	} `json:"current_weather"`
	Error error `json:"error"`
}

func main() {
	start := time.Now()
	var wg sync.WaitGroup
	geoCh := make(chan GeoRespch)
	weatherCh := make(chan WeatherResponseResult)

	var cities = []string{"Amsterdam", "Wien", "Moskow", "Warsaw", "Paris"}

	for _, city := range cities {
		wg.Add(5)

		go GetCords(city, geoCh, &wg)

		go func() {
			wg.Wait()
			close(geoCh)
		}()

		coords := <-geoCh
		if coords.Error != nil {
			fmt.Println("Error fetching coordinates:", coords.Error)
			return
		}

		wg.Add(5)
		go SendWeatherInfo(coords.Lat, coords.Lon, weatherCh, &wg)

		go func() {
			wg.Wait()
			close(weatherCh)
		}()

		weather := <-weatherCh
		if weather.Error != nil {
			fmt.Println("Error fetching weather info:", weather.Error)
			return
		}

		fmt.Printf("City: %s, Temperature: %.1f°C, Windspeed: %.2f m/s\n",
			city, weather.CurrentWeather.Temperature, weather.CurrentWeather.Windspeed)
	}
	tt := time.Since(start)
	fmt.Println(tt)
}

func GetCords(city string, ch chan<- GeoRespch, wg *sync.WaitGroup) {
	defer wg.Done()

	geocodeURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?city=%s&format=json", city)
	geocodeResp, err := http.Get(geocodeURL)
	if err != nil {
		ch <- GeoRespch{Error: fmt.Errorf("failed to fetch coordinates: %w", err)}
		return
	}
	defer geocodeResp.Body.Close()

	var geocode []GeoRespch
	if err := json.NewDecoder(geocodeResp.Body).Decode(&geocode); err != nil || len(geocode) == 0 {
		ch <- GeoRespch{Error: fmt.Errorf("failed to decode city coordinates: %w", err)}
		return
	}

	lat, lon := geocode[0].Lat, geocode[0].Lon
	ch <- GeoRespch{Lat: lat, Lon: lon, Error: nil}
}

func SendWeatherInfo(lat, lon string, ch chan<- WeatherResponseResult, wg *sync.WaitGroup) {
	defer wg.Done()

	weatherURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current_weather=true", lat, lon)
	weatherRespch, err := http.Get(weatherURL)
	if err != nil {
		ch <- WeatherResponseResult{Error: fmt.Errorf("failed to fetch weather info: %w", err)}
		return
	}
	defer weatherRespch.Body.Close()

	var weather WeatherResponse
	if err := json.NewDecoder(weatherRespch.Body).Decode(&weather); err != nil {
		ch <- WeatherResponseResult{Error: fmt.Errorf("failed to decode weather info: %w", err)}
		return
	}

	ch <- WeatherResponseResult{
		CurrentWeather: weather.CurrentWeather,
		Error:          nil,
	}
}
