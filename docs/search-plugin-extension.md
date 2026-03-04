# Mattermost 標準検索をプラグインで拡張するガイド

Avapmost が実装している「プラグインで標準検索を置き換える」仕組みを解説します。
独自の検索バックエンド（Elasticsearch、ベクトル検索、外部 API など）を Mattermost に組み込みたい開発者向けのリファレンスです。

---

## 概要

Mattermost のウェブアプリは検索実行時に `Client4.searchPostsWithParams()` を呼びます。
Avapmost ではこの呼び出しをフックし、まずプラグインの API エンドポイントへリクエストを送信します。
プラグインが応答しない場合は標準の DB 検索にフォールバックするため、既存機能との後方互換性が保たれます。

```
ユーザーが検索ワード入力
        ↓
searchPostsWithParams() (Redux action)
        ↓
[1] プラグイン API へ POST  ← ここで独自検索エンジンを呼べる
        ↓ (失敗 / プラグイン未インストール時)
[2] 標準 DB 検索 (POST /api/v4/posts/search)
        ↓
既存の検索結果 UI で表示
```

---

## 変更ファイル一覧

| ファイル | 役割 |
|---|---|
| `server/public/model/config.go` | `ServiceSettings` に設定フィールド追加 |
| `server/config/client.go` | クライアント設定への公開 |
| `webapp/platform/types/src/config.ts` | フロントエンド型定義 |
| `webapp/channels/src/packages/mattermost-redux/src/actions/search.ts` | 検索アクションのフック |

---

## Step 1: サーバー設定フィールドを追加する

プラグインのエンドポイント URL を管理者が変更できるよう、`ServiceSettings` に文字列フィールドを追加します。

**`server/public/model/config.go`**

```go
type ServiceSettings struct {
    // ... 既存フィールド ...

    // MyPlugin: 検索プラグインのエンドポイント（空 = デフォルトを使用）
    MySearchEndpoint *string `access:"write_restrictable"`
}
```

`SetDefaults()` 内にデフォルト値を設定します：

```go
if s.MySearchEndpoint == nil {
    s.MySearchEndpoint = NewPointer("/plugins/com.example.my-search/api/v1/search")
}
```

---

## Step 2: フロントエンドへ公開する

`/api/v4/config/client` を通じてウェブアプリへ設定値を渡します。

**`server/config/client.go`** の `GenerateClientConfig()` 内：

```go
props["MySearchEndpoint"] = *c.ServiceSettings.MySearchEndpoint
```

> **重要**: `GenerateClientConfig()` に追加することで認証済みユーザー全員に公開されます。
> 機密情報（APIキーなど）はここに入れないでください。
> 管理者のみに公開したい場合は `GenerateLimitedClientConfig()` を使わず、
> 別の API エンドポイントで提供する設計にしてください。

---

## Step 3: フロントエンドの型定義を追加する

**`webapp/platform/types/src/config.ts`**

```typescript
export type ClientConfig = {
    // ... 既存フィールド ...
    MySearchEndpoint: string;
};
```

---

## Step 4: 検索アクションをフックする

**`webapp/channels/src/packages/mattermost-redux/src/actions/search.ts`**

### エンドポイント取得ヘルパー

```typescript
import type {GlobalState} from '@mattermost/types/store';

const MY_SEARCH_DEFAULT_ENDPOINT = '/plugins/com.example.my-search/api/v1/search';

function getSearchEndpoint(state: GlobalState): string {
    return state.entities.general.config.MySearchEndpoint || MY_SEARCH_DEFAULT_ENDPOINT;
}
```

### プラグイン呼び出し関数

```typescript
async function searchViaPlugin(
    endpoint: string,
    params: SearchParameter,
): Promise<PostSearchResults | null> {
    try {
        const res = await fetch(endpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Requested-With': 'XMLHttpRequest',  // CSRF 対策必須
            },
            credentials: 'include',                     // セッション Cookie を送信
            body: JSON.stringify({
                query: params.terms,
                page: params.page ?? 0,
                per_page: params.per_page ?? 20,
            }),
        });
        if (!res.ok) {
            return null;  // null を返すとフォールバックへ
        }
        const data = await res.json() as PostSearchResults;
        if (!Array.isArray(data.order)) {
            return null;  // レスポンス形式が不正な場合もフォールバック
        }
        return data;
    } catch {
        return null;      // ネットワークエラー等でもフォールバック
    }
}
```

