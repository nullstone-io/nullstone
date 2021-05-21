# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.18"
  license "MIT"
  bottle :unneeded

  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.18/nullstone_0.0.18_Darwin_x86_64.tar.gz"
    sha256 "cfd28af698fcb672fbf479e8f4444b4127d4487cd5ca1242a3960c7ab80edfae"
  end
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.18/nullstone_0.0.18_Darwin_arm64.tar.gz"
    sha256 "74eac81295daeeaca2c06e5e95c0a9684dcfc07ab78254e87a490e362dee89f6"
  end
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.18/nullstone_0.0.18_Linux_x86_64.tar.gz"
    sha256 "16edbc7a61c33355022b34c93efe3d43996eb9287d52e27e3091661894c7c89d"
  end
  if OS.linux? && Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
    url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.18/nullstone_0.0.18_Linux_arm64.tar.gz"
    sha256 "7c55ebc07d4859924003c62a96171d03601b40cefca71b14729bcfe992467124"
  end

  depends_on "go"

  def install
    bin.install "nullstone"
  end
end
