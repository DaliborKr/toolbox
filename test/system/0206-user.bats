load 'libs/bats-support/load'
load 'libs/bats-assert/load'
load 'libs/helpers'

setup_file() {
  bats_require_minimum_version 1.10.0
  cleanup_all
  pushd "$HOME" || return 1

  if echo "$TOOLBX_TEST_SYSTEM_TAGS" | grep "fedora" >/dev/null 2>/dev/null; then
    create_distro_container fedora 42-aarch64 fedora-toolbox-42-aarch64
  fi
}

teardown_file() {
  popd || return 1
  cleanup_all
}


# bats test_tags=arch-fedora
@test "user: Separate namespace inside Fedora 42-aarch64" {
  local ns_host
  ns_host=$(readlink /proc/$$/ns/user)

  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 sh -c 'readlink /proc/$$/ns/user'

  assert_success
  assert_line --index 0 --regexp '^user:\[[[:digit:]]+\]$'
  refute_line --index 0 "$ns_host"
  assert [ ${#lines[@]} -eq 1 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]

  # -- Architecture 'aarch64' check
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 sh -c 'lscpu'

  assert_success
  assert_line --index 0 --regexp "^Architecture:[[:space:]]+aarch64$"
  assert [ ${#stderr_lines[@]} -eq 0 ]
}

# bats test_tags=arch-fedora
@test "user: root in shadow(5) inside Fedora 42-aarch64" {
  container_root_file_system="$(podman unshare podman mount fedora-toolbox-42-aarch64)"

  "$TOOLBX" run --distro fedora --release 42-aarch64 true

  run --keep-empty-lines --separate-stderr podman unshare cat "$container_root_file_system/etc/shadow"
  podman unshare podman unmount fedora-toolbox-42-aarch64

  assert_success
  assert_line --regexp '^root::.+$'
  assert [ ${#lines[@]} -gt 0 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}

# bats test_tags=arch-fedora
@test "user: $USER in passwd(5) inside Fedora 42-aarch64" {
  local user_gecos
  user_gecos="$(getent passwd "$USER" | cut --delimiter : --fields 5)"

  local user_id_real
  user_id_real="$(id --real --user)"

  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 cat /etc/passwd

  assert_success
  assert_line --regexp "^$USER::$user_id_real:$user_id_real:$user_gecos:$HOME:$SHELL$"
  assert [ ${#lines[@]} -gt 1 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}

# bats test_tags=arch-fedora
@test "user: $USER in shadow(5) inside Fedora 42-aarch64" {
  container_root_file_system="$(podman unshare podman mount fedora-toolbox-42-aarch64)"

  "$TOOLBX" run --distro fedora --release 42-aarch64 true

  run --keep-empty-lines --separate-stderr podman unshare cat "$container_root_file_system/etc/shadow"
  podman unshare podman unmount fedora-toolbox-42-aarch64

  assert_success
  refute_line --regexp "^$USER:.*$"
  assert [ ${#lines[@]} -gt 0 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}

# bats test_tags=arch-fedora
@test "user: $USER in group(5) inside Fedora 42-aarch64" {
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 cat /etc/group

  assert_success
  assert_line --regexp "^$USER:x:[[:digit:]]+:$USER$"
  assert_line --regexp "^wheel:x:[[:digit:]]+:$USER$"
  assert [ ${#lines[@]} -gt 1 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}

# bats test_tags=arch-fedora
@test "user: id(1) for $USER inside Fedora 42-aarch64" {
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 id

  assert_success
  assert [ ${#lines[@]} -eq 1 ]

  local output_id="${lines[0]}"

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]

  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 id "$USER"

  assert_success
  assert_line --index 0 "$output_id"
  assert [ ${#lines[@]} -eq 1 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}