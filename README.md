# pat
A cat that lives on the internet.

## Admin access

Admin pages are available at `/admin/` and are disabled by default. To enable them, start the server with both `-secret` and `-admin-password` flags.

### Generating a password hash

The `-admin-password` flag accepts a bcrypt hash, not the plaintext password. Generate one with:

```sh
go run _tools/pwd.go -password <PASSWORD>
```


### Starting the server

```sh
./pat \
  -secret="$(openssl rand -hex 32)" \
  -admin-password='$2a$10$...'
```

- `-secret` — an arbitrary random string used to sign session cookies. Changing it invalidates all active sessions.
- `-admin-password` — the bcrypt hash from above. Changing it does **not** invalidate existing sessions (use `-secret` rotation for that).

Omitting either flag disables admin pages entirely.
