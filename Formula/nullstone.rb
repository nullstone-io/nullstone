# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Nullstone < Formula
  desc "An internal developer platform running on your cloud"
  homepage "https://nullstone.io"
  version "0.0.122"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.122/nullstone_0.0.122_darwin_amd64.tar.gz"
      sha256 "9b49418f84f83d023d82b7533606dd7ddbea68b934ed6839b1fe27bd3f1ccfa6"

      def install
        bin.install "nullstone"
      end
    end
    on_arm do
      url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.122/nullstone_0.0.122_darwin_arm64.tar.gz"
      sha256 "89f384ee43313a86c3538e098cbc5634e87574fffe487ee7b8d651644fbb0a0a"

      def install
        bin.install "nullstone"
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.122/nullstone_0.0.122_linux_amd64.tar.gz"
        sha256 "27e7b9853f31837abe3d298c3dee3305f92f47afb23aae07755bd5c3d9768a5a"

        def install
          bin.install "nullstone"
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/nullstone-io/nullstone/releases/download/v0.0.122/nullstone_0.0.122_linux_arm64.tar.gz"
        sha256 "361c7707fa79f4b9928a8bb679663d6afeeaa8c2de592315e461ec0432596f87"

        def install
          bin.install "nullstone"
        end
      end
    end
  end
end