### `searchPostsWithParams` にフックを組み込む

```typescript
export function searchPostsWithParams(
    teamId: string,
    params: SearchParameter,
): ActionFuncAsync<PostSearchResults> {
    return async (dispatch, getState) => {
        // ... (dispatch REQUEST action) ...

        let posts;
        try {
            // プラグイン優先、失敗時は標準検索へフォールバック
            const pluginResult = await searchViaPlugin(getSearchEndpoint(getState()), params);
            posts = pluginResult ?? await Client4.searchPostsWithParams(teamId, params);
        } catch (error) {
            // ...
        }

        // ... (dispatch RECEIVED actions) ...
    };
}
```

---

## Step 5: プラグイン側 API の実装

### リクエスト形式

```json
POST /plugins/com.example.my-search/api/v1/search
Content-Type: application/json
Cookie: MMAUTHTOKEN=...

{
    "query": "検索キーワード",
    "page": 0,
    "per_page": 20
}
```

### レスポンス形式 (`PostSearchResults`)

既存の検索 UI と互換を保つため、`/api/v4/posts/search` と同じ形式で返す必要があります：

```json
{
    "order": ["post_id_1", "post_id_2"],
    "posts": {
        "post_id_1": {
            "id": "post_id_1",
            "channel_id": "channel_id",
            "user_id": "user_id",
            "message": "投稿本文",
            "create_at": 1700000000000,
            ...
        }
    },
    "matches": {
        "post_id_1": ["マッチしたテキスト"]
    },
    "first_inaccessible_post_time": 0
}
```

| フィールド | 型 | 説明 |
|---|---|---|
| `order` | `string[]` | 表示順の投稿 ID リスト |
| `posts` | `Record<string, Post>` | 投稿 ID → Post オブジェクトのマップ |
| `matches` | `Record<string, string[]>` | ハイライト用マッチテキスト（任意） |
| `first_inaccessible_post_time` | `number` | 有料機能用・通常は `0` |

### Go プラグイン実装例

```go
func (p *Plugin) searchHandler(w http.ResponseWriter, r *http.Request) {
    // 認証確認
    userID := r.Header.Get("Mattermost-User-Id")
    if userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var req struct {
        Query   string `json:"query"`
        Page    int    `json:"page"`
        PerPage int    `json:"per_page"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    // 独自検索ロジック（Elasticsearch, ベクトル DB など）
    results, err := p.mySearchEngine.Search(r.Context(), req.Query, req.Page, req.PerPage)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // model.PostSearchResults 形式で返す
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
}
```

---

## 設定方法

### config.json で設定

```json
{
    "ServiceSettings": {
        "MySearchEndpoint": "/plugins/com.example.my-search/api/v1/search"
    }
}
```

### 環境変数で設定

```bash
MM_SERVICESETTINGS_MYSEARCHENDPOINT="/plugins/com.example.my-search/api/v1/search"
```

フィールド名は `MM_SERVICESETTINGS_` プレフィックス + フィールド名の大文字変換です。

---

## 注意事項

### セキュリティ
- プラグイン API ハンドラは必ず `Mattermost-User-Id` ヘッダーで認証を確認すること
  - Mattermost のプラグインミドルウェアが認証済みリクエストにこのヘッダーを付与します
- 取得した投稿がリクエストユーザーに閲覧権限があるか確認すること
  - `p.API.HasPermissionToChannel(userID, channelID, model.PermissionReadChannel)` を使用

### ライセンス
- `webapp/` 配下の変更は Apache License v2.0 のスコープ
- `server/` 配下の変更は GNU AGPL v3.0 のスコープ（ソース公開義務あり）
- `server/enterprise/` は変更禁止（Mattermost Source Available License）

### フォールバック
- プラグインが `null` を返す（HTTP エラー、タイムアウト、形式不正）と自動的に標準 DB 検索へフォールバックします
- プラグイン未インストールの環境でも標準検索は動作します

---

## Avapmost の実装を参照する

本ガイドのリファレンス実装は以下のファイルです：

- **設定**: `server/public/model/config.go` (`AvapSearchEndpoint`)
- **公開**: `server/config/client.go` (`GenerateClientConfig`)
- **型定義**: `webapp/platform/types/src/config.ts`
- **フック**: `webapp/channels/src/packages/mattermost-redux/src/actions/search.ts`
- **プラグイン本体**: `plugin/avapmost-search/` (Elasticsearch + セマンティック検索)
