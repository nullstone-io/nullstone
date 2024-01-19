# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "An internal developer platform running on your cloud"
  homepage "https://nullstone.io"
  version "0.0.110"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.110/nullstone_0.0.110_darwin_arm64.tar.gz"
      sha256 "6ecbabc2b35e7abd6b31ac2a874c2ceca7d69c9146976b05fd5bc32eea552d3d"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.110/nullstone_0.0.110_darwin_amd64.tar.gz"
      sha256 "daecf65143f67686ba11af8e066a4fe1e410e723f495adf118163f9578ebcee6"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.110/nullstone_0.0.110_linux_arm64.tar.gz"
      sha256 "cc11c70dba7437c3fd3fae733d8a71f2756a6feca61caf724efb84ba833b7445"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.110/nullstone_0.0.110_linux_amd64.tar.gz"
      sha256 "5314b20102c61f898420af2755a913cd116dd5d8401b1d524e6119b9b7675656"

      def install
        bin.install "nullstone"
      end
    end
  end
end
