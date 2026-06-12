#!/bin/sh
set -eu

default_repo_url="https://github.com/dutch-casa/youtrack"
default_archive_url="https://github.com/dutch-casa/youtrack/archive/refs/heads/main.tar.gz"

program_name="yt"
alias_name=""
prefix="${PREFIX:-$HOME/.local}"
bin_dir="${BINDIR:-}"
repo_url="${YOUTRACK_REPO_URL:-$default_repo_url}"
archive_url="${YOUTRACK_ARCHIVE_URL:-$default_archive_url}"
dry_run=0

usage() {
	cat <<'USAGE'
Install the YouTrack CLI.

Usage:
  scripts/install.sh [--prefix DIR] [--bin-dir DIR] [--name NAME] [--repo URL] [--dry-run]
  scripts/install.sh --help

Options:
  --prefix DIR    Installation prefix. Defaults to $PREFIX or ~/.local.
  --bin-dir DIR   Directory for the installed binary. Defaults to PREFIX/bin.
  --name NAME     Installed binary name. Defaults to yt. Installing yt also registers youtrack, and installing youtrack also registers yt.
  --repo URL      Git repository to clone when not run from a checkout.
  --dry-run       Print what would happen without building or installing.
  --help          Show this help.

Environment:
  PREFIX          Default prefix when --prefix is not passed.
  BINDIR          Default binary directory when --bin-dir is not passed.
  GO              Go command to use. Defaults to go.
  YOUTRACK_REPO_URL
                  Default repository URL when --repo is not passed.
  YOUTRACK_ARCHIVE_URL
                  Source archive URL used when git is unavailable.
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
		--repo)
			[ "$#" -ge 2 ] || die "--repo requires a repository URL"
			repo_url=$2
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
case "$program_name" in
	yt) alias_name="youtrack" ;;
	youtrack) alias_name="yt" ;;
esac

clone_dir=""
build_dir=""

cleanup() {
	if [ -n "$build_dir" ]; then
		rm -rf "$build_dir"
	fi
	if [ -n "$clone_dir" ]; then
		rm -rf "$clone_dir"
	fi
}
trap cleanup EXIT INT TERM

checkout_root() {
	candidate=$1
	if [ -f "$candidate/go.mod" ] && [ -d "$candidate/cmd/yt" ]; then
		CDPATH= cd -- "$candidate" && pwd
		return 0
	fi
	return 1
}

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" 2>/dev/null && pwd || pwd)
repo_root=$(checkout_root "$script_dir/.." || checkout_root "." || true)

if [ -z "$bin_dir" ]; then
	bin_dir="$prefix/bin"
fi

go_cmd=${GO:-go}
target="$bin_dir/$program_name"
alias_target=""
if [ -n "$alias_name" ]; then
	alias_target="$bin_dir/$alias_name"
fi

if [ -z "$repo_root" ] && command -v git >/dev/null 2>&1; then
	clone_dir=$(mktemp -d "${TMPDIR:-/tmp}/youtrack-src.XXXXXX")
	say "cloning $repo_url"
	git clone --depth 1 "$repo_url" "$clone_dir/youtrack"
	repo_root=$(checkout_root "$clone_dir/youtrack") || die "cloned repository does not look like YouTrack CLI"
fi

if [ -z "$repo_root" ]; then
	if [ "$repo_url" != "$default_repo_url" ] && [ "$archive_url" = "$default_archive_url" ]; then
		die "git is required for custom --repo installs; set YOUTRACK_ARCHIVE_URL to a source archive for gitless installs"
	fi
	command -v curl >/dev/null 2>&1 || die "curl is required when git is unavailable"
	command -v tar >/dev/null 2>&1 || die "tar is required when git is unavailable"
	clone_dir=$(mktemp -d "${TMPDIR:-/tmp}/youtrack-src.XXXXXX")
	archive_file="$clone_dir/source.tar.gz"
	say "downloading $archive_url"
	curl -fsSL "$archive_url" -o "$archive_file"
	tar -xzf "$archive_file" -C "$clone_dir"
	for candidate in "$clone_dir"/*; do
		if repo_root=$(checkout_root "$candidate" 2>/dev/null); then
			break
		fi
	done
	[ -n "$repo_root" ] || die "downloaded archive does not look like YouTrack CLI"
fi

say "YouTrack CLI installer"
say "  source: $repo_root"
say "  target: $target"
if [ -n "$alias_target" ]; then
	say "  alias:  $alias_target"
fi
say "  go:     $go_cmd"

if [ "$dry_run" -eq 1 ]; then
	say "dry run: no files changed"
	exit 0
fi

command -v "$go_cmd" >/dev/null 2>&1 || die "Go is required; install Go or set GO=/path/to/go"

build_dir=$(mktemp -d "${TMPDIR:-/tmp}/youtrack-build.XXXXXX")

say "building $program_name"
(cd "$repo_root" && "$go_cmd" build -trimpath -ldflags "-s -w" -o "$build_dir/$program_name" ./cmd/yt)

mkdir -p "$bin_dir"
install -m 0755 "$build_dir/$program_name" "$target"

say "installed $target"
if [ -n "$alias_target" ]; then
	ln -sf "$program_name" "$alias_target"
	say "registered $alias_target"
fi

case ":$PATH:" in
	*":$bin_dir:"*) ;;
	*) say "add $bin_dir to PATH to run $program_name from any directory" ;;
esac

"$target" --help >/dev/null
say "verified $program_name --help"
if [ -n "$alias_target" ]; then
	"$alias_target" --help >/dev/null
	say "verified $alias_name --help"
fi
