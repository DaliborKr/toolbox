load 'libs/bats-support/load'
load 'libs/bats-assert/load'
load 'libs/helpers'

readonly RESOLVER_PYTHON3='\
import socket; \
import sys; \
family = {"A": socket.AddressFamily.AF_INET, "AAAA": socket.AddressFamily.AF_INET6}; \
addr = socket.getaddrinfo(sys.argv[2], None, family[sys.argv[1]], socket.SocketKind.SOCK_RAW)[0][4][0]; \
print(addr)'

# shellcheck disable=SC2016
readonly RESOLVER_SH='resolvectl --legend false --no-pager --type "$0" query "$1" \
                      | cut --delimiter " " --fields 4'

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
@test "network: No namespace inside Fedora 42-aarch64" {
  local ns_host
  ns_host=$(readlink /proc/$$/ns/net)

  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 sh -c 'readlink /proc/$$/ns/net'

  assert_success
  assert_line --index 0 "$ns_host"
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
@test "network: /etc/resolv.conf inside Fedora 42-aarch64" {
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 readlink /etc/resolv.conf

  assert_success

  if [ "${lines[0]}" = "/run/host/run/systemd/resolve/stub-resolv.conf" ]; then
    skip "host has absolute symlink"
  else
    assert_line --index 0 "/run/host/etc/resolv.conf"
  fi

  assert [ ${#lines[@]} -eq 1 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}

# bats test_tags=arch-fedora
@test "network: DNS inside Fedora 42-aarch64" {
  local ipv4_skip=false
  local ipv4_addr
  if ! ipv4_addr="$(python3 -c "$RESOLVER_PYTHON3" A k.root-servers.net)"; then
    ipv4_skip=true
  fi

  local ipv6_skip=false
  local ipv6_addr
  if ! ipv6_addr="$(python3 -c "$RESOLVER_PYTHON3" AAAA k.root-servers.net)"; then
    ipv6_skip=true
  fi

  if $ipv4_skip && $ipv6_skip; then
    skip "DNS not working on host"
  fi

  if ! $ipv4_skip; then
    run --keep-empty-lines --separate-stderr "$TOOLBX" run \
      --distro fedora \
      --release 42-aarch64 \
      python3 -c "$RESOLVER_PYTHON3" A k.root-servers.net

    assert_success
    assert_line --index 0 "$ipv4_addr"
    assert [ ${#lines[@]} -eq 1 ]
    assert [ ${#stderr_lines[@]} -eq 0 ]
  fi

  if ! $ipv6_skip; then
    run --keep-empty-lines --separate-stderr "$TOOLBX" run \
      --distro fedora \
      --release 42-aarch64 \
      python3 -c "$RESOLVER_PYTHON3" AAAA k.root-servers.net

    assert_success
    assert_line --index 0 "$ipv6_addr"
    assert [ ${#lines[@]} -eq 1 ]
    assert [ ${#stderr_lines[@]} -eq 0 ]
  fi
}

# bats test_tags=arch-fedora
@test "network: ping(8) inside Fedora 42-aarch64" {
  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 ping -c 2 f.root-servers.net

  if [ "$status" -eq 1 ]; then
    skip "lost packets"
  fi

  assert_success
  assert [ ${#lines[@]} -gt 0 ]

  # shellcheck disable=SC2154
  assert [ ${#stderr_lines[@]} -eq 0 ]
}
