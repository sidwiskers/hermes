#!/usr/bin/env bash
set -euo pipefail

for module in tgbotapi telebot telebotv4 gotgbot gotelegrambot telego gotg; do
	(
		cd "benchmarks/competitors/$module"
		go test ./...
	)
done
