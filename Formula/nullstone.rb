# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.84"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.84/nullstone_0.0.84_darwin_amd64.tar.gz"
      sha256 "4a574f58a0109c052b32371fc482b3f810599df440d39c75b2a73330719734aa"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.84/nullstone_0.0.84_darwin_arm64.tar.gz"
      sha256 "73e48b362b82888018629857b26a1c2018df1d7ee012c1378c7dfaa6b79203d9"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.84/nullstone_0.0.84_linux_arm64.tar.gz"
      sha256 "69980a1a8c79802ec8d233d8641fe6682fc22adae5637188ab008c5627a8625a"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.84/nullstone_0.0.84_linux_amd64.tar.gz"
      sha256 "18ef6536f360773c4ec24da54659f3f3e66eb568d9ca506b104e3e6984260b24"

      def install
        bin.install "nullstone"
      end
    end
  end
end
