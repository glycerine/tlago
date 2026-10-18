#!/usr/bin/env bash
set -euo pipefail

mode="${1:-all}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
regenerate="${REGENERATE:-0}"
java_classpath=""
apalache_command=""
apalache_jar=""
repo_root="$(cd "$root/.." && pwd)"
local_apalache_candidates=(
  "$repo_root/../apalache-mc/apalache/src/universal/bin/apalache-mc"
  "$repo_root/../../apalache-mc/apalache/src/universal/bin/apalache-mc"
)
local_apalache_jar_candidates=(
  "$repo_root/../apalache-mc/apalache/target/scala-2.13/apalache-pkg-0.62.3-SNAPSHOT-full.jar"
  "$repo_root/../../apalache-mc/apalache/target/scala-2.13/apalache-pkg-0.62.3-SNAPSHOT-full.jar"
)

xml_roots=("sany-xml" "tla-plus-bench/specs" "Examples/specifications")
air_roots=("sany-xml" "tla-plus-bench/specs" "Examples/specifications")
cleanup_paths=()
bench_manifest_dir_cache=""

cleanup() {
  if [[ "${#cleanup_paths[@]}" -gt 0 ]]; then
    rm -f "${cleanup_paths[@]}"
  fi
}
trap cleanup EXIT

die() {
  printf 'generate_goldens: %s\n' "$*" >&2
  exit 1
}

display_path() {
  local path="$1"
  if [[ "$path" == "$root/"* ]]; then
    printf '%s' "${path#"$root"/}"
  else
    printf '%s' "$path"
  fi
}

