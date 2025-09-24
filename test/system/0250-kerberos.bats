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
@test "kerberos: Smoke test with Fedora 42-aarch64" {
  create_distro_container fedora 42-aarch64 fedora-toolbox-42-aarch64

  run --keep-empty-lines --separate-stderr "$TOOLBX" run \
    --distro fedora \
    --release 42-aarch64 \
    cat /etc/krb5.conf.d/kcm_default_ccache

  assert_success
  assert_line --index 0 "# Written by Toolbx"
  assert_line --index 1 "# https://containertoolbx.org/"
  assert_line --index 2 "#"
  assert_line --index 3 "# # To disable the KCM credential cache, comment out the following lines."
  assert_line --index 4 ""
  assert_line --index 5 "[libdefaults]"
  assert_line --index 6 "    default_ccache_name = KCM:"
  assert [ ${#lines[@]} -eq 7 ]
  assert [ ${#stderr_lines[@]} -eq 0 ]

  run --keep-empty-lines --separate-stderr "$TOOLBX" run \
    --distro fedora \
    --release 42-aarch64 \
    stat \
      --format "%A %U:%G" \
      /etc/krb5.conf.d/kcm_default_ccache

  assert_success
  assert_line --index 0 "-rw-r--r-- root:root"
  assert [ ${#lines[@]} -eq 1 ]
  assert [ ${#stderr_lines[@]} -eq 0 ]
}