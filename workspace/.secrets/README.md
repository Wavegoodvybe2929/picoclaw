# Secrets

This folder is for local-only secrets.

## Gmail
Place your OAuth token JSON here:
- `.secrets/gmail_token.json`

The `./bin/check_gmail_unread` script expects a token created with Gmail readonly scope.

Install deps (if needed):

```bash
python3 -m pip install --upgrade google-api-python-client google-auth google-auth-oauthlib
```
