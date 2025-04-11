package domain

import (
	"time"
)

// メッセージの種類
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeAudio MessageType = "audio"
	MessageTypeFile  MessageType = "file"
)

// エラーコードの定数
const (
	ChatErrorCodeInvalidInput  = "INVALID_INPUT"
	ChatErrorCodeNotFound      = "NOT_FOUND"
	ChatErrorCodeUnauthorized  = "UNAUTHORIZED"
	ChatErrorCodeInternalError = "INTERNAL_ERROR"
)

// チャットの構造体
type Chat struct {
	ID        string    // チャットのID
	IsGroup   bool      // グループチャットかどうか
	Messages  []Message // メッセージのリスト
	CreatedAt time.Time // チャットの作成日時
	UpdatedAt time.Time // チャットの更新日時
}

// チャット参加者の構造体
type ChatParticipant struct {
	ID       string    // チャット参加者のID
	ChatID   string    // チャットのID
	UserID   string    // ユーザーのID
	Role     string    // チャット参加者のロール
	JoinedAt time.Time // チャット参加者の参加日時
}

// メッセージの構造体
type Message struct {
	ID         string      // メッセージのID
	ChatID     string      // チャットのID
	SenderID   string      // 送信者のID
	SenderName string      // 送信者の名前
	Type       MessageType // メッセージの種類
	Content    string      // メッセージの内容
	MediaURL   string      // メッセージのメディアのURL
	CreatedAt  time.Time   // メッセージの作成日時
	IsRead     bool        // メッセージが読まれたかどうか
	ReadBy     []string    // メッセージを読んだユーザーのID
	ReplyTo    string      // メッセージの返信先のID
}

// 連絡先を交換したユーザーの構造体
type Contact struct {
	ID       string    // 連絡先のID
	Username string    // 連絡先のユーザー名
	IconURL  string    // 連絡先のアイコンのURL
	LastSeen time.Time // 連絡先の最終接続日時
}

// ビジネスロジックの為のチャットのユースケース
type ChatUsecase interface {
	CreateChat(user, message string) error
	GetChatHistory(user *User) ([]Chat, error)
	GetContacts(user *User) ([]Contact, error)
}

// データアクセスを定義
type ChatRepository interface {
	AddChat(user, message string) error
	GetChats(userID string) ([]Chat, error)
	GetMessages(chatID string) ([]Message, error)
}

// チャットのコントローラー
type ChatController interface {
	HandleCreateChat(user, message string) error
	HandleGetChatHistory(user *User) ([]Chat, error)
	HandleGetContacts(user *User) ([]Contact, error)
}

// チャット関連のエラー
type ChatError struct {
	Code    string // エラーコード
	Message string // エラーメッセージ
}

// エラーの文字列を返す
func (e *ChatError) Error() string {
	return e.Message
}
