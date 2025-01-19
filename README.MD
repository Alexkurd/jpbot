## Telegram bot-helper for groups

### Features
- Welcome message + check human button
- Chat permissions for guests/members
- Triggers with helpful links
- Bad words filtering
- Menu for private chats with bot
- Check API for bots(casban\lols)
- Syslog support

### Quickstart
- Rename *.yaml.dist => *.yaml
- Fill bot_token and hostport(if you'll use webhooks) in config
- For your bot switch off privacy in @BotFather

#### Docker 
Build using Dockerfile.

#### Service
Run app with go
```
go get jpbot
go run jpbot
```

#### Webhooks
For webhooks you need SSL-cert stored in files jpbot.pem and jpbot.key. It could be self-signed and you can issue for IP.
Telegram servers doesn't support IPv6 yet(2024.02.01).