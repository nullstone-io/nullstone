# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.47"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.47/nullstone_0.0.47_darwin_amd64.tar.gz"
      sha256 "86290abdfa17e24c0287f96a76da26c79198b5d07071a507789b2dae05354dae"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.47/nullstone_0.0.47_darwin_arm64.tar.gz"
      sha256 "77abe1e9a8dbcc0a0c58d8c48c468f713d40ca2324a891e5967b68a983028971"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.47/nullstone_0.0.47_linux_amd64.tar.gz"
      sha256 "ba754dbbc726c3ecf1c7191fbc6b8d91f55d1d684a76079cf5a819fa488c8a49"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.47/nullstone_0.0.47_linux_arm64.tar.gz"
      sha256 "3699c80fafc43feef5f5db56be6fe9c6fec0b6b504fedd61811d02d2f0af06ef"

      def install
        bin.install "nullstone"
      end
    end
  end
end
