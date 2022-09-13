# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "Launch apps on your cloud in minutes"
  homepage "https://nullstone.io"
  version "0.0.79"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.79/nullstone_0.0.79_darwin_amd64.tar.gz"
      sha256 "e1be76f29674dd84242953dfe502c3f575a6af121f30bfc387182c3768338f0e"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.79/nullstone_0.0.79_darwin_arm64.tar.gz"
      sha256 "521f750cd96cf7fcd1e1ae9ac739f760f27149698069043e94c4328f80317731"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.79/nullstone_0.0.79_linux_arm64.tar.gz"
      sha256 "9d647e6b481da21505f62b0f326d9f5c90e4e603c7517b4665c215c68e08fc02"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.79/nullstone_0.0.79_linux_amd64.tar.gz"
      sha256 "1ff3646e0a5e684d4b5d45230d65c40a1783441f8b2fc91cd249875269363fb0"

      def install
        bin.install "nullstone"
      end
    end
  end
end
