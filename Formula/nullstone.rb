# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.59"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.59/nullstone_0.0.59_darwin_arm64.tar.gz"
      sha256 "878febdab0bba82210708cce6825929da057709d2106fa66d4e8795c6668b662"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.59/nullstone_0.0.59_darwin_amd64.tar.gz"
      sha256 "b509bc1fe6bda49214f944c89d64c2544f0b47cb03c0d4e6bd9c9955ee682adb"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.59/nullstone_0.0.59_linux_amd64.tar.gz"
      sha256 "5c7eee752508adaf66e38e9e81b314cda9aeddb8fd6d4c29cf4d806406159951"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.59/nullstone_0.0.59_linux_arm64.tar.gz"
      sha256 "1d10e2054757ed086fcdd71935b3e609c4c8bb06b1e338488f675d7a12bd7a67"

      def install
        bin.install "nullstone"
      end
    end
  end
end
