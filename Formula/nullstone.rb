# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "An internal developer platform running on your cloud"
  homepage "https://nullstone.io"
  version "0.0.132"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.132/nullstone_0.0.132_darwin_amd64.tar.gz"
      sha256 "73721dc4afa9ee3f27d5210d5cd3875cd13161ad45cd0746d001b3de7b48e9dc"

      def install
        bin.install "nullstone"
      end
    end
    on_arm do
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.132/nullstone_0.0.132_darwin_arm64.tar.gz"
      sha256 "959e010a7459f8d5d2f1b481748256f074f15181389ca877eb4b68d41deb7824"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.132/nullstone_0.0.132_linux_amd64.tar.gz"
        sha256 "2ebd2be34265670945186151a69e374ff3a72471f9ab3bc9c69cc2ac356914d6"

        def install
          bin.install "nullstone"
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.132/nullstone_0.0.132_linux_arm64.tar.gz"
        sha256 "3d45092b8c8f7703e7d70a16fb2676d61ae6776e8a17fbd71a0054c7013d4f1e"

        def install
          bin.install "nullstone"
        end
      end
    end
  end
end
