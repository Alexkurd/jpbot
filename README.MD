## Telegram bot-helper for groups

### Features
- Welcome message + human button
- Chat permissions for guests/members
- Triggers for message replies with heplful links
- Ranking(in progress)

### Quickstart
- Rename *.yaml.dist => *.yaml
- Fill bot_token and hostport(if you'll use webhooks) in config
- For your boot switch of Privacy

#### Webhooks
For webhooks you need SSL-cert stored in files jpbot.pem and jpbot.key. It could be self-signed and you can issue for IP.
Telegram servers doesn't support IPv6 yet(2024.02.01).

Run app with go
```
go get jpbot
go run jpbot
```