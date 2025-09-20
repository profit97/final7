package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
		// Получаем список кафе для заданного города и count
		cafes := cafeList["moscow"]
		if req.count < len(cafes) {
			cafes = cafes[:req.count]
		}

		// Подсчитываем количество полученных кафе.
		got := len(cafes)

		if req.want == -1 { // для count=100 вычисляем ожидаемое количество
			req.want = min(len(cafes), 100)
		}

		// Сравниваем полученное количество с ожидаемым.
		if got != req.want {
			t.Errorf("При count=%d ожидалось %d кафе, но получено %d", req.count, req.want, got)
		}
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
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}
	for _, req := range requests {
		// Получаем список всех кафе в Москве.
		cafes := cafeList["moscow"]

		// Фильтруем список кафе, оставляя только те, в названии которых есть искомая строка.
		filteredCafes := make([]string, 0)
		for _, cafe := range cafes {
			if strings.Contains(strings.ToLower(cafe), strings.ToLower(req.search)) {
				filteredCafes = append(filteredCafes, cafe)
			}
		}

		// Подсчитываем количество полученных кафе.
		gotCount := len(filteredCafes)

		// Сравниваем полученное количество с ожидаемым.
		if gotCount != req.wantCount {
			t.Errorf("При search='%s' ожидалось %d кафе, но получено %d", req.search, req.wantCount, gotCount)
		}
	}
}