absolute_spec_path() {
  local spec="$1"
  if [[ "$spec" = /* ]]; then
    printf '%s\n' "$spec"
  else
    printf '%s/%s\n' "$root" "$spec"
  fi
}

spec_matches_roots() {
  local spec="$1"
  shift
  local rel
  rel="$(display_path "$spec")"
  local root_dir
  for root_dir in "$@"; do
    if [[ "$rel" == "$root_dir" || "$rel" == "$root_dir/"* ]]; then
      return 0
    fi
  done
  return 1
}

collect_specs() {
  if [[ -n "${SPECS:-}" ]]; then
    local spec abs
    for spec in ${SPECS}; do
      abs="$(absolute_spec_path "$spec")"
      if [[ -f "$abs" && "${abs##*.}" == "tla" ]] && spec_matches_roots "$abs" "$@"; then
        printf '%s\n' "$abs"
      fi
    done
    return
  fi

  local root_dir
  for root_dir in "$@"; do
    if [[ -d "$root/$root_dir" ]]; then
      find "$root/$root_dir" -type f -name '*.tla'
    fi
  done | sort
}

ensure_bench_manifest_dir_cache() {
  if [[ -n "$bench_manifest_dir_cache" ]]; then
    return
  fi
  command -v python3 >/dev/null 2>&1 || return
  bench_manifest_dir_cache="$(mktemp)"
  cleanup_paths+=("$bench_manifest_dir_cache")
  python3 - "$root" >"$bench_manifest_dir_cache" <<'PY'
import json
import os
import sys
from pathlib import Path

root = Path(sys.argv[1])
specs_root = root / "tla-plus-bench" / "specs"
manifest_path = root / "tla-plus-bench" / "manifest.json"

with manifest_path.open() as f:
    records = json.load(f)

by_source_dir = {}
current_paths = {}
for record in records:
    source_path = record["source_path"]
    by_source_dir.setdefault(os.path.dirname(source_path), []).append(record)
    base = os.path.basename(source_path)
    candidates = (
        specs_root / record["tier"] / base,
        specs_root / record["tier"] / str(record["spec_id"]) / base,
    )
    for candidate in candidates:
        if candidate.exists():
            current_paths[record["spec_id"]] = str(candidate.absolute())
            break

for record in records:
    current = current_paths.get(record["spec_id"])
    if not current:
        continue
    dirs = []
    seen = set()
    for candidate in by_source_dir[os.path.dirname(record["source_path"])]:
        candidate_path = current_paths.get(candidate["spec_id"])
        if not candidate_path:
            continue
        directory = str(Path(candidate_path).parent)
        if directory in seen:
            continue
        seen.add(directory)
        dirs.append(directory)
    print("\t".join([current] + sorted(dirs)))
PY
}

bench_manifest_dirs() {
  local spec="$1"
  if [[ "$spec" != "$root/tla-plus-bench/specs/"* ]]; then
    return
  fi
  ensure_bench_manifest_dir_cache
  if [[ -z "$bench_manifest_dir_cache" ]]; then
    return
  fi
  awk -F '\t' -v spec="$spec" '$1 == spec { for (i = 2; i <= NF; i++) print $i }' "$bench_manifest_dir_cache"
}

should_generate() {
  local gold="$1"
  local red="$2"
  [[ "$regenerate" == "1" || ( ! -f "$gold" && ! -f "$red" ) ]]
}

write_red() {
  local red="$1"
  local kind="$2"
  local spec="$3"
  local status="$4"
  local stdout_file="$5"
  local stderr_file="$6"
  local command_text="$7"

  {
    printf '%s oracle failed for %s\n' "$kind" "$(display_path "$spec")"
    printf 'exit_status: %s\n' "$status"
    printf 'command: %s\n' "$command_text"
    printf '\n--- stdout ---\n'
    cat "$stdout_file"
    printf '\n--- stderr ---\n'
    cat "$stderr_file"
  } >"$red"
}

xml_is_well_formed() {
  local path="$1"
  python3 - "$path" <<'PY' >/dev/null 2>&1
import sys
import xml.etree.ElementTree as ET

ET.parse(sys.argv[1])
PY
}

ensure_java_sany() {
  if [[ -n "$java_classpath" ]]; then
    return
  fi
  command -v java >/dev/null 2>&1 || die "java is required to generate .xml.gold files"

  local jar="${TLA2TOOLS_JAR:-$root/java-sany/tla2tools.jar}"
  if [[ "$jar" != *:* && ! -f "$jar" ]]; then
    die "missing tla2tools jar: $jar"
  fi
  java_classpath="$jar"

  local community_jar="$root/java-sany/CommunityModules.jar"
  if [[ -f "$community_jar" ]]; then
    java_classpath="$java_classpath:$community_jar"
  fi
  local tlaps_community_jar="$root/tlaps-stdlib/CommunityModules.jar"
  if [[ -f "$tlaps_community_jar" ]]; then
    java_classpath="$java_classpath:$tlaps_community_jar"
  fi
}

java_include_args() {
	local spec="$1"
	local source_dir parent_dir parent_name
	source_dir="$(dirname "$spec")"
	parent_dir="$(dirname "$source_dir")"
	parent_name="$(basename "$parent_dir")"

	printf '%s\n' "-I" "$root/java-sany/StandardModules"
	printf '%s\n' "-I" "$root/tlaplus-standard-modules"
	printf '%s\n' "-I" "$root/CommunityModules/modules"
	printf '%s\n' "-I" "$root/CommunityModules/tests"
	printf '%s\n' "-I" "$root/tlaps-stdlib"
	printf '%s\n' "-I" "$root/apalache-stdlib"
	if [[ "$spec" == "$root/tla-plus-bench/specs/"* ]]; then
		while IFS= read -r dir; do
			printf '%s\n' "-I" "$dir"
		done < <(bench_manifest_dirs "$spec")
		printf '%s\n' "-I" "$root/tla-plus-bench/specs/gold"
		printf '%s\n' "-I" "$root/tla-plus-bench/specs/silver"
	fi
  case "$parent_name" in
    gold|silver)
      printf '%s\n' "-I" "$parent_dir"
      ;;
  esac
}

generate_xml_gold() {
  local spec="$1"
  local gold="$spec.xml.gold"
  local red="$spec.xml.red"
  if ! should_generate "$gold" "$red"; then
    return 0
  fi
  ensure_java_sany

  local source_dir spec_base tmp_stdout tmp_stderr tmp_xml
  source_dir="$(dirname "$spec")"
  spec_base="$(basename "$spec")"
  tmp_stdout="$(mktemp)"
  tmp_stderr="$(mktemp)"
  tmp_xml="$(mktemp)"

  local include_args=()
  mapfile -t include_args < <(java_include_args "$spec")

  printf 'XML  %s\n' "$(display_path "$spec")"
  set +e
  (cd "$source_dir" && java -cp "$java_classpath" tla2sany.xml.XMLExporter -o "${include_args[@]}" "$spec_base") >"$tmp_stdout" 2>"$tmp_stderr"
  local status=$?
  set -e

  awk 'found || /<\?xml/ { found = 1; print }' "$tmp_stdout" >"$tmp_xml"
  if grep -q '<\?xml' "$tmp_xml" && xml_is_well_formed "$tmp_xml"; then
    mv "$tmp_xml" "$gold"
    chmod 0644 "$gold"
    rm -f "$red"
    rm -f "$tmp_stdout" "$tmp_stderr"
    return 0
  fi

  if [[ "$status" -ne 0 ]]; then
    write_red "$red" "XML" "$spec" "$status" "$tmp_stdout" "$tmp_stderr" "java -cp <frozen-classpath> tla2sany.xml.XMLExporter -o ${include_args[*]} $spec_base"
    rm -f "$gold"
    rm -f "$tmp_stdout" "$tmp_stderr" "$tmp_xml"
    printf 'RED  %s\n' "$(display_path "$red")"
    return 0
  fi

  write_red "$red" "XML" "$spec" "malformed-xml" "$tmp_stdout" "$tmp_stderr" "java -cp <frozen-classpath> tla2sany.xml.XMLExporter -o ${include_args[*]} $spec_base"
  rm -f "$gold"
  rm -f "$tmp_stdout" "$tmp_stderr" "$tmp_xml"
  printf 'RED  %s\n' "$(display_path "$red")"
}

ensure_apalache() {
  if [[ -n "$apalache_command" ]]; then
    return
  fi
  if [[ -n "${APALACHE_MC:-}" ]]; then
    apalache_command="$APALACHE_MC"
    return
  fi
  apalache_command="$(command -v apalache-mc || true)"
  if [[ -z "$apalache_command" ]]; then
    local candidate
    for candidate in "${local_apalache_candidates[@]}"; do
      if [[ -x "$candidate" ]]; then
        apalache_command="$candidate"
        if [[ -z "${APALACHE_JAR:-}" ]]; then
          local jar_candidate
          for jar_candidate in "${local_apalache_jar_candidates[@]}"; do
            if [[ -f "$jar_candidate" ]]; then
              apalache_jar="$jar_candidate"
              break
            fi
          done
        fi
        break
      fi
    done
  fi
  if [[ -z "$apalache_command" ]]; then
    die "APALACHE_MC or apalache-mc on PATH is required to generate .air.gold files"
  fi
}

path_list_with_existing() {
  local name="$1"
  local value="$2"
  local existing="${!name-}"
  if [[ -n "$existing" ]]; then
    printf '%s:%s' "$value" "$existing"
  else
    printf '%s' "$value"
  fi
}

apalache_library_path() {
	local spec="$1"
	local source_dir parent_dir parent_name
	source_dir="$(dirname "$spec")"
	parent_dir="$(dirname "$source_dir")"
	parent_name="$(basename "$parent_dir")"

	local dirs=("$source_dir" "$root/java-sany/StandardModules" "$root/tlaplus-standard-modules" "$root/CommunityModules/modules" "$root/CommunityModules/tests" "$root/tlaps-stdlib" "$root/apalache-stdlib")
	if [[ "$spec" == "$root/tla-plus-bench/specs/"* ]]; then
		local manifest_dirs=()
		mapfile -t manifest_dirs < <(bench_manifest_dirs "$spec")
		dirs+=("${manifest_dirs[@]}")
		dirs+=("$root/tla-plus-bench/specs/gold" "$root/tla-plus-bench/specs/silver")
	fi
	case "$parent_name" in
		gold|silver)
			dirs+=("$parent_dir")
			;;
  esac

  local IFS=:
  printf '%s' "${dirs[*]}"
}

generate_air_gold() {
	local spec="$1"
	local gold="$spec.air.gold"
	local red="$spec.air.red"
  if ! should_generate "$gold" "$red"; then
    return 0
  fi
  ensure_apalache

	local source_dir spec_base tmp_dir out_json lib_path tmp_stdout tmp_stderr
	source_dir="$(dirname "$spec")"
	spec_base="$(basename "$spec")"
	tmp_dir="$(mktemp -d)"
	out_json="$tmp_dir/out.json"
	lib_path="$(apalache_library_path "$spec")"
	tmp_stdout="$tmp_dir/stdout"
	tmp_stderr="$tmp_dir/stderr"

	printf 'AIR  %s\n' "$(display_path "$spec")"
	set +e
	(
		cd "$source_dir"
		if [[ -n "$apalache_jar" ]]; then
			export APALACHE_JAR="$apalache_jar"
		fi
    TLA_LIBRARY_PATH="$(path_list_with_existing TLA_LIBRARY_PATH "$lib_path")" \
    TLA_PATH="$(path_list_with_existing TLA_PATH "$lib_path")" \
    "$apalache_command" --out-dir="$tmp_dir/apalache-out" parse --output="$out_json" "$spec_base"
  ) >"$tmp_stdout" 2>"$tmp_stderr"
  local status=$?
  set -e
  if [[ "$status" -ne 0 || ! -f "$out_json" ]]; then
    write_red "$red" "AIR" "$spec" "$status" "$tmp_stdout" "$tmp_stderr" "$apalache_command --out-dir=<tmp>/apalache-out parse --output=<tmp>/out.json $spec_base"
    rm -f "$gold"
    rm -rf "$tmp_dir"
    printf 'RED  %s\n' "$(display_path "$red")"
    return 0
  fi

  mv "$out_json" "$gold"
  chmod 0644 "$gold"
  rm -f "$red"
  rm -rf "$tmp_dir"
}

generate_xml() {
  local spec
  while IFS= read -r spec; do
    [[ -n "$spec" ]] || continue
    generate_xml_gold "$spec"
  done < <(collect_specs "${xml_roots[@]}")
}

generate_air() {
  local spec
  while IFS= read -r spec; do
    [[ -n "$spec" ]] || continue
    generate_air_gold "$spec"
  done < <(collect_specs "${air_roots[@]}")
}

case "$mode" in
  all)
    generate_xml
    generate_air
    ;;
  xml)
    generate_xml
    ;;
  air)
    generate_air
    ;;
  *)
    die "usage: $0 [all|xml|air]"
    ;;
esac
