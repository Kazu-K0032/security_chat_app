package handler

import (
	"fmt"
	"log"
	"net/http"

	"security_chat_app/internal/domain"
	"security_chat_app/internal/infrastructure/firebase"
	"security_chat_app/internal/interface/markup"
	"security_chat_app/internal/interface/middleware"
)

// 設定ページのデータ構造体
type SettingsPageData struct {
	IsLoggedIn       bool         // ログイン状態
	User             *domain.User // ユーザー情報
	ShowPasswordForm bool         // パスワード変更フォームの表示状態
	ShowUsernameForm bool         // ユーザー名変更フォームの表示状態
	PasswordForm     struct {
		CurrentPassword    string // 現在のパスワード
		NewPassword        string // 新しいパスワード
		NewPasswordConfirm string // 新しいパスワードの確認
	}
	UsernameForm struct {
		NewUsername string // 新しいユーザー名
	}
	ValidationErrors         []string // バリデーションエラー
	UsernameValidationErrors []string // ユーザー名のバリデーションエラー
}

// 設定ページのハンドラ
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	// セッションの検証
	session, err := middleware.ValidateSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// ユーザー名変更の処理
	if r.URL.Path == "/settings/username" && r.Method == http.MethodPost {
		log.Printf("ユーザー名変更リクエストを受信: %s", r.URL.Path)
		// フォームデータの解析
		r.ParseForm()
		newUsername := r.FormValue("new_username")
		log.Printf("新しいユーザー名: %s", newUsername)

		// バリデーション
		var validationErrors []string
		if newUsername == "" {
			validationErrors = append(validationErrors, "新しいユーザー名を入力してください")
		} else if len(newUsername) < 2 {
			validationErrors = append(validationErrors, "ユーザー名は2文字以上で入力してください")
		} else if len(newUsername) > 20 {
			validationErrors = append(validationErrors, "ユーザー名は20文字以下で入力してください")
		}

		if len(validationErrors) > 0 {
			data := SettingsPageData{
				IsLoggedIn:       true,
				User:             session.User,
				ShowUsernameForm: true,
				UsernameForm: struct {
					NewUsername string
				}{
					NewUsername: newUsername,
				},
				UsernameValidationErrors: validationErrors,
			}
			markup.GenerateHTML(w, data, "layout", "header", "settings", "footer")
			return
		}

		// ユーザー名の更新
		log.Printf("ユーザー名更新開始: userID=%s, newUsername=%s", session.User.ID, newUsername)
		err = firebase.UpdateField("users", session.User.ID, "Name", newUsername)
		if err != nil {
			log.Printf("ユーザー名更新エラー: %v, userID=%s, newUsername=%s", err, session.User.ID, newUsername)
			validationErrors = append(validationErrors, "ユーザー名の更新に失敗しました")
			data := SettingsPageData{
				IsLoggedIn:       true,
				User:             session.User,
				ShowUsernameForm: true,
				UsernameForm: struct {
					NewUsername string
				}{
					NewUsername: newUsername,
				},
				UsernameValidationErrors: validationErrors,
			}
			markup.GenerateHTML(w, data, "layout", "header", "settings", "footer")
			return
		}

		// セッションのユーザー情報を更新
		session.User.Name = newUsername
		err = middleware.UpdateSession(w, r, session)
		if err != nil {
			http.Error(w, "セッションの更新に失敗しました", http.StatusInternalServerError)
			return
		}

		// 成功時は設定ページにリダイレクト
		http.Redirect(w, r, "/settings?success=ユーザー名を更新しました", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// パスワード変更フォームの処理
		r.ParseForm()
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		newPasswordConfirm := r.FormValue("new_password_confirm")

		// バリデーション
		var validationErrors []string
		if currentPassword == "" {
			validationErrors = append(validationErrors, "現在のパスワードを入力してください")
		}
		if newPassword == "" {
			validationErrors = append(validationErrors, "新しいパスワードを入力してください")
		}
		if newPassword != newPasswordConfirm {
			validationErrors = append(validationErrors, "新しいパスワードが一致しません")
		}

		if len(validationErrors) > 0 {
			data := SettingsPageData{
				IsLoggedIn:       true,
				User:             session.User,
				ShowPasswordForm: true,
				PasswordForm: struct {
					CurrentPassword    string // 現在のパスワード
					NewPassword        string // 新しいパスワード
					NewPasswordConfirm string // 新しいパスワードの確認
				}{
					CurrentPassword:    currentPassword,
					NewPassword:        newPassword,
					NewPasswordConfirm: newPasswordConfirm,
				},
				ValidationErrors: validationErrors,
			}
			markup.GenerateHTML(w, data, "layout", "header", "settings", "footer")
			return
		}

		// パスワード更新
		err = firebase.UpdateField("users", session.User.ID, "Password", newPassword)
		if err != nil {
			log.Printf("パスワード更新エラー: %v, userID=%s", err, session.User.ID)
			validationErrors = append(validationErrors, "パスワードの更新に失敗しました")
			data := SettingsPageData{
				IsLoggedIn:       true,
				User:             session.User,
				ShowPasswordForm: true,
				PasswordForm: struct {
					CurrentPassword    string // 現在のパスワード
					NewPassword        string // 新しいパスワード
					NewPasswordConfirm string // 新しいパスワードの確認
				}{
					CurrentPassword:    currentPassword,
					NewPassword:        newPassword,
					NewPasswordConfirm: newPasswordConfirm,
				},
				ValidationErrors: validationErrors,
			}
			markup.GenerateHTML(w, data, "layout", "header", "settings", "footer")
			return
		}

		// 成功時は設定ページにリダイレクト
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	// 設定ページのデータを取得
	data, err := getSettingsPageData(session.User, r)
	if err != nil {
		http.Error(w, "設定ページのデータの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// テンプレートのレンダリング
	markup.GenerateHTML(w, data, "layout", "header", "settings", "footer")
}

// 設定ページのデータを取得
func getSettingsPageData(user *domain.User, r *http.Request) (SettingsPageData, error) {
	if user == nil {
		return SettingsPageData{}, fmt.Errorf("ユーザー情報が無効です")
	}

	// クエリパラメータからフォームの表示状態を取得
	showPasswordForm := r.URL.Query().Get("show_password_form") == "true"
	showUsernameForm := r.URL.Query().Get("show_username_form") == "true"

	return SettingsPageData{
		IsLoggedIn:       true,
		User:             user,
		ShowPasswordForm: showPasswordForm,
		ShowUsernameForm: showUsernameForm,
	}, nil
}
