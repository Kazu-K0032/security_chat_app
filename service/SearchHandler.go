package service

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"security_chat_app/repository"
)

// 検索ページのデータ構造体
type SearchPageData struct {
	IsLoggedIn bool
	User       *repository.User
	Query      string
	Users      []map[string]interface{}
}

// 検索ハンドラ
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// セッションの検証
	session, err := ValidateSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 検索ページのデータを取得
	data, err := getSearchPageData(session.User, r)
	if err != nil {
		http.Error(w, "検索データの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// テンプレートのレンダリング
	GenerateHTML(w, data, "layout", "header", "search", "footer")
}

// 検索ページのデータを取得
func getSearchPageData(user *repository.User, r *http.Request) (SearchPageData, error) {
	if user == nil {
		return SearchPageData{}, fmt.Errorf("ユーザー情報が無効です")
	}

	// 検索クエリを取得
	query := r.URL.Query().Get("username")
	var users []map[string]interface{}
	var err error

	// 検索クエリがある場合は検索を実行、ない場合は全ユーザーを取得
	if query != "" {
		users, err = repository.SearchUsers(query)
	} else {
		users, err = repository.GetAllData("users")
	}

	if err != nil {
		return SearchPageData{}, fmt.Errorf("ユーザー情報の取得に失敗しました: %v", err)
	}

	// デバッグ用のログ
	fmt.Printf("取得した全ユーザー数: %d\n", len(users))
	fmt.Printf("現在のユーザーID: %s\n", user.ID)

	// 自分以外のユーザーをフィルタリングし、新規追加順にソート
	var filteredUsers []map[string]interface{}
	for _, u := range users {
		userID := u["id"].(string)
		fmt.Printf("ユーザーID: %s\n", userID)
		if userID != user.ID {
			filteredUsers = append(filteredUsers, u)
		}
	}

	fmt.Printf("フィルタリング後のユーザー数: %d\n", len(filteredUsers))

	// 新規追加順にソート（created_atの降順）
	sort.Slice(filteredUsers, func(i, j int) bool {
		timeI := filteredUsers[i]["created_at"].(time.Time)
		timeJ := filteredUsers[j]["created_at"].(time.Time)
		return timeI.After(timeJ)
	})

	// 検索ページのデータを取得
	data := SearchPageData{
		IsLoggedIn: true,
		User:       user,
		Query:      query,
		Users:      filteredUsers,
	}

	return data, nil
}

// チャット開始ハンドラ
func StartChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// セッションの検証
	session, err := ValidateSession(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// リクエストボディからデータを取得
	targetUserID := r.FormValue("userID")
	if targetUserID == "" {
		http.Error(w, "ユーザーIDが指定されていません", http.StatusBadRequest)
		return
	}

	// チャットを開始
	chatID, err := repository.StartChat(session.User.ID, targetUserID)
	if err != nil {
		http.Error(w, "チャットの開始に失敗しました", http.StatusInternalServerError)
		return
	}

	// チャットページにリダイレクト
	http.Redirect(w, r, fmt.Sprintf("/chat?chat_id=%s", chatID), http.StatusSeeOther)
}

// ユーザーを検索
func SearchUsers(query string) ([]map[string]interface{}, error) {
	// ユーザーを検索
	users, err := repository.SearchUsers(query)
	if err != nil {
		return nil, fmt.Errorf("ユーザーの検索に失敗しました: %v", err)
	}

	return users, nil
}

// ユーザー情報を取得
func GetUserData(userID string) (*repository.User, error) {
	// ユーザー情報を取得
	userData, err := repository.GetData("users", userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %v", err)
	}

	return &repository.User{
		ID:       userData["id"].(string),
		Name:     userData["name"].(string),
		Email:    userData["email"].(string),
		Icon:     userData["icon"].(string),
		IsOnline: userData["isOnline"].(bool),
	}, nil
}
