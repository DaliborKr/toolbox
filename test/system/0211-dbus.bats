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
@test "dbus: Session bus inside Fedora 42-aarch64" {
  local expected_response
  expected_response="$(gdbus call \
                         --session \
                         --dest org.freedesktop.DBus \
                         --object-path /org/freedesktop/DBus \
                         --method org.freedesktop.DBus.Peer.Ping)"

  run --keep-empty-lines --separate-stderr "$TOOLBX" run \
    --distro fedora \
    --release 42-aarch64 \
    gdbus call \
      --session \
      --dest org.freedesktop.DBus \
      --object-path /org/freedesktop/DBus \
      --method org.freedesktop.DBus.Peer.Ping

  assert_success
  assert_line --index 0 "$expected_response"
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
@test "dbus: System bus inside Fedora 42-aarch64" {
  local expected_response
  expected_response="$(gdbus call \
                         --system \
                         --dest org.freedesktop.systemd1 \
                         --object-path /org/freedesktop/systemd1 \
                         --method org.freedesktop.DBus.Properties.Get \
                         org.freedesktop.systemd1.Manager \
                         Version)"

  run --keep-empty-lines --separate-stderr "$TOOLBX" run \
    --distro fedora \
    --release 42-aarch64 \
    gdbus call \
      --system \
      --dest org.freedesktop.systemd1 \
      --object-path /org/freedesktop/systemd1 \
      --method org.freedesktop.DBus.Properties.Get \
      org.freedesktop.systemd1.Manager \
      Version

  assert_success
  assert_line --index 0 "$expected_response"
  assert [ ${#lines[@]} -eq 1 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}