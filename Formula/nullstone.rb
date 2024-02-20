# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "An internal developer platform running on your cloud"
  homepage "https://nullstone.io"
  version "0.0.116"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.116/nullstone_0.0.116_darwin_arm64.tar.gz"
      sha256 "077f79d6b4c706d8ae728b5b9b40a0282550dfb9d35b764882278e7927a96a31"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.116/nullstone_0.0.116_darwin_amd64.tar.gz"
      sha256 "80dad62c63d3dd8dfa2ff774402465e922423fb27ac7e5b730800db07f0b10f0"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.116/nullstone_0.0.116_linux_arm64.tar.gz"
      sha256 "1d57922f22e8cb9d67bbe52ce66816cf27a8987a9e4818cf7ef7aa177973d818"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.116/nullstone_0.0.116_linux_amd64.tar.gz"
      sha256 "e0c68b7176b9f06388c5b17bf3a9a6c793e975292eb0488f61cd62510c00f300"

      def install
        bin.install "nullstone"
      end
    end
  end
end
