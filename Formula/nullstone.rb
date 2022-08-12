# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.76"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.76/nullstone_0.0.76_darwin_arm64.tar.gz"
      sha256 "3065e452e18375652ffd2f3184575dd49d151c6f7cab501b005b054bbd7da864"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.76/nullstone_0.0.76_darwin_amd64.tar.gz"
      sha256 "03a58d7c59cb1ca7fae0ce321e00d07d0191848d9d959d81956e51fa5ac0202d"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.76/nullstone_0.0.76_linux_arm64.tar.gz"
      sha256 "5657247a856ff09e1f93f967014a2b2ce1c42f66ff01bbfdead93f9575999393"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.76/nullstone_0.0.76_linux_amd64.tar.gz"
      sha256 "9c140b3b168dd563ac5e2a79770f430b516dca7b134ee888fd08f1666f9f259f"

      def install
        bin.install "nullstone"
      end
    end
  end
end
