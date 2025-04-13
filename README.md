# security_chat_app

Goを使用した基本的なチャットアプリになります。

## アプリURL
https://tmp.kuji-server.com/search

## スクリーンショット(2025.04.13時点)
<img src="https://github.com/user-attachments/assets/95b74b85-780b-4213-9b30-c1aa09b91593" width="400">
<img src="https://github.com/user-attachments/assets/5926a535-a794-4e32-b324-a1c1cbc07857" width="400">
<img src="https://github.com/user-attachments/assets/3d0c6fc1-b7d6-49cf-9dc6-c70a10aaa6dd" width="400">
<img src="https://github.com/user-attachments/assets/352cb22a-1f25-412c-8630-091e8dcc0c21" width="400">

## 実装済み機能
- 認証機能（登録/ログイン/ログアウト）
- プロフィール（ユーザー名/画像/パスワードなどの変更）
- 検索機能（登録済みユーザーのフィルタリング）
- チャット機能（他ユーザーと連絡）

## 使用技術
- Go
- Firebase(firestore, storage)
- HTML/CSS(SCSS)/JavaScript
- stylelint, prettier, gulp

## ローカルへの導入手順

1. **プロジェクトのclone**
   
    ```bash
    git clone git@github.com:Kazu-K0032/security_chat_app.git
    ```

2. **FirebaseからFirebase Admin SDKの認証ファイルを取り込む**

    * [Firebase](https://console.firebase.google.com/u/1/?hl=ja)からプロジェクトを作成
    * 作成したプロジェクトにアクセスし、「プロジェクトの設定」⇒「サービスアカウント」⇒「新しい鍵を生成」
      <img src="https://github.com/user-attachments/assets/c0820422-87d5-4490-80aa-cfe02c564456" width="400">
      <img src="https://github.com/user-attachments/assets/de34f37d-d44b-40a4-8e6f-44ec215f11c9" height="300">
    * ダウンロードしたファイル名を「serviceAccountKey.json」に変更し、クローンしたプロジェクトの`internal/config/`に配置してください

3. **Firestoreの設定**

    * Firebaseプロジェクト⇒「Firestore Database」から、データベースを作成
      
      <img src="https://github.com/user-attachments/assets/85cb2709-e414-4e69-84d1-abbeda4f10f7" height="300">
      <img src="https://github.com/user-attachments/assets/8bcdd85f-75ac-4771-88ca-8d73fa04e35a" height="300">
  
    * Firestoreの「ルール」から以下のルールに変更
      
      ```js
      rules_version = '2';
      
      service cloud.firestore {
        match /databases/{database}/documents {
          match /{document=**} {
            allow read, write: if false;
          }
        }
      }
      ```

4. **Storageの設定**

    <img src="https://github.com/user-attachments/assets/6e107ac9-c117-45a3-8307-3c3b494b9b57" height="300">

5. **設定ファイルの修正（`config.ini`）**

    * ポートの設定をしています。ご自身の環境に合わせて、随時修正してください。

    ```txt
    [web]
    port = 8050
    logfile = debug.log
    static = app/views
    
    [firebase]
    serviceKeyPath = internal/config/serviceAccountKey.json // serviceAccountKey.jsonの相対パス
    projectId = // <プロジェクトの設定> -> <全般> -> <プロジェクトID> の値
    storageBucket = // <Storage> -> <バケット>
    ```

    ### 参考(projectId)
    <img src="https://github.com/user-attachments/assets/ee00624a-0634-4f30-8ccf-65b1eedb23d7" height="300">
    
    ### 参考(storageBucket)
    <img src="https://github.com/user-attachments/assets/7918ecc4-1617-478d-82db-65b183fbbf33" height="300">


6. **モジュール初期化および依存解決**

    * 事前に、Go及びNode.jsをダウンロードしてください。
    * バージョンは、Goは最低1.21以上, Node.jsはv16.0.0以上を目安に更新してください。
    
        ```:bash
        cd security_chat_app/
        # Go モジュールの初期化
        go mod tidy
        # Node.jsの依存解決
        npm install
        ```

7. **サーバーの起動**

     ```bash
     go run cmd/app/main.go
     ```
     * 実行後、`debug.log`が生成されます。
     * デフォルトだと、`localhost:8050`にアクセスできるようになります。

9. **追加認証**
   
    * `localhost:8050/profile`にアクセスしたとき、コマンドラインまたは `debug.log`にて「投稿」「返信」「いいね」に対する認証のリンクが表示されます。リンクにアクセスし、認証を行ってください。
        <img src="https://github.com/user-attachments/assets/1454b620-f78e-4fe7-bb39-04b5810aa576" height="300">
