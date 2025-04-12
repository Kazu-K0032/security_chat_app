package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"security_chat_app/internal/config"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func InitFirebase() (*firestore.Client, error) {
	log.Printf("Firebase初期化開始")

	opt := option.WithCredentialsFile(config.Config.FirebaseServiceAccountKey)
	log.Printf("認証情報ファイル読み込み: %s", config.Config.FirebaseServiceAccountKey)

	// Firebase設定を明示的に指定
	config := &firebase.Config{
		ProjectID: "go-chat-app-cf888",
	}
	log.Printf("Firebase設定: ProjectID=%s", config.ProjectID)

	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Printf("Firebaseアプリの初期化に失敗: %v", err)
		return nil, err
	}
	log.Printf("Firebaseアプリ初期化成功")

	// タイムアウトを設定したコンテキストを使用
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("Firestoreクライアント作成に失敗: %v", err)
		return nil, err
	}
	log.Printf("Firestoreクライアント作成成功")

	return client, nil
}

// InitFirebaseClient Firebaseクライアントを初期化する
func InitFirebaseClient() (*firebase.App, error) {
	// サービスアカウントキーの読み込み
	serviceAccountKey, err := os.ReadFile(config.Config.FirebaseServiceAccountKey)
	if err != nil {
		return nil, fmt.Errorf("サービスアカウントキーの読み込みに失敗: %v", err)
	}

	// サービスアカウントキーのパース
	var serviceAccount struct {
		Type                    string `json:"type"`
		ProjectID               string `json:"project_id"`
		PrivateKeyID            string `json:"private_key_id"`
		PrivateKey              string `json:"private_key"`
		ClientEmail             string `json:"client_email"`
		ClientID                string `json:"client_id"`
		AuthURI                 string `json:"auth_uri"`
		TokenURI                string `json:"token_uri"`
		AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
		ClientX509CertURL       string `json:"client_x509_cert_url"`
	}

	if parseErr := json.Unmarshal(serviceAccountKey, &serviceAccount); parseErr != nil {
		return nil, fmt.Errorf("サービスアカウントキーのパースに失敗: %v", parseErr)
	}

	// Firebase初期化オプションの設定
	opt := option.WithCredentialsFile(config.Config.FirebaseServiceAccountKey)

	// Firebaseアプリの初期化
	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: serviceAccount.ProjectID,
	}, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase初期化に失敗: %v", err)
	}

	return app, nil
}
