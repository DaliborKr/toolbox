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
@test "environment variables: HISTFILESIZE inside Fedora 42-aarch64" {
  # shellcheck disable=SC2031
  if [ "$HISTFILESIZE" = "" ]; then
    # shellcheck disable=SC2030
    HISTFILESIZE=1001
  else
    ((HISTFILESIZE++))
  fi

  export HISTFILESIZE

  # shellcheck disable=SC2016
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 bash -c 'echo "$HISTFILESIZE"'

  assert_success
  assert_line --index 0 "$HISTFILESIZE"
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
@test "environment variables: HISTSIZE inside Fedora 42-aarch64" {
  skip "https://pagure.io/setup/pull-request/48"

  # shellcheck disable=SC2031
  if [ "$HISTSIZE" = "" ]; then
    # shellcheck disable=SC2030
    HISTSIZE=1001
  else
    ((HISTSIZE++))
  fi

  export HISTSIZE

  # shellcheck disable=SC2016
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 bash -c 'echo "$HISTSIZE"'

  assert_success
  assert_line --index 0 "$HISTSIZE"
  assert [ ${#lines[@]} -eq 1 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}

# bats test_tags=arch-fedora
@test "environment variables: HOSTNAME inside Fedora 42-aarch64" {
  # shellcheck disable=SC2016
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 bash -c 'echo "$HOSTNAME"'

  assert_success
  assert_line --index 0 "$HOSTNAME"
  assert [ ${#lines[@]} -eq 1 ]
  assert [ ${#stderr_lines[@]} -eq 0 ]
}

# bats test_tags=arch-fedora
@test "environment variables: KONSOLE_VERSION inside Fedora 42-aarch64" {
  # shellcheck disable=SC2031
  if [ "$KONSOLE_VERSION" = "" ]; then
    # shellcheck disable=SC2030
    export KONSOLE_VERSION=230804
  fi

  # shellcheck disable=SC2016
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 bash -c 'echo "$KONSOLE_VERSION"'

  assert_success
  assert_line --index 0 "$KONSOLE_VERSION"
  assert [ ${#lines[@]} -eq 1 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}

# bats test_tags=arch-fedora
@test "environment variables: XTERM_VERSION inside Fedora 42-aarch64" {
  # shellcheck disable=SC2031
  if [ "$XTERM_VERSION" = "" ]; then
    # shellcheck disable=SC2030
    export XTERM_VERSION="XTerm(385)"
  fi

  # shellcheck disable=SC2016
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 bash -c 'echo "$XTERM_VERSION"'

  assert_success
  assert_line --index 0 "$XTERM_VERSION"
  assert [ ${#lines[@]} -eq 1 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}
