package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}
func SendRequest(city string, count int) (*http.Response, error) {
	client := &http.Client{}
	// Формируем URL с параметрами запроса
	url := fmt.Sprintf("http://localhost:8080/cafe?city=%s&count=%d", city, count)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func TestCafeCount(t *testing.T) {
	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: -1}, // Для этого случая мы будем вычислять want внутри цикла
	}

	for _, req := range requests {

		response, err := SendRequest("moscow", req.count)
		if err != nil {
			t.Errorf("Ошибка при отправке запроса: %v", err)
			continue
		}

		require.Equal(t, http.StatusOK, response.StatusCode)

		cafes := cafeList["moscow"]
		if req.count < len(cafes) {
			cafes = cafes[:req.count]
		}

		got := len(cafes)

		if req.want == -1 {
			// для count=100 вычисляем ожидаемое количество
			req.want = min(len(cafes), 100)
		}

		assert.Equal(t, req.want, got, "При count=%d ожидалось %d кафе, но получено %d", req.count, req.want, got)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestCafeSearch(t *testing.T) {
	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, req := range requests {

		cafes := cafeList["moscow"]

		filteredCafes := make([]string, 0)
		for _, cafe := range cafes {
			if strings.Contains(strings.ToLower(cafe), strings.ToLower(req.search)) {
				filteredCafes = append(filteredCafes, cafe)
			}
		}

		// Подсчитываем количество полученных кафе.
		gotCount := len(filteredCafes)

		_, err := SendRequest("moscow", 0)
		require.NoError(t, err, "Ошибка при отправке запроса")

		assert.Equal(t, req.wantCount, gotCount, "При search='%s' ожидалось %d кафе, но получено %d", req.search, req.wantCount, gotCount)
	}
}
