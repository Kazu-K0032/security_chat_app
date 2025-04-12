package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"security_chat_app/internal/domain"
	"security_chat_app/internal/infrastructure/firebase"
	"security_chat_app/internal/infrastructure/repository"
	"security_chat_app/internal/interface/markup"
	"security_chat_app/internal/interface/middleware"
)

// チャット開始ハンドラ
func StartChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// セッションの検証
	session, err := middleware.ValidateSession(w, r)
	if err != nil {
		http.Error(w, "認証されていません", http.StatusUnauthorized)
		return
	}

	// セッションからユーザー情報を取得
	user, err := repository.GetUserByID(session.User.ID)
	if err != nil {
		log.Printf("ユーザー情報の取得に失敗: %v", err)
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// URLから対象ユーザーIDを取得
	targetUserID := r.URL.Path[len("/chat/"):]
	log.Printf("対象ユーザーID: %s", targetUserID)
	if targetUserID == "" {
		http.Error(w, "ユーザーIDが指定されていません", http.StatusBadRequest)
		return
	}

	// 対象ユーザーの存在確認
	targetUser, err := GetUserData(targetUserID)
	if err != nil {
		log.Printf("対象ユーザーの取得に失敗: %v", err)
		http.Error(w, "対象ユーザーが見つかりません", http.StatusNotFound)
		return
	}
	log.Printf("対象ユーザー情報: %+v", targetUser)

	// チャットを開始
	chatID, err := firebase.StartChat(user.ID, targetUserID)
	if err != nil {
		log.Printf("チャット開始に失敗: %v", err)
		http.Error(w, "チャットの開始に失敗しました", http.StatusInternalServerError)
		return
	}
	log.Printf("チャットID: %s", chatID)

	// チャットページにリダイレクト
	redirectURL := fmt.Sprintf("/chat?chat_id=%s", chatID)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// チャットページのハンドラ
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	// セッションの検証
	session, err := middleware.ValidateSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// セッションからユーザー情報を取得
	user, err := repository.GetUserByID(session.User.ID)
	if err != nil {
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// POSTリクエストの場合はメッセージ送信処理
	if r.Method == http.MethodPost {
		// フォームデータから情報を取得
		chatID := r.FormValue("chatID")
		content := r.FormValue("content")

		if chatID == "" || content == "" {
			http.Error(w, "チャットIDとメッセージ内容が必要です", http.StatusBadRequest)
			return
		}

		// メッセージを作成
		message := map[string]interface{}{
			"sender_id":   user.ID,
			"sender_name": user.Name,
			"content":     content,
			"created_at":  time.Now(),
			"is_read":     false,
		}

		// メッセージを保存
		err = firebase.AddChatMessage(chatID, message)
		if err != nil {
			http.Error(w, "メッセージの送信に失敗しました", http.StatusInternalServerError)
			return
		}

		// JSONレスポンスを返す
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": content,
			"time":    time.Now().Format("15:04"),
		})
		return
	}

	// チャット履歴を取得
	chats, err := getChatHistory(user)
	if err != nil {
		http.Error(w, "チャット一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// URLからチャットIDを取得
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		// チャットIDがない場合は、チャット一覧を表示
		data := domain.TemplateData{
			IsLoggedIn: true,
			User:       user,
			Chats:      chats,
			ChatID:     "", // 空のチャットIDを設定
		}
		markup.GenerateHTML(w, data, "layout", "header", "chat", "footer")
		return
	}

	// チャットの存在確認
	exists, err := firebase.CheckChatExists(chatID)
	if err != nil {
		http.Error(w, "チャットの確認に失敗しました", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "チャットが見つかりません", http.StatusNotFound)
		return
	}

	// チャットの参加者を取得
	participants, err := firebase.GetChatParticipants(chatID)
	if err != nil {
		http.Error(w, "チャットの参加者情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 対象ユーザーを特定
	var targetUserID string
	for _, p := range participants {
		if p != user.ID {
			targetUserID = p
			break
		}
	}

	// 対象ユーザーの存在確認
	_, err = GetUserData(targetUserID)
	if err != nil {
		http.Error(w, "対象ユーザーが見つかりません", http.StatusNotFound)
		return
	}

	// 対象ユーザーの情報を取得
	targetUser, err := GetUserData(targetUserID)
	if err != nil {
		http.Error(w, "対象ユーザーの情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// メッセージを取得
	messagesData, err := firebase.GetChatMessages(chatID)
	if err != nil {
		http.Error(w, "メッセージの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// メッセージの型変換
	var messages []domain.Message
	for _, msg := range messagesData {
		message := domain.Message{
			ID:         msg["id"].(string),
			Content:    msg["content"].(string),
			SenderID:   msg["sender_id"].(string),
			SenderName: msg["sender_name"].(string),
			CreatedAt:  msg["created_at"].(time.Time),
			IsRead:     msg["is_read"].(bool),
		}
		messages = append(messages, message)
	}

	// 現在のチャットを特定
	var currentChat *domain.Chat
	for _, chat := range chats {
		if chat.ID == chatID {
			currentChat = &chat
			break
		}
	}

	// チャットページのデータを取得
	data := domain.TemplateData{
		IsLoggedIn:  true,
		User:        user,
		Messages:    messages,
		Contacts:    []domain.Contact{{ID: targetUser.ID, Username: targetUser.Name, Icon: targetUser.Icon}},
		Chats:       chats,
		CurrentChat: currentChat,
		ChatID:      chatID,
	}
	log.Printf("テンプレートデータ: %+v", data)

	// テンプレートのレンダリング
	markup.GenerateHTML(w, data, "layout", "header", "chat", "footer")
}

// チャット履歴を取得
func getChatHistory(user *domain.User) ([]domain.Chat, error) {
	// チャット履歴を取得
	chats, err := firebase.GetAllData("chats")
	if err != nil {
		return nil, fmt.Errorf("チャット履歴の取得に失敗しました: %v", err)
	}

	var chatHistory []domain.Chat
	seenChats := make(map[string]bool) // 重複チェック用のマップ

	for _, chatData := range chats {
		// チャットIDの取得
		chatID, ok := chatData["id"].(string)
		if !ok {
			continue
		}

		// 参加者の取得
		participants, ok := chatData["participants"].([]interface{})
		if !ok {
			continue
		}

		// 現在のユーザーが参加者かチェック
		isParticipant := false
		var targetUserID string
		for _, p := range participants {
			if participantID, ok := p.(string); ok {
				if participantID == user.ID {
					isParticipant = true
				} else {
					targetUserID = participantID
				}
			}
		}

		if !isParticipant || seenChats[chatID] {
			continue
		}
		seenChats[chatID] = true

		// メッセージの取得
		messagesData, err := firebase.GetChatMessages(chatID)
		if err != nil {
			continue
		}

		// メッセージの型変換
		var messages []domain.Message
		var lastMessageTime time.Time
		for _, msg := range messagesData {
			messageTime := msg["created_at"].(time.Time)
			message := domain.Message{
				ID:         msg["id"].(string),
				Content:    msg["content"].(string),
				SenderID:   msg["sender_id"].(string),
				SenderName: msg["sender_name"].(string),
				CreatedAt:  messageTime,
				IsRead:     msg["is_read"].(bool),
			}
			messages = append(messages, message)

			// 最新のメッセージ時刻を更新
			if messageTime.After(lastMessageTime) {
				lastMessageTime = messageTime
			}
		}

		// チャット相手の情報を取得
		targetUser, err := GetUserData(targetUserID)
		if err != nil {
			continue
		}

		// チャット履歴に追加
		chatHistory = append(chatHistory, domain.Chat{
			ID: chatID,
			Contact: domain.Contact{
				ID:       targetUser.ID,
				Username: targetUser.Name,
				Icon:     targetUser.Icon,
				LastSeen: time.Now(),
			},
			Messages:  messages,
			UpdatedAt: lastMessageTime,
		})
	}

	// 更新時刻でソート（新しい順）
	sort.Slice(chatHistory, func(i, j int) bool {
		return chatHistory[i].UpdatedAt.After(chatHistory[j].UpdatedAt)
	})

	return chatHistory, nil
}

// ChatControllerImpl チャットコントローラーの実装
type ChatControllerImpl struct {
	chatUsecase domain.ChatUsecase
}

// NewChatController チャットコントローラーを作成する
func NewChatController(chatUsecase domain.ChatUsecase) *ChatControllerImpl {
	return &ChatControllerImpl{
		chatUsecase: chatUsecase,
	}
}

// Create チャットを作成する
func (c *ChatControllerImpl) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := r.FormValue("user")
	message := r.FormValue("message")

	if err := c.chatUsecase.CreateChat(user, message); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// メッセージ送信ハンドラ
func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// セッションの検証
	session, err := middleware.ValidateSession(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// セッションからユーザー情報を取得
	user, err := repository.GetUserByID(session.User.ID)
	if err != nil {
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// フォームデータから情報を取得
	chatID := r.FormValue("chatID")
	content := r.FormValue("content")

	if chatID == "" || content == "" {
		http.Error(w, "チャットIDとメッセージ内容が必要です", http.StatusBadRequest)
		return
	}

	// メッセージを作成
	message := map[string]interface{}{
		"sender_id":   user.ID,
		"sender_name": user.Name,
		"content":     content,
		"created_at":  time.Now(),
		"is_read":     false,
	}

	// メッセージを保存
	err = firebase.AddChatMessage(chatID, message)
	if err != nil {
		http.Error(w, "メッセージの送信に失敗しました", http.StatusInternalServerError)
		return
	}

	// チャットページにリダイレクト
	http.Redirect(w, r, fmt.Sprintf("/chat?chat_id=%s", chatID), http.StatusSeeOther)
}

// メッセージIDを生成する
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
