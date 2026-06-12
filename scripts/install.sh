#!/bin/sh
set -eu

program_name="yt"
prefix="${PREFIX:-$HOME/.local}"
bin_dir="${BINDIR:-}"
dry_run=0

usage() {
	cat <<'USAGE'
Install the YouTrack CLI from this checkout.

Usage:
  scripts/install.sh [--prefix DIR] [--bin-dir DIR] [--name NAME] [--dry-run]
  scripts/install.sh --help

Options:
  --prefix DIR    Installation prefix. Defaults to $PREFIX or ~/.local.
  --bin-dir DIR   Directory for the installed binary. Defaults to PREFIX/bin.
  --name NAME     Installed binary name. Defaults to yt.
  --dry-run       Print what would happen without building or installing.
  --help          Show this help.

Environment:
  PREFIX          Default prefix when --prefix is not passed.
  BINDIR          Default binary directory when --bin-dir is not passed.
  GO              Go command to use. Defaults to go.
USAGE
}

die() {
	printf '%s\n' "install.sh: $*" >&2
	exit 1
}

say() {
	printf '%s\n' "$*"
}

while [ "$#" -gt 0 ]; do
	case "$1" in
		--prefix)
			[ "$#" -ge 2 ] || die "--prefix requires a directory"
			prefix=$2
			shift 2
			;;
		--bin-dir)
			[ "$#" -ge 2 ] || die "--bin-dir requires a directory"
			bin_dir=$2
			shift 2
			;;
		--name)
			[ "$#" -ge 2 ] || die "--name requires a binary name"
			program_name=$2
			shift 2
			;;
		--dry-run)
			dry_run=1
			shift
			;;
		--help|-h)
			usage
			exit 0
			;;
		*)
			die "unknown option: $1"
			;;
	esac
done

[ -n "$program_name" ] || die "--name cannot be empty"
case "$program_name" in
	*/*) die "--name must be a file name, not a path" ;;
esac

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)

[ -f "$repo_root/go.mod" ] || die "could not find go.mod at $repo_root"
[ -d "$repo_root/cmd/yt" ] || die "could not find CLI package at $repo_root/cmd/yt"

if [ -z "$bin_dir" ]; then
	bin_dir="$prefix/bin"
fi

go_cmd=${GO:-go}
target="$bin_dir/$program_name"

say "YouTrack CLI installer"
say "  source: $repo_root"
say "  target: $target"
say "  go:     $go_cmd"

if [ "$dry_run" -eq 1 ]; then
	say "dry run: no files changed"
	exit 0
fi

command -v "$go_cmd" >/dev/null 2>&1 || die "Go is required; install Go or set GO=/path/to/go"

tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/youtrack-install.XXXXXX")
cleanup() {
	rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

say "building $program_name"
(cd "$repo_root" && "$go_cmd" build -trimpath -ldflags "-s -w" -o "$tmp_dir/$program_name" ./cmd/yt)

mkdir -p "$bin_dir"
install -m 0755 "$tmp_dir/$program_name" "$target"

say "installed $target"

case ":$PATH:" in
	*":$bin_dir:"*) ;;
	*) say "add $bin_dir to PATH to run $program_name from any directory" ;;
esac

"$target" --help >/dev/null
say "verified $program_name --help"
