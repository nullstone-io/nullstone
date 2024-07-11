# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "An internal developer platform running on your cloud"
  homepage "https://nullstone.io"
  version "0.0.121"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.121/nullstone_0.0.121_darwin_amd64.tar.gz"
      sha256 "a113a7257316ad2491604c8b59f9423694fc4eb5da0e31b77fcb36db2926debd"

      def install
        bin.install "nullstone"
      end
    end
    on_arm do
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.121/nullstone_0.0.121_darwin_arm64.tar.gz"
      sha256 "88a89ca9f8079f984cbf842aa876f6bbc9929bb352efcc9e3b80dfe9a6354375"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.121/nullstone_0.0.121_linux_amd64.tar.gz"
        sha256 "ea393d23cbe5f5a4aa990d20bb2391c4d1f47f332bc72a1c534ba5667870d9e7"

        def install
          bin.install "nullstone"
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.121/nullstone_0.0.121_linux_arm64.tar.gz"
        sha256 "8f7ee03dd8cc3436755fab53cbe541bd04238fd966f4f05f5e8076859ae6759e"

        def install
          bin.install "nullstone"
        end
      end
    end
  end
end
