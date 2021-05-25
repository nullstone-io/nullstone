# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.20"
  license "MIT"
  bottle :unneeded

  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.20/nullstone_0.0.20_Darwin_x86_64.tar.gz"
    sha256 "6dc458e94f1d5ab7984eca2e0ab7a0cb51e6e5a95ba3b160e7f38b2f06074b31"
  end
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.20/nullstone_0.0.20_Darwin_arm64.tar.gz"
    sha256 "09b8f394dd2141ea58e644fb05550ae1b07b0065259e31ed6ed0bea001470bb9"
  end
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.20/nullstone_0.0.20_Linux_x86_64.tar.gz"
    sha256 "2c42851762836a8988b9eafb8ce643f6a9bde0949a29ea782ebfc9e02eb3cab3"
  end
  if OS.linux? && Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
    url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.20/nullstone_0.0.20_Linux_arm64.tar.gz"
    sha256 "1cc84671ad7f9f44a58eef83de70c02584ef5b2a92e929998c54c47b4fd2e036"
  end

  depends_on "go"

  def install
    bin.install "nullstone"
  end
end
