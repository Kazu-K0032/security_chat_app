package middleware

import (
	"context"
	"net/http"
)

// contextKey コンテキストのキーとして使用するカスタム型
type contextKey string

const templateDataKey contextKey = "templateData"

// Middleware セッション管理のミドルウェア
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// セッションの検証
		session, err := ValidateSession(w, r)
		if err != nil {
			// セッションが無効な場合は、ログインしていない状態として処理
			data := TemplateData{IsLoggedIn: false}
			r = r.WithContext(context.WithValue(r.Context(), templateDataKey, data))
		} else {
			// セッションが有効な場合は、ログインしている状態として処理
			data := service.TemplateData{
				IsLoggedIn: true,
				User:       session.User,
			}
			r = r.WithContext(context.WithValue(r.Context(), templateDataKey, data))
		}

		next.ServeHTTP(w, r)
	})
}
