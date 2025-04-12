package handler

import (
	"log"
	"net/http"

	"security_chat_app/internal/domain"
	"security_chat_app/internal/infrastructure/repository"
	"security_chat_app/internal/interface/markup"
	"security_chat_app/internal/interface/middleware"
)

// LoginHandler ログイン処理を実行
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// ログイン画面の表示
	if r.Method == http.MethodGet {
		data := domain.TemplateData{
			IsLoggedIn: false,
		}
		markup.GenerateHTML(w, data, "layout", "header", "login", "footer")
		return
	}

	// ログイン処理
	if r.Method == http.MethodPost {
		r.ParseForm()
		form := domain.LoginForm{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		// バリデーション
		var validationErrors []string
		if form.Email == "" {
			validationErrors = append(validationErrors, "メールアドレスを入力してください")
		}
		if form.Password == "" {
			validationErrors = append(validationErrors, "パスワードを入力してください")
		}

		if len(validationErrors) > 0 {
			log.Printf("バリデーションエラー: %v", validationErrors)
			data := domain.TemplateData{
				IsLoggedIn:       false,
				LoginForm:        domain.LoginForm{Email: form.Email, Password: form.Password},
				ValidationErrors: validationErrors,
			}
			markup.GenerateHTML(w, data, "layout", "header", "login", "footer")
			return
		}

		// ユーザー認証
		user, err := repository.GetUserByEmail(form.Email)
		if err != nil {
			log.Printf("ユーザー認証エラー: %v", err)
			data := domain.TemplateData{
				IsLoggedIn:       false,
				LoginForm:        domain.LoginForm{Email: form.Email, Password: form.Password},
				ValidationErrors: []string{"認証エラーが発生しました"},
			}
			markup.GenerateHTML(w, data, "layout", "header", "login", "footer")
			return
		}

		if user == nil || user.Password != form.Password {
			log.Printf("ユーザー認証エラー: %v", err)
			data := domain.TemplateData{
				IsLoggedIn:       false,
				LoginForm:        domain.LoginForm{Email: form.Email, Password: form.Password},
				ValidationErrors: []string{"メールアドレスまたはパスワードが誤っています"},
			}
			markup.GenerateHTML(w, data, "layout", "header", "login", "footer")
			return
		}

		// セッションの作成
		session, err := middleware.CreateSession(user)
		if err != nil {
			log.Printf("セッション作成エラー: %v", err)
			data := domain.TemplateData{
				IsLoggedIn:       false,
				LoginForm:        domain.LoginForm{Email: form.Email, Password: form.Password},
				ValidationErrors: []string{"セッション作成エラーが発生しました"},
			}
			markup.GenerateHTML(w, data, "layout", "header", "login", "footer")
			return
		}

		// セッションクッキーの設定
		middleware.SetSessionCookie(w, session)
		// ログイン成功後、プロフィールページにリダイレクト
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
}
