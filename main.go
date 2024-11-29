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
	var cities = []string{"Amsterdam", "Wien", "Moskow", "Warsaw", "Paris"}

	for _, city := range cities {
		wg.Add(1)
		go func(city string) {
			defer wg.Done()

			coords, err := GetCoords(city)
			if err != nil {
				fmt.Println("Error fetching coordinates:", err)
				return
			}

			weather, err := SendWeatherInfo(coords.Lat, coords.Lon)
			if err != nil {
				fmt.Println("Error fetching weather info:", err)
				return
			}

			fmt.Printf("City: %s, Temperature: %.1f°C, Windspeed: %.2f m/s\n",
				city, weather.CurrentWeather.Temperature, weather.CurrentWeather.Windspeed)
		}(city)
	}

	wg.Wait()
	tt := time.Since(start)
	fmt.Println(tt)
}

func GetCoords(city string) (GeoRespch, error) {
	geocodeURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?city=%s&format=json", city)
	geocodeResp, err := http.Get(geocodeURL)
	if err != nil {
		return GeoRespch{}, fmt.Errorf("failed to fetch coordinates: %w", err)
	}
	defer geocodeResp.Body.Close()

	var geocode []GeoRespch
	if err := json.NewDecoder(geocodeResp.Body).Decode(&geocode); err != nil || len(geocode) == 0 {
		return GeoRespch{}, fmt.Errorf("failed to decode city coordinates: %w", err)
	}

	lat, lon := geocode[0].Lat, geocode[0].Lon
	return GeoRespch{Lat: lat, Lon: lon}, nil
}

func SendWeatherInfo(lat, lon string) (WeatherResponseResult, error) {
	weatherURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current_weather=true", lat, lon)
	weatherRespch, err := http.Get(weatherURL)
	if err != nil {
		return WeatherResponseResult{}, fmt.Errorf("failed to fetch weather info: %w", err)
	}
	defer weatherRespch.Body.Close()

	var weather WeatherResponse
	if err := json.NewDecoder(weatherRespch.Body).Decode(&weather); err != nil {
		return WeatherResponseResult{}, fmt.Errorf("failed to decode weather info: %w", err)
	}

	return WeatherResponseResult{
		CurrentWeather: weather.CurrentWeather,
	}, nil
}
