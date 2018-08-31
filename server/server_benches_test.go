package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkDateDistinctSecondsHandler(b *testing.B) {
	benchHelperSendDistinctRequest(b, func() string {
		dateTempl := "2015-08-%d %d:%d:%d"
		day := rand.Intn(3) + 1
		hour := rand.Intn(24)
		minute := rand.Intn(60)
		second := rand.Intn(60)
		date := fmt.Sprintf(dateTempl, day, hour, minute, second)
		return date
	})
}

func BenchmarkDateDistinctMinutesHandler(b *testing.B) {
	benchHelperSendDistinctRequest(b, func() string {
		dateTempl := "2015-08-%d %d:%d"
		day := rand.Intn(3) + 1
		hour := rand.Intn(24)
		minute := rand.Intn(60)
		date := fmt.Sprintf(dateTempl, day, hour, minute)
		return date
	})
}

func BenchmarkDateDistinctHoursHandler(b *testing.B) {
	benchHelperSendDistinctRequest(b, func() string {
		dateTempl := "2015-08-%d %d"
		day := rand.Intn(3) + 1
		hour := rand.Intn(24)
		date := fmt.Sprintf(dateTempl, day, hour)
		return date
	})
}

func BenchmarkDateDistinctDaysHandler(b *testing.B) {
	benchHelperSendDistinctRequest(b, func() string {
		dateTempl := "2015-08-%d"
		day := rand.Intn(3) + 1
		date := fmt.Sprintf(dateTempl, day)
		return date
	})
}

func BenchmarkDateDistinctMonthHandler(b *testing.B) {
	benchHelperSendDistinctRequest(b, func() string {
		return "2015-08"
	})
}

func BenchmarkDateDistinctYearHandler(b *testing.B) {
	benchHelperSendDistinctRequest(b, func() string {
		return "2015"
	})
}

func benchHelperSendDistinctRequest(b *testing.B, getRandDateFn func() string) {
	urlTmpl := "/1/queries/count/%s"
	for n := 0; n < b.N; n++ {
		date := getRandDateFn()
		reqURL := fmt.Sprintf(urlTmpl, date)
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			b.Fatalf("could not create a benchmark request: %v", err)
		}
		rec := httptest.NewRecorder()
		s.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Errorf("expected HTTP status code: %d, actual code: %d", http.StatusOK, rec.Code)
		}
	}
}

func BenchmarkDateUniqueCountSecondsHandler(b *testing.B) {
	benchHelperSendUniqueCountRequest(b, func() string {
		dateTempl := "2015-08-%d %d:%d:%d"
		day := rand.Intn(3) + 1
		hour := rand.Intn(24)
		minute := rand.Intn(60)
		second := rand.Intn(60)
		date := fmt.Sprintf(dateTempl, day, hour, minute, second)
		return date
	})
}

func BenchmarkDateUniqueCountMinutesHandler(b *testing.B) {
	benchHelperSendUniqueCountRequest(b, func() string {
		dateTempl := "2015-08-%d %d:%d"
		day := rand.Intn(3) + 1
		hour := rand.Intn(24)
		minute := rand.Intn(60)
		date := fmt.Sprintf(dateTempl, day, hour, minute)
		return date
	})
}

func BenchmarkDateUniqueCountHoursHandler(b *testing.B) {
	benchHelperSendUniqueCountRequest(b, func() string {
		dateTempl := "2015-08-%d %d"
		day := rand.Intn(3) + 1
		hour := rand.Intn(24)
		date := fmt.Sprintf(dateTempl, day, hour)
		return date
	})
}

func BenchmarkDateUniqueCountDaysHandler(b *testing.B) {
	benchHelperSendUniqueCountRequest(b, func() string {
		dateTempl := "2015-08-%d"
		day := rand.Intn(3) + 1
		date := fmt.Sprintf(dateTempl, day)
		return date
	})
}

func BenchmarkDateUniqueCountMonthHandler(b *testing.B) {
	benchHelperSendUniqueCountRequest(b, func() string {
		return "2015-08"
	})
}

func BenchmarkDatUniqueCountYearHandler(b *testing.B) {
	benchHelperSendUniqueCountRequest(b, func() string {
		return "2015"
	})
}

func benchHelperSendUniqueCountRequest(b *testing.B, getRandDateFn func() string) {
	urlTmpl := "/1/queries/count/%s?size=%d"
	for n := 0; n < b.N; n++ {
		date := getRandDateFn()
		size := rand.Intn(20) + 1
		reqURL := fmt.Sprintf(urlTmpl, date, size)
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			b.Fatalf("could not create a benchmark request: %v", err)
		}
		rec := httptest.NewRecorder()
		s.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Errorf("expected HTTP status code: %d, actual code: %d", http.StatusOK, rec.Code)
		}
	}
}
