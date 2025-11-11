package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"unicode"
)

const (
	cepUrl     = "https://viacep.com.br/ws/"
	weatherUrl = "http://api.weatherapi.com/v1/current.json?key=3e67e3649e5e49bab99153600251111&aqi=no"
)

type cepInfo struct {
	Localidade string `json:"localidade"`
}

type cityInfo struct {
	Current struct {
		TempC float64 `json:"temp_c"`
		TempF float64 `json:"temp_f"`
	} `json:"current"`
}

type city struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func main() {
	http.HandleFunc("/weather", weatherHandler)

	addr := ":8080"
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cep := r.URL.Query().Get("cep")
	if !isValidCEP(cep) {
		http.Error(w, "cep must have exactly 8 digits", http.StatusUnprocessableEntity)
		return
	}

	cepInformation := cepInfo{}
	cityInformation := cityInfo{}

	resp, err := http.Get(fmt.Sprintf("%s%s/json/", cepUrl, cep))
	if err != nil {
		http.Error(w, "failed to reach CEP service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		http.Error(w, "cep not found", http.StatusNotFound)
		return
	case http.StatusOK:
	default:
		http.Error(w, "cep service error", http.StatusBadGateway)
		return
	}

	if err := json.NewDecoder(resp.Body).Decode(&cepInformation); err != nil {
		http.Error(w, "invalid cep response", http.StatusBadGateway)
		return
	}

	respW, err := http.Get(weatherUrl + "&q=" + url.QueryEscape(cepInformation.Localidade))
	if err != nil {
		http.Error(w, "failed to reach weather service", http.StatusBadGateway)
		return
	}
	defer respW.Body.Close()

	if respW.StatusCode != http.StatusOK {
		http.Error(w, "can not find zipcode", http.StatusBadGateway)
		return
	}

	if err := json.NewDecoder(respW.Body).Decode(&cityInformation); err != nil {
		http.Error(w, "invalid weather response", http.StatusBadGateway)
		return
	}

	cityResponse := city{
		TempC: cityInformation.Current.TempC,
		TempF: cityInformation.Current.TempF,
		TempK: cityInformation.Current.TempC + 273.15,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cityResponse); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func isValidCEP(cep string) bool {
	if len(cep) != 8 {
		return false
	}
	for _, r := range cep {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
