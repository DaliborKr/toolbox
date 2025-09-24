load 'libs/bats-support/load'
load 'libs/bats-assert/load'
load 'libs/helpers'

setup() {
  bats_require_minimum_version 1.10.0
  cleanup_all
  pushd "$HOME" || return 1
}

teardown() {
  popd || return 1
  cleanup_all
}

# bats test_tags=arch-fedora
@test "rpm: %_netsharedpath inside Fedora 42-aarch64" {
  create_distro_container fedora 42-aarch64 fedora-toolbox-42-aarch64

  run --keep-empty-lines --separate-stderr "$TOOLBX" run --distro fedora --release 42-aarch64 rpm --eval %_netsharedpath

  assert_success
  assert_line --index 0 "/dev:/media:/mnt:/proc:/sys:/tmp:/var/lib/flatpak:/var/lib/libvirt"
  assert [ ${#lines[@]} -eq 1 ]
  assert [ ${#stderr_lines[@]} -eq 0 ]

  run --keep-empty-lines --separate-stderr "$TOOLBX" run \
    --distro fedora \
    --release 42-aarch64 \
    cat /usr/lib/rpm/macros.d/macros.toolbox

  assert_success
  assert_line --index 0 "# Written by Toolbx"
  assert_line --index 1 "# https://containertoolbx.org/"
  assert_line --index 2 ""
  assert_line --index 3 "%_netsharedpath /dev:/media:/mnt:/proc:/sys:/tmp:/var/lib/flatpak:/var/lib/libvirt"
  assert [ ${#lines[@]} -eq 4 ]
  assert [ ${#stderr_lines[@]} -eq 0 ]

  run --keep-empty-lines --separate-stderr "$TOOLBX" run \
    --distro fedora \
    --release 42-aarch64 \
    stat \
      --format "%A %U:%G" \
      /usr/lib/rpm/macros.d/macros.toolbox

  assert_success
  assert_line --index 0 "-rw-r--r-- root:root"
  assert [ ${#lines[@]} -eq 1 ]
  assert [ ${#stderr_lines[@]} -eq 0 ]
}
