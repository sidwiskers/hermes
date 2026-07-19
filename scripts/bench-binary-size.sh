#!/usr/bin/env bash
set -euo pipefail

output="$(mktemp -d)"
trap 'rm -rf "$output"' EXIT

build_size() {
	local name="$1"
	local directory="$2"
	local package="$3"
	(
		cd "$directory"
		go build -trimpath -ldflags='-s -w' -o "$output/$name" "$package"
	)
	printf '%s\t%s\n' "$name" "$(wc -c < "$output/$name")"
}

printf 'library\tstripped_bytes\n'
build_size hermes . ./benchmarks/binary/hermes
build_size tgbotapi benchmarks/competitors/tgbotapi ./cmd/minimal
build_size telebot benchmarks/competitors/telebot ./cmd/minimal
build_size telebotv4 benchmarks/competitors/telebotv4 ./cmd/minimal
build_size gotgbot benchmarks/competitors/gotgbot ./cmd/minimal
build_size gotelegrambot benchmarks/competitors/gotelegrambot ./cmd/minimal
build_size telego benchmarks/competitors/telego ./cmd/minimal
build_size gotg benchmarks/competitors/gotg ./cmd/minimal
