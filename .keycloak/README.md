# The development realm

`ensoria-realm.json` is imported into Keycloak on its first start by the
`ensoria-ensoria-template-keycloak` service in `compose.yaml`. It exists so that
the application can be developed against **real tokens from a real issuer**,
which `AUTH_MODE=hs256` cannot show.

It is a disposable development realm and is not a starting point for a
production one: every secret in it is public, the user's password is written
down in it, and its lifetimes are chosen for convenience.

Realm JSON has no comment syntax — Keycloak refuses a file containing a field it
does not recognise — so what would have been comments is here instead.

## What is in it

| Thing | Value | Why |
|---|---|---|
| Realm | `ensoria` | Issuer becomes `http://localhost:8081/realms/ensoria` |
| Client | `ensoria-frontend` | Public client. Standard flow for a browser on `http://localhost:3000`, direct access grants so `curl` can get a token in one request. |
| User | `alice` / `alice` | Holds every permission the template's endpoints declare |
| Realm roles | `users:read`, `users:write`, `orders:read`, `admin:jobs:*`, `admin:tasks:*` | The scopes the template's `Security` declarations ask for |

## The two mappers, and why they are needed

Keycloak's defaults do not produce a token this application accepts. Two
mappers close the gap, and both are worth understanding before pointing the
application at a realm somebody else configured.

**`realm-roles-as-scope`.** The application reads permissions from the
space-separated `scope` claim (RFC 8693), which is where an OAuth 2 access token
carries them. Keycloak puts realm roles in `realm_access.roles` instead, and
nothing reads that. The mapper writes the same roles into `scope`, so an
endpoint declaring `Scopes: []string{"users:read"}` is satisfied by a realm role
of that name.

**`ensoria-audience`.** `AUTH_AUDIENCE` is checked against the token's `aud`
when it is set. Keycloak's default `aud` is `account`, so the check would refuse
every token. The mapper adds `ensoria` to the audience. Leaving `AUTH_AUDIENCE`
unset skips the check instead — but then any token this issuer signs is
accepted, including one minted for a different application in the same realm.

## Changing it

Keycloak imports this file **once**, into its own storage. Editing it afterwards
changes nothing. To import again:

```bash
docker compose --profile keycloak down -v   # -v drops the realm storage
docker compose --profile keycloak up -d
```

Changes made through the admin console (http://localhost:8081, `admin` /
`admin`) live in that storage and are lost the same way. Export them back into
this file if they are worth keeping.
