# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.52"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.52/nullstone_0.0.52_darwin_arm64.tar.gz"
      sha256 "807252707275af68c0534727a28bdf84effea29b9629dd687db8c22f0135e434"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.52/nullstone_0.0.52_darwin_amd64.tar.gz"
      sha256 "71e7627ffebe919f7547acd0ff9e7bbba5200a9f85d8ab9c9512087cca94ae57"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.52/nullstone_0.0.52_linux_arm64.tar.gz"
      sha256 "be45cb96aebc60fe95f29f0418e4dc7eb914095c8b074943cb84e5dfa3c0fb08"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.52/nullstone_0.0.52_linux_amd64.tar.gz"
      sha256 "e368d516bb5ea81118deda34f8591df75c67d1b2c4775ccf44813f0399707450"

      def install
        bin.install "nullstone"
      end
    end
  end
end
