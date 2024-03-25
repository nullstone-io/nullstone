# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "An internal developer platform running on your cloud"
  homepage "https://nullstone.io"
  version "0.0.119"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.119/nullstone_0.0.119_darwin_arm64.tar.gz"
      sha256 "18646a846f5c64c7fa4a629fb316b7c33bcbd31ac78292c21cad29b173950014"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.119/nullstone_0.0.119_darwin_amd64.tar.gz"
      sha256 "b3f412d31843efe7b8843682e4e1401cbfa568699fea7b2818620aef82c76147"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.119/nullstone_0.0.119_linux_arm64.tar.gz"
      sha256 "4738ebc4e025c53a7a510d19615ce925f13160de8ca8e1f47b811543f474defe"

      def install
        bin.install "nullstone"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.119/nullstone_0.0.119_linux_amd64.tar.gz"
      sha256 "f92210f0098e211b4e5715a4432d8fb3c83812131b69ba535733273c15ff6704"

      def install
        bin.install "nullstone"
      end
    end
  end
end
