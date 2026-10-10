# Twitter OAuth 2.0 Login for Scraping Pool

## Context

Add Twitter OAuth 2.0 with PKCE to capture credentials for multiple Twitter accounts used in scraping rotation. Users initiate OAuth from admin panel, authorize via Twitter, callback saves tokens to PostgreSQL. Multiple accounts per installation, not tied to admin users.

## Approach

### 1. Migration: twitter_accounts table

Create `migrations/20261010000000-twitter_accounts.sql`:

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS twitter_accounts (
    id SERIAL PRIMARY KEY,
    twitter_user_id TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    display_name TEXT,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_twitter_accounts_username ON twitter_accounts(username);

-- +migrate Down
DROP TABLE IF EXISTS twitter_accounts;
```

Tokens stored encrypted via `cryptoext.Encrypt` before insert (same pattern as `SetConfig`).

### 2. Domain types in admin/admin.go

Add after `ProvisionedResource` struct:

```go
// TwitterAccount represents a Twitter account for scraping.
type TwitterAccount struct {
    ID             int
    TwitterUserID  string
    Username       string
    DisplayName    string
    AccessToken    string    // decrypted
    RefreshToken   string    // decrypted
    TokenExpiresAt *time.Time
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

Add to `IStore` interface:

```go
// Twitter Accounts
CreateTwitterAccount(ctx context.Context, account *TwitterAccount) error
GetTwitterAccount(ctx context.Context, id int) (*TwitterAccount, error)
GetTwitterAccountByTwitterID(ctx context.Context, twitterUserID string) (*TwitterAccount, error)
ListTwitterAccounts(ctx context.Context) ([]TwitterAccount, error)
UpdateTwitterAccountTokens(ctx context.Context, id int, accessToken, refreshToken string, expiresAt *time.Time) error
DeleteTwitterAccount(ctx context.Context, id int) error
```

### 3. PostgreSQL implementation in admin/postgres/postgres.go

Add methods implementing the interface. Encrypt `access_token` and `refresh_token` on write using `cryptoext.Encrypt(token, s.encryptionKey)`. Decrypt on read. Follow existing `SetConfig`/`GetConfig` encryption pattern.

`CreateTwitterAccount`: INSERT with ON CONFLICT (twitter_user_id) DO UPDATE to handle re-auth of same account.

### 4. OAuth state storage

Add to `IStore`:

```go
// OAuth State (temporary, for PKCE flow)
SetOAuthState(ctx context.Context, state, codeVerifier string, expiresAt time.Time) error
GetOAuthState(ctx context.Context, state string) (codeVerifier string, err error)
DeleteOAuthState(ctx context.Context, state string) error
```

Simple `oauth_states` table in same migration:

```sql
CREATE TABLE IF NOT EXISTS oauth_states (
    state TEXT PRIMARY KEY,
    code_verifier TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
```

Implementation cleans expired states on each `GetOAuthState` call.

### 5. Handlers in admin/handlers_twitter.go

New file with:

**TwitterAccountsPageHandler**: GET `/admin/twitter` - list accounts, show "Connect Twitter Account" button. Render `twitter_accounts.html`.

**TwitterAuthHandler**: GET `/admin/twitter/auth` - initiate OAuth:
1. Generate `state` (32 random bytes, hex)
2. Generate `code_verifier` (43-128 chars, base64url)
3. Compute `code_challenge` = base64url(sha256(code_verifier))
4. Store state + code_verifier via `SetOAuthState` (5 min TTL)
5. Redirect to `https://twitter.com/i/oauth2/authorize` with params:
   - `response_type=code`
   - `client_id` from env `TWITTER_CLIENT_ID`
   - `redirect_uri` from env `TWITTER_REDIRECT_URI` (e.g. `https://yoursite.com/admin/twitter/callback`)
   - `scope=tweet.read users.read offline.access`
   - `state`
   - `code_challenge`
   - `code_challenge_method=S256`

**TwitterCallbackHandler**: GET `/admin/twitter/callback` - handle callback:
1. Validate `state` param via `GetOAuthState`, get `code_verifier`
2. Delete state immediately
3. Exchange `code` for tokens: POST `https://api.twitter.com/2/oauth2/token` with:
   - `grant_type=authorization_code`
   - `code`
   - `redirect_uri`
   - `code_verifier`
   - Basic auth: `client_id:client_secret` (client_secret from env `TWITTER_CLIENT_SECRET`)
4. Parse response: `access_token`, `refresh_token`, `expires_in`
5. Fetch user info: GET `https://api.twitter.com/2/users/me` with Bearer token
6. Call `CreateTwitterAccount` with user data + tokens
7. Redirect to `/admin/twitter?success=Account+connected`

**DeleteTwitterAccountHandler**: POST `/admin/twitter/{id}/delete` - remove account.

### 6. Routes in admin/routes.go

Add inside authenticated group (after existing routes):

```go
r.Get("/twitter", TwitterAccountsPageHandler(appState))
r.Get("/twitter/auth", TwitterAuthHandler(appState))
r.Get("/twitter/callback", TwitterCallbackHandler(appState))
r.Post("/twitter/{id}/delete", DeleteTwitterAccountHandler(appState))
```

### 7. Template admin/templates/twitter_accounts.html

Follow `api_keys.html` pattern:
- Table listing accounts: username, display_name, connected date, expires_at, delete button
- "Connect Twitter Account" button linking to `/admin/twitter/auth`
- Success/error message display from query params

### 8. Environment variables

Document in `saas/constants.go` or usage:
- `TWITTER_CLIENT_ID` - from Twitter Developer Portal
- `TWITTER_CLIENT_SECRET` - from Twitter Developer Portal  
- `TWITTER_REDIRECT_URI` - full callback URL

No code reads these at startup; handlers read via `os.Getenv` on demand (fail if missing when auth initiated).

## Critical files & anchors

| Path | Anchor | Reason |
|------|--------|--------|
| `admin/postgres/postgres.go` | `func (s *store) SetConfig` | Encryption pattern to copy for token storage |
| `admin/handlers_settings.go` | `SettingsPageHandler` | Handler + template render pattern |
| `admin/templates/api_keys.html` | entire file | Table + action button template pattern |
| `admin/routes.go` | `r.Get("/api-keys"` | Route registration pattern in auth group |

## Verification

1. **Migration**: Run `make migrate` or app startup; confirm `twitter_accounts` and `oauth_states` tables exist.

2. **OAuth flow**:
   - Set env: `TWITTER_CLIENT_ID`, `TWITTER_CLIENT_SECRET`, `TWITTER_REDIRECT_URI=http://localhost:8080/admin/twitter/callback`
   - Register callback URL in Twitter Developer Portal
   - Start server, login to admin
   - Navigate to `/admin/twitter`, click "Connect Twitter Account"
   - Should redirect to Twitter auth page
   - After Twitter login + authorize, callback redirects to `/admin/twitter?success=...`
   - Account appears in list with username

3. **Token storage**: Query `SELECT twitter_user_id, username, access_token FROM twitter_accounts` - access_token should be encrypted (gibberish), not plaintext.

4. **Delete**: Click delete on account, confirm removed from list and DB.

## Assumptions & contingencies

- **Twitter Developer account exists** with OAuth 2.0 app configured. If not, user must create at developer.twitter.com before this feature works.
- **Callback URL must be HTTPS in production**. Twitter requires it. For local dev, `http://localhost:*` is allowed.
- If `TWITTER_CLIENT_ID` missing when auth initiated: return error page "Twitter OAuth not configured", do not crash.
